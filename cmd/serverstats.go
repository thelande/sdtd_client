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

// serverstatsCmd represents the serverstats command
var serverstatsCmd = &cobra.Command{
	Use:   "serverstats",
	Short: "Collect and return the server stats",
	// Long: ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := Client.GetServerStats()
		if err != nil {
			return err
		}

		gameTime := resp.Data.GameTime
		table := pterm.TableData{
			{"Server Time", fmt.Sprintf("%d Days, %02d:%02d", gameTime.Days, gameTime.Hours, gameTime.Minutes)},
			{"Players", fmt.Sprintf("%d", resp.Data.Players)},
			{"Animals", fmt.Sprintf("%d", resp.Data.Animals)},
			{"Zombies", fmt.Sprintf("%d", resp.Data.Hostiles)},
		}
		pterm.DefaultTable.WithBoxed().WithData(table).Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverstatsCmd)
}
