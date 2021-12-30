package translator

import (
	"fmt"
	"net/http"
	"time"
)

type TranslatorProvider interface {
	Translate(text, from, to string) (string, error)
}

type TranslatorOptions struct {
	Kind       string
	ApiToken   string
	Endpoint   string
	HttpClient *http.Client
}

func NewTranslator(options TranslatorOptions) (TranslatorProvider, error) {
	if options.HttpClient == nil {
		options.HttpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	switch options.Kind {
	case "libre":
		libre, err := NewLibreTranslator(LibreOptions{
			Url:    options.Endpoint,
			ApiKey: options.ApiToken,
			Client: options.HttpClient,
		})
		return &libre, err
	case "bing":
		bing, err := NewBingTranslator(BingTranslatorOptions{
			KeyA:     options.ApiToken,
			Endpoint: options.Endpoint,
			Client:   options.HttpClient,
		})
		return &bing, err

	}
	return nil, fmt.Errorf("No valid translator could be initiated")
}
