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
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// serverinfoCmd represents the serverinfo command
var serverinfoCmd = &cobra.Command{
	Use:   "serverinfo",
	Short: "Return the server configuration.",
	Long:  `Returns the contents of serverconfig.xml as a table of settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := Client.GetServerInfo()
		if err != nil {
			return err
		}

		table := pterm.TableData{
			{"Setting", "Type", "Value"},
		}
		for idx := range resp.Data {
			setting := &resp.Data[idx]
			table = append(table, []string{setting.Name, setting.Type, fmt.Sprintf("%v", setting.Value)})
		}

		pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(table).Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverinfoCmd)
}
