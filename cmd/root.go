/*
Copyright © 2024 Tom Helander thomas.helander@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/prometheus/common/promlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sdtdclient "github.com/thelande/sdtd_client/pkg/sdtd_client"
)

const (
	envNamespace = "SDTD"
)

var (
	Client *sdtdclient.SDTDClient
	logger log.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sdtd_client",
	Short: "A 7 Days to Die Webserver API client.",
	// 	Long: `A longer description that spans multiple lines and likely contains
	// examples and usage of using your application. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	// status, err := Client.GetUserStatus()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%#v\n", status)
	// },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = promlog.New(&promlog.Config{})
		var err error

		Client, err = sdtdclient.NewSDTDClient(
			viper.GetString("host"),
			&sdtdclient.SDTDAuth{
				TokenName:   viper.GetString("token-name"),
				TokenSecret: viper.GetString("token-secret"),
			},
			true,
			&logger,
		)
		if err != nil {
			return err
		}

		err = Client.Connect()
		if err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.sdtd_client")
	viper.AddConfigPath(".")

	var cfgFile string
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.sdtd_client/config.yaml)")
	if len(cfgFile) != 0 {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, proceed with flags and defaults.
		} else {
			panic(err)
		}
	}

	rootCmd.PersistentFlags().StringP(
		"host",
		"H",
		"http://127.0.0.1:8080",
		fmt.Sprintf("Base URL of the API [env: %s_HOST]", envNamespace),
	)
	rootCmd.PersistentFlags().StringP(
		"token-name",
		"n",
		"",
		fmt.Sprintf("Name of the token to use [env: %s_TOKEN_NAME]", envNamespace),
	)
	rootCmd.PersistentFlags().StringP(
		"token-secret",
		"s",
		"",
		fmt.Sprintf("The token secret to use [env: %s_TOKEN_SECRET]", envNamespace),
	)

	rootCmd.MarkFlagRequired("host")
	rootCmd.MarkFlagRequired("token-name")
	rootCmd.MarkFlagRequired("token-secret")

	viper.SetEnvPrefix(envNamespace)
	viper.AutomaticEnv()

	for _, name := range []string{"host", "token-name", "token-secret"} {
		if err := viper.BindPFlag(name, rootCmd.PersistentFlags().Lookup(name)); err != nil {
			panic(err)
		}
	}

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
