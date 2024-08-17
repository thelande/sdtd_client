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

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type SDTDAuth struct {
	TokenName   string
	TokenSecret string
}

type SDTDClient struct {
	Host string
	Auth *SDTDAuth

	// Private fields
	allocsEnabled bool
	client        *http.Client
	logger        *log.Logger
}

// Perform a GET request against the API and return the populated response
// struct.
func Get[R Response](c *SDTDClient, path string, resp *R) error {
	body, err := c.Do("GET", path)
	if err != nil {
		return err
	}

	// fmt.Println(string(body))

	err = json.Unmarshal(body, resp)
	if err != nil {
		return err
	}

	return nil
}

func NewSDTDClient(host string, auth *SDTDAuth, sslVerify bool, logger *log.Logger) (*SDTDClient, error) {
	if len(host) == 0 {
		return nil, ErrNoHostSet
	}
	if host[0:7] != "http://" && host[0:8] != "https://" {
		return nil, ErrInvalidHostScheme
	}
	if auth == nil {
		return nil, ErrNilAuth
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify},
	}
	client := SDTDClient{
		Host:          host,
		Auth:          auth,
		allocsEnabled: false,
		client: &http.Client{
			Transport: tr,
		},
		logger: logger,
	}
	return &client, nil
}

// Return the authentication headers for communicating with the API server.
func (c *SDTDClient) GetHeaders() http.Header {
	headers := http.Header{}
	headers.Add("Accept", "application/json")

	// Manually set the header names since Header.Add uses CanonicalMIMEHeaderKey
	// to set the name, which mucks with the name. Alloc's Server Fixes expect
	// the headers to be entirely upper-case (while the vanilla web server does
	// not care).
	headers["X-SDTD-API-TOKENNAME"] = []string{c.Auth.TokenName}
	headers["X-SDTD-API-SECRET"] = []string{c.Auth.TokenSecret}
	return headers
}

// Make a request against the API.
func (c *SDTDClient) Do(method string, path string) ([]byte, error) {
	headers := c.GetHeaders()

	url, err := url.JoinPath(c.Host, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	level.Debug(*c.logger).Log("url", url, "method", method)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	level.Debug(*c.logger).Log("url", url, "method", method, "statusCode", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		level.Warn(*c.logger).Log("status", resp.Status, "statusCode", resp.StatusCode)
		return nil, ErrNon2XXResponse
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Attempt to connect to the API and verify the credentials. Also determines if
// Alloc's server fixes are available.
func (c *SDTDClient) Connect() error {
	if _, err := c.GetServerInfo(); err != nil {
		return err
	}
	level.Debug(*c.logger).Log("msg", "Server responded, checking for Alloc's Server Fixes APIs")

	path := "/api/getstats"
	err := Get(c, path, &ServerStatsResponse{})
	if err != nil && !errors.Is(err, ErrNon2XXResponse) {
		return err
	} else if err != nil {
		level.Warn(*c.logger).Log("msg", "Failed to detect Alloc's Server Fixes API")
	} else {
		level.Info(*c.logger).Log("msg", "Alloc's Server Fixes detected")
		c.allocsEnabled = true
	}

	return nil
}

// Returns statistics for the server (current game time and number of players,
// animals, and hostiles).
func (c *SDTDClient) GetServerStats() (*ServerStatsResponse, error) {
	path := "/api/serverstats"
	status := ServerStatsResponse{}
	err := Get(c, path, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Returns the server configuration.
func (c *SDTDClient) GetServerInfo() (*ServerInfoResponse, error) {
	path := "/api/serverinfo"
	status := ServerInfoResponse{}
	err := Get(c, path, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *SDTDClient) GetUserStatus() (*UserStatusResponse, error) {
	path := "userstatus"
	status := UserStatusResponse{}
	err := Get(c, path, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Returns the list of players currently online.
func (c *SDTDClient) GetOnlinePlayers() (*PlayersResponse, error) {
	path := "/api/player"
	players := PlayersResponse{}
	err := Get(c, path, &players)
	if err != nil {
		return nil, err
	}
	return &players, nil
}

// Returns all players known to the server. Requires Alloc's Server Fixes Mod.
func (c *SDTDClient) GetAllPlayersM() (*PlayersResponseM, error) {
	path := "/api/getplayerlist"
	players := PlayersResponseM{}
	if !c.allocsEnabled {
		return &players, nil
	}

	err := Get(c, path, &players)
	if err != nil {
		return nil, err
	}
	return &players, nil
}
