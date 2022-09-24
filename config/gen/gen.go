package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runar-rkmedia/skiver/config"
)

func writeJsonSchemaForConfig(path string) error {
	b, err := config.CreateJsonSchemaForConfig()
	if err != nil {
		return nil
	}
	return os.WriteFile(path, b, 0644)
}

func main() {
	p := "./config-schema.json"
	abs, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	writeJsonSchemaForConfig(abs)
	fmt.Printf("\nWrote file %s (%s)\n", p, abs)
}
