package commands

import (
	"log"
	"log/syslog"
	"net"
	"net/http"
	"os"

	"github.com/SkycoinProject/skycoin/src/util/logging"
	"github.com/SkycoinProject/skywire-mainnet/pkg/util/buildinfo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	logrussyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/cobra"

	"github.com/skycoin/dmsg/cmd/dmsg-discovery/internal/api"
	"github.com/skycoin/dmsg/cmd/dmsg-discovery/internal/store"

	"github.com/skycoin/dmsg/metrics"
)

const redisPasswordEnvName = "REDIS_PASSWORD"

var (
	addr        string
	metricsAddr string
	redisURL    string
	logEnabled  bool
	syslogAddr  string
	tag         string
	testMode    bool
)

var rootCmd = &cobra.Command{
	Use:   "dmsg-discovery",
	Short: "Dmsg Discovery Server for skywire",
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := buildinfo.Get().WriteTo(log.Writer()); err != nil {
			log.Printf("Failed to output build info: %v", err)
		}

		redisPassword := os.Getenv(redisPasswordEnvName)

		s, err := store.NewStore("redis", redisURL, redisPassword)
		if err != nil {
			log.Fatal("Failed to initialize redis store: ", err)
		}

		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal("Failed to open listener: ", err)
		}

		apiLogger := logging.MustGetLogger(tag)
		if !logEnabled {
			apiLogger = nil
		}

		if syslogAddr != "" {
			hook, err := logrussyslog.NewSyslogHook("udp", syslogAddr, syslog.LOG_INFO, tag)
			if err != nil {
				log.Fatalf("Unable to connect to syslog daemon on %v", syslogAddr)
			}
			logging.AddHook(hook)
		}

		logger := api.Logger(apiLogger)
		metrics := api.Metrics(metrics.NewPrometheus("msgdiscovery"))
		testingMode := api.UseTestingMode(testMode)

		api := api.New(s, logger, metrics, testingMode)

		go func() {
			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(metricsAddr, nil); err != nil {
				log.Println("Failed to start metrics API:", err)
			}
		}()

		if apiLogger != nil {
			apiLogger.Infof("Listening on %s", addr)
		}
		log.Fatal(http.Serve(l, api))
	},
}

func init() {
	rootCmd.Flags().StringVarP(&addr, "addr", "a", ":9090", "address to bind to")
	rootCmd.Flags().StringVarP(&metricsAddr, "metrics", "m", ":2121", "address to bind metrics API to")
	rootCmd.Flags().StringVar(&redisURL, "redis", "redis://localhost:6379", "connections string for a redis store")
	rootCmd.Flags().BoolVarP(&logEnabled, "log", "l", true, "enable request logging")
	rootCmd.Flags().StringVar(&syslogAddr, "syslog", "", "syslog server address. E.g. localhost:514")
	rootCmd.Flags().StringVar(&tag, "tag", "dmsg-discovery", "logging tag")
	rootCmd.Flags().BoolVarP(&testMode, "test-mode", "t", false, "in testing mode")
}

// Execute executes root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
