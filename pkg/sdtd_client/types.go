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
package sdtdclient

import "errors"

var (
	ErrNon2XXResponse        = errors.New("received non 2XX status code")
	ErrNoHostSet             = errors.New("host not set")
	ErrInvalidHostScheme     = errors.New("the host scheme is invalid, must be http or https")
	ErrNilAuth               = errors.New("cannot use nil Auth")
	ErrAllocsModNotInstalled = errors.New("alloc's server fixes not installed")
)

type BaseResponse struct {
	Meta struct {
		ServerTime string `json:"serverTime"`
	} `json:"meta"`
}

type ServerInfoData struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type ServerInfoResponse struct {
	BaseResponse
	Data []ServerInfoData `json:"data"`
}

type AllowedVerbs struct {
	Get    bool `json:"GET"`
	Post   bool `json:"POST"`
	Put    bool `json:"PUT"`
	Delete bool `json:"DELETE"`
}

type Permission struct {
	Module  string       `json:"module"`
	Allowed AllowedVerbs `json:"allowed"`
}

type UserStatusData struct {
	Username        string       `json:"username"`
	LoggedIn        bool         `json:"loggedIn"`
	PermissionLevel int          `json:"permissionLevel"`
	Permissions     []Permission `json:"permissions"`
}

type UserStatusResponse struct {
	BaseResponse
	Data UserStatusData `json:"data"`
}

type ServerStatsData struct {
	GameTime struct {
		Days    int `json:"days"`
		Hours   int `json:"hours"`
		Minutes int `json:"minutes"`
	} `json:"gameTime"`
	Players  int `json:"players"`
	Hostiles int `json:"hostiles"`
	Animals  int `json:"animals"`
}

type ServerStatsResponse struct {
	BaseResponse
	Data ServerStatsData `json:"data"`
}

type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type KillsData struct {
	Zombies int `json:"zombies"`
	Players int `json:"players"`
}

type BannedData struct {
	BanActive bool   `json:"banActive"`
	Reason    string `json:"reason"`
	Until     string `json:"until"`
}

type Player struct {
	EntityID             string     `json:"entityId"`
	Name                 string     `json:"name"`
	PlatformID           string     `json:"platformId"`
	CrossPlatformID      string     `json:"crossplatformId"`
	TotalPlayTimeSeconds int        `json:"totalPlayTimeSeconds"`
	LastOnline           string     `json:"lastOnline"`
	Online               bool       `json:"online"`
	IP                   string     `json:"ip"`
	Ping                 int        `json:"ping"`
	Position             Location   `json:"position"`
	Level                int        `json:"level"`
	Health               int        `json:"health"`
	Stamina              float32    `json:"stamina"`
	Score                int        `json:"score"`
	Deaths               int        `json:"deaths"`
	Kills                KillsData  `json:"kills"`
	Banned               BannedData `json:"banned"`
}

// Alloc's Server Fixes Mod variant of the player data
type PlayerM struct {
	EntityID             int      `json:"entityid"`
	Name                 string   `json:"name"`
	PlatformID           string   `json:"steamid"`
	CrossPlatformID      string   `json:"crossplatformid"`
	TotalPlayTimeSeconds int      `json:"totalplaytime"`
	LastOnline           string   `json:"lastonline"`
	Online               bool     `json:"online"`
	IP                   string   `json:"ip"`
	Ping                 int      `json:"ping"`
	Position             Location `json:"position"`
	Banned               bool     `json:"banned"`
}

type PlayersData struct {
	Players []Player `json:"players"`
}

// Alloc's Server Fixes Mod variant of the player list response
type PlayersResponseM struct {
	Total   int       `json:"total"`
	Players []PlayerM `json:"players"`
}

type PlayersResponse struct {
	BaseResponse
	Data PlayersData `json:"data"`
}

type LogEntry struct {
	ID       int    `json:"id"`      // Consecutive ID/number of this log line
	Msg      string `json:"msg"`     // The log message
	Type     string `json:"type"`    // Severity type (Error, Assert, Warning, Log, Exception)
	Trace    string `json:"trace"`   // Stacktrace if entry is an Exception
	IsoTime  string `json:"isotime"` // Date/time of the log entry
	UptimeMs string `json:"uptime"`  // Time since server was started in milliseconds
}

type LogData struct {
	Entries   []LogEntry `json:"entries"`
	FirstLine int        `json:"firstLine"` // Number of first line retrieved
	LastLine  int        `json:"lastLine"`  // Number of next line to retrieve to follow up without missing entries
}

type LogResponse struct {
	BaseResponse
	Data LogData `json:"data"`
}

type Response interface {
	ServerInfoResponse |
		UserStatusResponse |
		ServerStatsResponse |
		PlayersResponse |
		PlayersResponseM |
		LogResponse
}
