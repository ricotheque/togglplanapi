/*
Package togglplanapi provides a client interface for interacting
with the Toggl Plan API, facilitating authentication and sending
HTTP requests to the Toggl Plan API endpoints.

Example Usage:

	import (
		"fmt"
		"github.com/ricotheque/togglplanapi"
	)

	func main() {
		// Initialize a new Toggl Plan API client
		pa := togglplanapi.New(username, password, clientId, clientSecret, "")

		// Send a GET request to the Toggl Plan API
		result, err := togglplanapi.Request(pa, "https://api.plan.toggl.com/api/v5/me", "GET", []byte{}, map[string]string{})

		fmt.Println(result, err)

		// You can reuse the same client instance for another request
		result2, err2 := togglplanapi.Request(pa, "https://api.plan.toggl.com/api/v5/me", "GET", []byte{}, map[string]string{})

		fmt.Println(result2, err2)
	}
*/
package togglplanapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// togglPlanApi represents the client structure for Toggl Plan API.
// It holds necessary authentication details to perform API requests.
type togglPlanApi struct {
	username     string
	password     string
	clientId     string
	clientSecret string
	bearerToken  string
}

// authDetails represents authentication details required for making API requests.
type authDetails struct {
	Type       string // "Basic", "Bearer", etc.
	Credential string
}

// New initializes and returns a new togglPlanApi instance.
func New(username string, password string, clientId string, clientSecret string, bearerToken string) *togglPlanApi {
	return &togglPlanApi{
		username:     username,
		password:     password,
		clientId:     clientId,
		clientSecret: clientSecret,
		bearerToken:  bearerToken,
	}
}

// Request sends an authenticated request to the Toggl Plan API.
// If bearerToken is not set, it attempts to fetch a new token.
// Arguments:
//
//	pa: togglPlanApi instance
//	url: The API endpoint
//	method: HTTP method (GET, POST, etc.)
//	body: Request body, if any (use `[]byte{}` if you're not passing a body)
//	headers: Additional request headers (use `map[string]string{}` if you don't have an additional headers)
func Request(pa *togglPlanApi, url string, method string, body []byte, headers map[string]string) (string, error) {
	if pa.bearerToken == "" {
		result, err := getToken(pa)
		if err == nil {
			pa.bearerToken = result
		} else {
			return "Couldn't authenticate", err
		}
	}

	defaultHeaders := map[string]string{
		"Content-Type": "application/json",
	}

	finalHeaders := mergeMaps(defaultHeaders, headers)

	auth := &authDetails{
		Type:       "Bearer",
		Credential: pa.bearerToken,
	}

	result, err := doRequest(pa, url, method, body, finalHeaders, auth)

	return result, err
}

// doRequest is a helper function to send an API request.
// It includes retry logic for certain HTTP status codes.
// Arguments:
//
//	pa: togglPlanApi instance
//	url: The API endpoint
//	method: HTTP method (GET, POST, etc.)
//	body: Request body, if any
//	headers: Additional request headers
//	auth: Authentication details
func doRequest(pa *togglPlanApi, url string, method string, body []byte, headers map[string]string, auth *authDetails) (string, error) {
	client := retryablehttp.NewClient()

	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return true, nil
		}
		if resp.StatusCode == 401 {
			return false, nil
		}
		if resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented {
			return true, nil
		}
		return false, nil
	}

	client.RetryMax = 5
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 30 * time.Second

	req, err := retryablehttp.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "Error building request", err
	}

	for headerKey, headerValue := range headers {
		req.Header.Set(headerKey, headerValue)
	}

	if auth != nil {
		req.Header.Set("Authorization", auth.Type+" "+auth.Credential)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "Error running request", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "Unauthorized", errors.New("401")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprint(resp.StatusCode), errors.New(http.StatusText(resp.StatusCode))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error reading response", err
	}

	return string(bodyBytes), nil
}

// getToken fetches a new authentication token for the Toggl Plan API.
// It uses the client ID, client secret, username, and password to fetch the token.
// Arguments:
//
//	pa: togglPlanApi instance
func getToken(pa *togglPlanApi) (string, error) {
	type TokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	auth := &authDetails{
		Type:       "Basic",
		Credential: base64.StdEncoding.EncodeToString([]byte(pa.clientId + ":" + pa.clientSecret)),
	}

	body := []byte("grant_type=password&username=" + pa.username + "&password=" + pa.password)

	result, err := doRequest(pa, "https://api.plan.toggl.com/api/v5/authenticate/token", "POST", body, headers, auth)

	if err == nil {
		var tokenResponse TokenResponse
		err = json.Unmarshal([]byte(result), &tokenResponse)

		if err != nil {
			return "Couldn't parse authentication attempt response", err
		}

		if tokenResponse.AccessToken == "" {
			return "", errors.New("access_token not found in response")
		}

		return tokenResponse.AccessToken, nil
	} else {
		return "Couldn't request for a new bearer token", err
	}
}

// GetToken retrieves the bearerToken of the specified togglPlanApi instance,
// so that it can be stored for later user as needed.
func GetToken(pa *togglPlanApi) string {
	return pa.bearerToken
}

// mergeMaps takes two map[string]string instances as input and returns a new map
// that contains all the key-value pairs from both input maps.
//
// If the same key exists in both maps, the value from the second map will
// overwrite the value from the first map in the resulting merged map.
// Arguments:
//
// - map1: The first map to merge.
// - map2: The second map to merge.
func mergeMaps(map1, map2 map[string]string) map[string]string {
	mergedMap := make(map[string]string)

	for key, value := range map1 {
		mergedMap[key] = value
	}

	for key, value := range map2 {
		mergedMap[key] = value
	}

	return mergedMap
}
