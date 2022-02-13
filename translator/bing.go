package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	DefaultEndpoint = "https://api.cognitive.microsofttranslator.com/"
)

type BingTranslator struct {
	BingTranslatorOptions
}

type BingTranslatorOptions struct {
	KeyA, KeyB, Endpoint string
	Client               *http.Client
}

func NewBingTranslator(options BingTranslatorOptions) (BingTranslator, error) {

	if options.Endpoint == "" {
		options.Endpoint = DefaultEndpoint
	}
	return BingTranslator{options}, nil
}

// incomplete but good enough for now, see https://docs.microsoft.com/en-us/azure/cognitive-services/translator/reference/v3-0-translate#response-body
type bingTranslateResponse struct {
	Translations []struct {
		Text string `json:"text"`
		To   string `json:"to"`
	} `json:"translations"`
}

// Based on https://docs.microsoft.com/en-us/azure/cognitive-services/translator/reference/v3-0-translate
func (bt *BingTranslator) Translate(text, from, to string) (string, error) {

	type Text struct {
		Text string `json:"Text"`
	}
	type Input []Text
	input := Input{{Text: text}}
	b, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to return marshal input: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, bt.Endpoint+"translate", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", bt.KeyA)
	req.Header.Set("Ocp-Apim-Subscription-Region", "northeurope")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	q := req.URL.Query()
	q.Set("api-version", "3.0")
	q.Set("from", from)
	q.Set("to", to)

	req.URL.RawQuery = q.Encode()

	res, err := bt.Client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}
	defer res.Body.Close()

	var j []bingTranslateResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal: %w (%s)", err, string(body))
	}
	translated := ""
	if len(j) > 0 && len(j[0].Translations) > 0 {
		translated = j[0].Translations[0].Text
	}

	if res.StatusCode >= 300 {
		return translated, fmt.Errorf("Non 2xx-status returned: %d %s %s for url %s", res.StatusCode, res.Status, string(body), res.Request.URL)
	}

	return translated, nil

}
