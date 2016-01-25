package cmd

import (
	"github.com/spf13/cobra"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/remote"
	"fmt"
	"os"
)

type RootCmd struct {
	*cobra.Command
	Verbose bool
	ConfigFiles []string
	IsRemote bool
	ForceLocal bool
	OverrideLocalDir string
	Context *context.Context
}

func NewRootCmd(cmds ...Command) *RootCmd {
	c := &RootCmd{}
	c.Command = &cobra.Command{
		Use: "admin",
		Short: "Admin tool",
		PersistentPreRun: func(childCmd *cobra.Command, args []string) {
			if err := c.preRun(childCmd, args); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(-1)
			}
		},
	}
	c.PersistentFlags().BoolVarP(&c.Verbose, "verbose", "v", false, "Verbose output")
	c.PersistentFlags().StringSliceVarP(&c.ConfigFiles, "config", "c", nil, "Config file(s)")
	c.PersistentFlags().BoolVar(&c.IsRemote, "is-remote", false, "If remote we ignore several things")
	c.PersistentFlags().BoolVar(&c.IsRemote, "force-local", false, "Never run remote regardless of config")
	c.PersistentFlags().StringVar(&c.OverrideLocalDir, "override-local-dir", "", "The path to the main go file")

	c.AddCommand(new(ShowConfigCmd))
	for _, childCmd := range cmds {
		c.AddCommand(childCmd)
	}
	return c
}

func (r *RootCmd) remoteAllowed(childCmd *cobra.Command) bool {
	return !r.ForceLocal && childCmd.Name() != "showconfig"
}

func (r *RootCmd) preRun(childCmd *cobra.Command, args []string) error {
	if !r.IsRemote {
		// Here we send to the remote server if one is provided and we're not ignoring it
		ctx, err := context.FromConfigFiles(r.ConfigFiles, r.Verbose, r.OverrideLocalDir)
		if err != nil {
			return fmt.Errorf("Unable to load from config files: %v", err)
		}
		if r.remoteAllowed(childCmd) {
			if remote, err := remote.RemoteIfPresent(ctx); err != nil {
				return err
			} else if remote != nil {
				if err := remote.RunRemotely(); err != nil {
					return fmt.Errorf("Remote error: %v", err)
				}
				os.Exit(0)
			}
		}
		r.Context = ctx
	} else {
		// TODO: encrypt
		ctx, err := context.FromRemoteStdPipe()
		if err != nil {
			return fmt.Errorf("Unable to begin remote command over std pipes: %v", err)
		}
		r.Context = ctx
	}
	return nil
}

type Command interface {
	CmdInfo() *cobra.Command
	Run(*context.Context) error
}

// This overrides Run
func (r *RootCmd) AddCommand(cmd Command) {
	info := cmd.CmdInfo()
	info.Run = func(childCmd *cobra.Command, args []string) {
		if err := cmd.Run(r.Context); err != nil {
			r.Context.Infof("Error: %v", err)
			os.Exit(-1)
		}
	}
	r.Command.AddCommand(info)
}