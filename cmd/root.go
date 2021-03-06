package cmd

import (
	"errors"
	"fmt"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/remote"
	"github.com/spf13/cobra"
	"os"
)

type RootCmd struct {
	*cobra.Command
	Verbose          bool
	ConfigFiles      []string
	IsRemote         bool
	ForceLocal       bool
	OverrideLocalDir string
	Context          *context.Context
	cleanedUp        bool
}

func NewRootCmd(cmds ...Command) *RootCmd {
	c := &RootCmd{}
	c.Command = &cobra.Command{
		Use:   "admin",
		Short: "Admin tool",
		PersistentPreRun: func(childCmd *cobra.Command, args []string) {
			if err := c.preRun(childCmd, args); err != nil {
				fmt.Printf("Error: %v\n", err)
				c.cleanUp()
				os.Exit(-1)
			}
		},
		PersistentPostRun: func(childCmd *cobra.Command, args []string) {
			c.cleanUp()
		},
	}
	c.PersistentFlags().BoolVarP(&c.Verbose, "verbose", "v", false, "Verbose output")
	c.PersistentFlags().StringSliceVarP(&c.ConfigFiles, "config", "c", nil, "Config file(s)")
	c.PersistentFlags().BoolVar(&c.IsRemote, "is-remote", false, "If remote we ignore several things")
	c.PersistentFlags().BoolVar(&c.ForceLocal, "force-local", false, "Never run remote regardless of config")
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
		if r.OverrideLocalDir == "" {
			return errors.New("Must have --override-local-dir for remote")
		}
		ctx, err := context.FromRemoteStdPipe(r.Verbose, r.OverrideLocalDir)
		if err != nil {
			return fmt.Errorf("Unable to begin remote command over std pipes: %v", err)
		}
		r.Context = ctx
	}
	return nil
}

func (r *RootCmd) cleanUp() {
	if !r.cleanedUp {
		// If we're remote we need to delete ourself
		if r.IsRemote {
			r.Context.Debugf("Removing self at %v", os.Args[0])
			if err := os.Remove(os.Args[0]); err != nil {
				r.Context.Infof("Failed to remove %v: %v", os.Args[0], err)
			}
		}
		// We need to remove the entire temp directory every time if context was created
		if r.Context != nil {
			r.Context.Debugf("Removing temp directory at %v", r.Context.TempDir)
			if err := os.RemoveAll(r.Context.TempDir); err != nil {
				r.Context.Debugf("Unable to remove temp dir %v: %v", r.Context.TempDir, err)
			}
		}
		r.cleanedUp = true
	}
}

type Command interface {
	CmdInfo() *cobra.Command
}

type ParentCommand interface {
	Command
	Children() []Command
}

type RunnableCommand interface {
	Command
	Run(*context.Context) error
}

// This overrides Run if it's a RunnableCommand
func (r *RootCmd) AddCommand(cmd Command) {
	r.addCommand(r.Command, cmd)
}

func (r *RootCmd) addCommand(parent *cobra.Command, cmds ...Command) {
	for _, cmd := range cmds {
		info := cmd.CmdInfo()
		switch cmd := cmd.(type) {
		case ParentCommand:
			r.addCommand(info, cmd.Children()...)
		case RunnableCommand:
			info.Run = func(childCmd *cobra.Command, args []string) {
				if err := cmd.Run(r.Context); err != nil {
					r.Context.Infof("Error: %v", err)
					r.cleanUp()
					os.Exit(-1)
				}
			}
		}
		parent.AddCommand(info)
	}
}
