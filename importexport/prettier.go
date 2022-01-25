package importexport

import (
	"fmt"
	"os/exec"
	"strings"
)

func Prettier(s string) (string, error) {
	// TODO: should we replace with https://github.com/robertkrimen/otto and include the prettier-js-code?
	// Or perhaps https://github.com/owenthereal/godzilla

	proc := exec.Command("prettier", "--stdin-filepath", "out.ts")
	// proc := exec.Command("node", "../frontend/prettier.js", "--stdin-filepath", "out.ts")
	r := strings.NewReader(s)
	proc.Stdin = r
	// stdErr := bytes.NewBufferString("")
	// proc.Stderr = stdErr

	buff, err := proc.Output()

	if err != nil {
		return string(buff), fmt.Errorf("Failed to prettify input: %w", err)
	}
	return string(buff), nil
}
