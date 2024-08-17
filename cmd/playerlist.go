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
	"github.com/spf13/viper"
	sdtdclient "github.com/thelande/sdtd_client/pkg/sdtd_client"
)

type PlayerWrapper struct {
	Players  []sdtdclient.Player
	PlayersM []sdtdclient.PlayerM
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the players known to the server.",
	Long: `List all players currently logged into the server. Use the -O flag
to list offline players (requires Alloc's Server Fixes Mod).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		offline := viper.GetBool("offline")
		stdFields := []string{
			"Name",
			"Entity ID",
			"Platform ID",
			"Online",
			"Last Online",
			"Playtime",
			"Location",
			"Ping (msec)",
		}
		table := pterm.TableData{stdFields}

		if !offline {
			resp, err := Client.GetOnlinePlayers()
			if err != nil {
				return err
			}
			for idx := range resp.Data.Players {
				player := &resp.Data.Players[idx]
				table = append(table, []string{
					player.Name,
					player.EntityID,
					player.PlatformID,
					fmt.Sprintf("%v", player.Online),
					player.LastOnline,
					player.GetPlaytime(),
					player.Position.GetCoordinates(),
					fmt.Sprintf("%v", player.Ping),
				})
			}
		} else {
			resp, err := Client.GetAllPlayersM()
			if err != nil {
				return err
			}
			for idx := range resp.Players {
				player := &resp.Players[idx]
				table = append(table, []string{
					player.Name,
					fmt.Sprintf("%v", player.EntityID),
					player.PlatformID,
					fmt.Sprintf("%v", player.Online),
					player.LastOnline,
					player.GetPlaytime(),
					player.Position.GetCoordinates(),
					fmt.Sprintf("%v", player.Ping),
				})
			}
		}

		pterm.DefaultTable.WithBoxed().WithHasHeader().WithData(table).Render()

		return nil
	},
}

func init() {
	playerCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("offline", "O", false, "Include offline players.")
	viper.BindPFlag("offline", listCmd.Flags().Lookup("offline"))
}
