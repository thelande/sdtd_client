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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
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
func Get[R Response](c *SDTDClient, path string, resp *R, params *url.Values) error {
	body, err := c.Do("GET", path, params, nil)
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

// Perform a POST request against the API and return the populated response
// struct.
func Post[R Response](c *SDTDClient, path string, resp *R, params *url.Values, data []byte) error {
	body, err := c.Do("POST", path, params, data)
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

// Perform a DELETE request against the API and return the populated response
// struct.
func Delete[R Response](c *SDTDClient, path string, resp *R, params *url.Values, data []byte) error {
	body, err := c.Do("DELETE", path, params, data)
	if err != nil {
		return err
	}

	// fmt.Println(string(body))

	if len(body) > 0 {
		err = json.Unmarshal(body, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// Perform a GET request against the Alloc's Server Fixes API and return the
// populated response struct.
func GetM[R Response](c *SDTDClient, path string, resp *R, params *url.Values) error {
	if !c.allocsEnabled {
		return ErrAllocsModNotInstalled
	}
	return Get(c, path, resp, params)
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
func (c *SDTDClient) Do(method string, path string, params *url.Values, data []byte) ([]byte, error) {
	headers := c.GetHeaders()
	if method != "GET" && method != "DELETE" {
		headers["Content-Type"] = []string{"application/json"}
	}

	fullPath, err := url.JoinPath(c.Host, path)
	if err != nil {
		return nil, err
	}

	baseUrl, err := url.Parse(fullPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		baseUrl.RawQuery = params.Encode()
	}

	req, err := http.NewRequest(method, baseUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	if data != nil {
		req.Body = io.NopCloser(bytes.NewReader(data))
	}

	level.Debug(*c.logger).Log("url", baseUrl.String(), "method", method)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	level.Debug(*c.logger).Log("url", baseUrl.String(), "method", method, "statusCode", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		level.Warn(*c.logger).Log("status", resp.Status, "statusCode", resp.StatusCode, "body", body)
		return nil, ErrNon2XXResponse
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
	err := Get(c, path, &ServerStatsResponse{}, nil)
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
	err := Get(c, path, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Returns the server configuration.
func (c *SDTDClient) GetServerInfo() (*ServerInfoResponse, error) {
	path := "/api/serverinfo"
	status := ServerInfoResponse{}
	err := Get(c, path, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Returns the game preferences.
func (c *SDTDClient) GetGamePrefs() (*GamePrefsResponse, error) {
	path := "/api/gameprefs"
	status := GamePrefsResponse{}
	err := Get(c, path, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *SDTDClient) GetUserStatus() (*UserStatusResponse, error) {
	path := "userstatus"
	status := UserStatusResponse{}
	err := Get(c, path, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Returns the list of players currently online.
func (c *SDTDClient) GetOnlinePlayers() (*PlayersResponse, error) {
	path := "/api/player"
	players := PlayersResponse{}
	err := Get(c, path, &players, nil)
	if err != nil {
		return nil, err
	}
	return &players, nil
}

// Get an amount of lines from the server log
//
// count is the number of lines to fetch. If negative fetches count lines from
// the firstLine. Defaults to 50.
//
// firstLine is the first line number to fetch. Defaults to the oldest stored
// log line if count is positive. Defaults to the most recent log line if count
// is negative.
func (c *SDTDClient) GetLog(count *int, firstLine *int) (*LogResponse, error) {
	path := "/api/log"
	params := url.Values{}
	if count != nil {
		params.Add("count", fmt.Sprintf("%v", *count))
	}
	if firstLine != nil {
		params.Add("firstLine", fmt.Sprintf("%v", *firstLine))
	}

	log := LogResponse{}
	err := Get(c, path, &log, &params)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Fetch a list of all whitelisted users / groups.
func (c *SDTDClient) GetWhitelist() error {
	return nil
}

// Add a user to the whitelist.
func (c *SDTDClient) AddWhitelistUser(id string, name string) error {
	path := fmt.Sprintf("/api/whitelist/user/%v", id)
	data := WhitelistRequestBody{name}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = Post(c, path, &BaseResponse{}, nil, body)
	if err != nil {
		return err
	}
	return nil
}

// Remove a user from the whitelist.
func (c *SDTDClient) DeleteWhitelistUser(id string) error {
	path := fmt.Sprintf("/api/whitelist/user/%v", id)
	err := Delete(c, path, &BaseResponse{}, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
