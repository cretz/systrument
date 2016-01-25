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
	Context *context.Context
}

func NewRootCmd(cmds ...Command) *RootCmd {
	c := &RootCmd{}
	c.Command = &cobra.Command{
		Use: "admin",
		Short: "Admin tool",
		PersistentPreRunE: func(childCmd *cobra.Command, args []string) error {
			if !c.IsRemote {
				// Here we send to the remote server if one is provided and we're not ignoring it
				ctx, err := context.FromConfigFiles(c.ConfigFiles, c.Verbose)
				if err != nil {
					return fmt.Errorf("Unable to load from config files: %v", err)
				}
				if !c.ForceLocal {
					if remote, err := remote.RemoteIfPresent(ctx); err != nil {
						return err
					} else if remote != nil {
						if err := remote.RunRemotely(); err != nil {
							return fmt.Errorf("Remote error: %v", err)
						}
						os.Exit(0)
					}
				}
				c.Context = ctx
			} else {
				// TODO: encrypt
				ctx, err := context.FromRemoteStdPipe()
				if err != nil {
					return fmt.Errorf("Unable to begin remote command over std pipes: %v", err)
				}
				c.Context = ctx
			}
			return nil
		},
	}
	c.PersistentFlags().BoolVarP(&c.Verbose, "verbose", "v", false, "Verbose output")
	c.PersistentFlags().StringSliceVarP(&c.ConfigFiles, "config", "c", nil, "Config file(s)")
	c.PersistentFlags().BoolVar(&c.IsRemote, "is-remote", false, "If remote we ignore several things")
	c.PersistentFlags().BoolVar(&c.IsRemote, "force-local", false, "Never run remote regardless of config")

	c.AddCommand(new(ShowConfigCmd))
	for _, childCmd := range cmds {
		c.AddCommand(childCmd)
	}
	return c
}

type Command interface {
	CmdInfo() *cobra.Command
	Run(*context.Context) error
}

// This overrides RunE
func (r *RootCmd) AddCommand(cmd Command) {
	info := cmd.CmdInfo()
	info.RunE = func(childCmd *cobra.Command, args []string) error {
		return cmd.Run(r.Context)
	}
	r.Command.AddCommand(info)
}