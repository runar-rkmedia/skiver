package translator

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

type Translator struct {
	ctx context.Context
	// g   *translate.Client
	SupportedLanguages map[string]string
	sync.RWMutex
	LibreOptions
}

type LibreOptions struct {
	Url, ApiKey string
	Client      *http.Client
}

// Based on https://libretranslate.com/docs/#/
// See also https://github.com/LibreTranslate/LibreTranslate
// There is also a more complete binding at https://github.com/LibreTranslate/LibreTranslate
func NewLibreTranslator(options ...LibreOptions) (t Translator, err error) {
	if len(options) > 0 {
		t.LibreOptions = options[0]
	}
	if t.Client == nil {
		t.Client = http.DefaultClient
	}
	if t.Url == "" {
		t.Url = "https://libretranslate.de/"
	}
	go func() {
		s, err := t.GetLanguages()
		if err != nil {
			return
		}
		t.Lock()
		t.SupportedLanguages = map[string]string{}
		for i := 0; i < len(s); i++ {
			t.SupportedLanguages[s[i].Code] = s[i].Name
		}

	}()
	return
}

type LibreTranslateInput struct {
	TextSource string `json:"q"`
	SourceLang string `json:"source"`
	TargetLang string `json:"target"`
	Format     string `json:"format"`
	ApiKey     string `json:"api_key,omitempty"`
}

type SupportedLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (t *Translator) GetLanguages() ([]SupportedLanguage, error) {
	var j []SupportedLanguage
	_, err := SimpleRequest(t.Client, http.MethodGet, t.Url+"languages", nil, &j)
	return j, err
}

func (t *Translator) IsValidLanguage(lang ...string) error {
	t.RLock()
	for i := 0; i < len(lang); i++ {
		if _, ok := t.SupportedLanguages[lang[i]]; !ok {
			return fmt.Errorf("%s is not supported", lang[i])
		}
	}
	defer t.RUnlock()
	return nil
}
func (t *Translator) Translate(text, from, to string) (string, error) {
	input := LibreTranslateInput{
		SourceLang: from,
		TargetLang: to,
		Format:     "text",
		TextSource: text,
	}
	if t.ApiKey != "" {
		input.ApiKey = t.ApiKey
	}

	var j struct{ TranslatedText string }
	_, err := SimpleRequest(t.Client, http.MethodPost, t.Url+"translate", &input, &j)
	return j.TranslatedText, err
}
