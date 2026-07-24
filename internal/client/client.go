package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const endpoint = "https://api.linear.app/graphql"

var (
	apiKey     string
	httpClient = &http.Client{}
)

// SetAPIKey sets the Linear personal API key used for the Authorization header.
func SetAPIKey(key string) {
	apiKey = strings.TrimSpace(key)
}

// HasAPIKey reports whether a key has been set.
func HasAPIKey() bool {
	return apiKey != ""
}

type gqlResp struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Query runs a GraphQL operation against the Linear API and unmarshals the
// `data` object into out.
func Query(gql string, variables map[string]any, out any) error {
	if apiKey == "" {
		return errors.New("no Linear API key set — run `linear auth login` or set LINEAR_API_KEY")
	}

	body, _ := json.Marshal(map[string]any{
		"query":     gql,
		"variables": variables,
	})

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Linear personal API keys are passed raw, with no Bearer/Basic prefix.
	req.Header.Set("Authorization", apiKey)

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	text, _ := io.ReadAll(res.Body)
	var gr gqlResp
	if err := json.Unmarshal(text, &gr); err != nil {
		return fmt.Errorf("Linear request failed: %d %s\n%s", res.StatusCode, res.Status, string(text))
	}

	if len(gr.Errors) > 0 {
		msgs := make([]string, len(gr.Errors))
		for i, e := range gr.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("Linear API error: %s", strings.Join(msgs, "; "))
	}

	if res.StatusCode == 401 || res.StatusCode == 403 {
		return errors.New("authentication failed — check your API key (401/403)")
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("Linear request failed: %d %s", res.StatusCode, res.Status)
	}
	if len(gr.Data) == 0 {
		return errors.New("Linear response missing data")
	}

	return json.Unmarshal(gr.Data, out)
}
