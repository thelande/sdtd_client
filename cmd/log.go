/*
Copyright © 2024 Tom Helander <thomas.helander@gmail.com>

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
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sdtdclient "github.com/thelande/sdtd_client/pkg/sdtd_client"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Retrieve the server logs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var count, firstLine *int

		if viper.GetString("log.count") != "" {
			count = new(int)
			*count = viper.GetInt("log.count")
		}

		if viper.GetString("log.firstline") != "" {
			firstLine = new(int)
			*firstLine = viper.GetInt("log.firstline")
		}

		log, err := Client.GetLog(count, firstLine)
		if err != nil {
			return err
		}

		headers := []string{"Timestamp", "Time since boot", "Sev", "Message"}
		tableData := pterm.TableData{headers}
		for idx := range log.Data.Entries {
			entry := &log.Data.Entries[idx]
			uptime, err := strconv.ParseInt(entry.UptimeMs, 10, 64)
			if err != nil {
				return err
			}

			tsb := sdtdclient.SecondsToDaysHoursMinutesSeconds(int(uptime) / 1000)
			tableData = append(tableData, []string{
				entry.IsoTime,
				tsb,
				entry.Type,
				entry.Msg,
			})
		}

		pterm.DefaultTable.WithBoxed().WithHasHeader().WithData(tableData).Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().StringP(
		"count",
		"C",
		"",
		`The number of lines to fetch.
If negative fetches count lines from the firstline`,
	)

	logCmd.Flags().StringP(
		"firstline",
		"F",
		"",
		`The first line number to fetch.
Defaults to the oldest stored log line if count is positive.
Defaults to the most recent log line if count is negative`,
	)

	viper.BindPFlag("log.count", logCmd.Flags().Lookup("count"))
	viper.BindPFlag("log.firstline", logCmd.Flags().Lookup("firstline"))
}
