package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func SimpleRequest(client *http.Client, method, url string, requestBody, responseBody interface{}) (*http.Response, error) {
	var body io.Reader
	if requestBody != nil {
		b, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal body: %w", err)
		}
		body = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return res, fmt.Errorf("failed to get languages: %w", err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return res, fmt.Errorf("failed to read body: %w", err)
	}
	defer res.Body.Close()
	err = json.Unmarshal(resBody, &responseBody)
	if err != nil {
		return res, fmt.Errorf("failed to unmarshal responseBody: %w (%s) url: %s", err, string(resBody), url)
	}
	if res.StatusCode >= 300 {
		return res, fmt.Errorf("Non 2xx-statuscode returned: %d %s %s url: %s", res.StatusCode, res.Status, string(resBody), url)
	}
	return res, nil
}
