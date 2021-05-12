package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var Settings *EnvSettings

//TODO add more info and enhance rootCmd

func NewRootCmd(out io.Writer, args []string) *cobra.Command{
	cmd := &cobra.Command{
		Use: "adopt",
		Short: "adopt cluster resources into a new helm chart",
		Long: "adopt cluster resources into a new helm chart",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("no argument accepted")
			}
			return nil
		},
	}
	flags := cmd.PersistentFlags()
	flags.Parse(args)

	Settings = new(EnvSettings)
	if ctx := os.Getenv("HELM_KUBECONTEXT"); ctx != ""{
		Settings.KubeContext = ctx
	}
	cmd.AddCommand(NewResourcesCmd(out))

	return cmd
}