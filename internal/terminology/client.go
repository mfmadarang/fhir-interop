package terminology

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://tx.fhir.org/r4/CodeSystem/$validate-code"

const (
	SystemLOINC  = "http://loinc.org"
	SystemSNOMED = "http://snomed.info/sct"
)

type Result struct {
	Valid   bool
	Message string
	Display string
}

type parametersResponse struct {
	Parameter []struct {
		Name         string `json:"name"`
		ValueBoolean *bool  `json:"valueBoolean,omitempty"`
		ValueString  string `json:"valueString,omitempty"`
	} `json:"parameter"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) ValidateCode(ctx context.Context, system, code string) (*Result, error) {
	q := url.Values{}
	q.Set("url", system)
	q.Set("code", code)

	reqURL := baseURL + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("terminology: build request: %w", err)
	}
	req.Header.Set("Accept", "application/fhir+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("terminology: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("terminology: unexpected status %d", resp.StatusCode)
	}

	var parsed parametersResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("terminology: decode response: %w", err)
	}

	result := &Result{}
	for _, p := range parsed.Parameter {
		switch p.Name {
		case "result":
			if p.ValueBoolean != nil {
				result.Valid = *p.ValueBoolean
			}
		case "message":
			result.Message = p.ValueString
		case "display":
			result.Display = p.ValueString
		}
	}

	return result, nil
}
