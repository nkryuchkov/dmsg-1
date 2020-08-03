package commands

import (
	"context"
	"log"

	"github.com/SkycoinProject/skywire-mainnet/pkg/util/buildinfo"
	"github.com/spf13/cobra"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cmdutil"
	"github.com/skycoin/dmsg/dmsgpty"
)

var cli = dmsgpty.DefaultCLI()

func init() {
	rootCmd.PersistentFlags().StringVar(&cli.Net, "clinet", cli.Net,
		"network to use for dialing to dmsgpty-host")

	rootCmd.PersistentFlags().StringVar(&cli.Addr, "cliaddr", cli.Addr,
		"address to use for dialing to dmsgpty-host")
}

var remoteAddr dmsg.Addr
var cmdName = dmsgpty.DefaultCmd
var cmdArgs []string

func init() {
	rootCmd.Flags().Var(&remoteAddr, "addr",
		"remote dmsg address of format 'pk:port'. If unspecified, the pty will start locally")

	rootCmd.Flags().StringVarP(&cmdName, "cmd", "c", cmdName,
		"name of command to run")

	rootCmd.Flags().StringSliceVarP(&cmdArgs, "args", "a", cmdArgs,
		"command arguments")
}

var rootCmd = &cobra.Command{
	Use:   "dmsgpty-cli",
	Short: "Run commands over dmsg",
	PreRun: func(*cobra.Command, []string) {
		if remoteAddr.Port == 0 {
			remoteAddr.Port = dmsgpty.DefaultPort
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := buildinfo.Get().WriteTo(log.Writer()); err != nil {
			log.Printf("Failed to output build info: %v", err)
		}

		ctx, cancel := cmdutil.SignalContext(context.Background(), nil)
		defer cancel()

		if remoteAddr.PK.Null() {
			// Local pty.
			return cli.StartLocalPty(ctx, cmdName, cmdArgs...)
		}
		// Remote pty.
		return cli.StartRemotePty(ctx, remoteAddr.PK, remoteAddr.Port, cmdName, cmdArgs...)
	},
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
