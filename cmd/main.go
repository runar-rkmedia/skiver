package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/types"
)

func main() {

	inFile := "../../phonero/dinbedriftweb/src/locales/nb.json"
	outFile := "../../phonero/dinbedriftweb/src/locales/exported_keys.ts"
	createFileIfMissing := false
	overwrite := true

	_, err := os.Stat(inFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	f, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !createFileIfMissing {
				fmt.Println("Not creating file")
				os.Exit(1)
			}
			f, err = os.Create(outFile)
		} else {
			panic(err)
		}
	} else if !overwrite {
		fmt.Println("File exists, not overwriting")
		os.Exit(1)
	}
	endpoint := "http://localhost:8756/"
	api := NewAPI(endpoint)
	err = api.Login("jom", "jom")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	b, err := os.ReadFile(inFile)
	b = []byte(`{"nb":` + string(b) + "}")
	fmt.Println("importing", inFile)
	err = api.Import("dbw", "i18n", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
		return
	}

	if f == nil {
		panic("file is nil")
	}
	fmt.Println("exporting", outFile)
	err = api.Export("dbw", "typescript", "nb", f)
	if err != nil {
		fmt.Println(err)
		return
	}

}

type Api struct {
	endpoint string
	cookies  []*http.Cookie
	login    *types.LoginResponse
	client   *http.Client
}

func NewAPI(endpoint string) Api {
	c := http.Client{Timeout: time.Minute}
	return Api{
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &c,
	}
}

func (a *Api) Login(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("Missing username/password")
	}
	payload := struct{ Username, Password string }{Username: username, Password: password}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal login-payload: %w", err)
	}
	r, err := http.NewRequest(http.MethodPost, a.endpoint+"/api/login/", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create login-request: %w", err)
	}
	r.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("login-request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed reading body of login-request: %w", err)
	}
	if res.StatusCode >= 300 {
		var j models.APIError
		err = json.Unmarshal(body, &j)
		if err == nil {
			return fmt.Errorf("Login-request return a %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("Login-request return a %d-response: %s", res.StatusCode, string(body))
	}
	var j types.LoginResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return fmt.Errorf("failed reading body of login-request: %w", err)
	}
	a.cookies = res.Cookies()
	fmt.Println(res.StatusCode, res.Request.URL, r.Method, res.Request.Method, j)
	return nil

}
func (a Api) Import(projectName string, kind string, reader io.Reader) error {
	if len(a.cookies) == 0 {
		return fmt.Errorf("Not logged in")
	}
	r, err := http.NewRequest(http.MethodPost, a.endpoint+"/api/import/"+kind+"/"+projectName, reader)
	if err != nil {
		return fmt.Errorf("failed to create import-request: %w", err)
	}
	r.Header.Add("Content-Type", "application/json")
	for _, c := range a.cookies {
		r.AddCookie(c)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("import-request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed reading body of import-request: %w", err)
	}
	if res.StatusCode >= 300 {
		var j models.APIError
		err = json.Unmarshal(body, &j)
		if err == nil {
			return fmt.Errorf("import-request return a %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("import-request return a %d-response: %s", res.StatusCode, string(body))
	}
	var j handlers.ImportResult
	err = json.Unmarshal(body, &j)
	if err != nil {
		return fmt.Errorf("failed reading body of import-request: %w", err)
	}
	fmt.Println(res.StatusCode, res.Request.URL, r.Method, res.Request.Method)
	fmt.Printf("\nTranslationCreations: %d", len(j.Changes.TranslationCreations))
	fmt.Printf("\nCategoryCreations: %d", len(j.Changes.CategoryCreations))
	fmt.Printf("\nTranslationValueUpdates: %d", len(j.Changes.TranslationValueUpdates))
	fmt.Printf("\nTranslationsValueCreations: %d\n", len(j.Changes.TranslationsValueCreations))
	return nil

}

func (a Api) Export(projectName string, format string, locale string, writer io.Writer) error {
	if len(a.cookies) == 0 {
		return fmt.Errorf("Not logged in")
	}
	r, err := http.NewRequest(http.MethodGet, a.endpoint+"/api/export/", nil)
	if err != nil {
		return fmt.Errorf("failed to create export-request: %w", err)
	}
	q := r.URL.Query()
	q.Set("format", format)
	q.Set("locale", locale)
	q.Set("project", projectName)
	r.URL.RawQuery = q.Encode()
	r.Header.Add("Content-Type", "application/json")
	for _, c := range a.cookies {
		r.AddCookie(c)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("export-request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed reading body of export-request: %w", err)
	}
	fmt.Println(res.StatusCode, res.Request.URL, r.Method, res.Request.Method)
	if res.StatusCode >= 300 {
		var j models.APIError
		err = json.Unmarshal(body, &j)
		if err == nil {
			return fmt.Errorf("export-request return a %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("export-request return a %d-response: %s", res.StatusCode, string(body))
	}
	_, err = writer.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}
	return nil

}
