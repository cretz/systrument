package cmd
import (
	"github.com/spf13/cobra"
	"github.com/cretz/systrument/context"
	"encoding/json"
	"fmt"
)

type ShowConfigCmd struct {}

func (_ *ShowConfigCmd) CmdInfo() *cobra.Command {
	return &cobra.Command{
		Use: "showconfig",
		Short: "Show configuration after applying all templates",
	}
}

func (_ *ShowConfigCmd) Run(ctx *context.Context) error {
	byts, err := json.MarshalIndent(ctx.Data.Values, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to marshal JSON: %v", err)
	}
	ctx.Logger.Infof("Config:\n%v", string(byts))
	return nil
}
