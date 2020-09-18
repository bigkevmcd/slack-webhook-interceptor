package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bigkevmcd/slack-webhook-interceptor/pkg/interception"
)

const (
	portFlag   = "port"
	secretFlag = "secret"
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
}

// Execute is the main entry point into this component.
func Execute() {
	if err := makeHTTPCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func makeHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interceptor",
		Short: "intercept trigger hooks and convert from Slack",
		Run: func(cmd *cobra.Command, args []string) {
			http.HandleFunc("/", interception.MakeHandler(viper.GetString(secretFlag)))
			addr := fmt.Sprintf(":%d", viper.GetInt(portFlag))
			log.Printf("Listening on %s\n", addr)
			http.ListenAndServe(addr, nil)
		},
	}

	cmd.Flags().Int(
		portFlag,
		8080,
		"port to serve requests on",
	)
	logIfError(viper.BindPFlag(portFlag, cmd.Flags().Lookup(portFlag)))

	cmd.Flags().String(
		secretFlag,
		"",
		"shared secret for authenticating Slack hooks",
	)
	logIfError(viper.BindPFlag(secretFlag, cmd.Flags().Lookup(secretFlag)))
	return cmd
}

func logIfError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
