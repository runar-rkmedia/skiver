package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/runar-rkmedia/go-common/logger"
)

var (
	countrydataUrl = "https://github.com/gretzky/iso-locale-tools/raw/main/packages/countries/src/data/countries-expanded.json"
	l              = logger.InitLogger(logger.LogConfig{
		Level:      "debug",
		Format:     "human",
		WithCaller: false,
	})
)

type Countries []struct{}

func main() {

	cachePath := ".country-data.json"
	var countries Countries
	if !fileExists(cachePath) {
		l.Info().Str("cachePath", cachePath).Str("url", countrydataUrl).Msg("Cache not found, downloading")
		r, _ := http.NewRequest(http.MethodGet, countrydataUrl, nil)
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed during request")
		}
		if res.StatusCode >= 300 {
			log.Fatal("Go statuscode: ", res.StatusCode)
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to read body-data")
			log.Fatal(err)
		}
		// c, err := UnmarshalCountries(b)
		var c Countries
		err = json.Unmarshal(b, &c)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to unmarshal")
		}
		if err := ioutil.WriteFile(cachePath, b, 0677); err != nil {
			l.Fatal().Err(err).Str("cachePath", cachePath).Msg("Failed to write cache-file")
		}
		countries = c
	} else {
		l.Info().Str("cachePath", cachePath).Str("url", countrydataUrl).Msg("Using cached data")
		b, err := os.ReadFile(cachePath)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to read file-data")
			log.Fatal(err)
		}
		// c, err := UnmarshalCountries(b)
		var c Countries
		err = json.Unmarshal(b, &c)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to unmarshal")
		}
		countries = c
	}
	l.Info().Int("countries", len(countries)).Msg("Working set of countries")
	// b, err := yaml.Marshal(countries)
	// if err != nil {
	// 	l.Fatal().Err(err).Msg("Failed to marshal")

	// }
	// fmt.Println(string(b))

}

func fileExists(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		l.Fatal().Err(err).Str("path", path).Msg("Failed to read file")
	}
	return !s.IsDir()
}
