package handlers_test

import (
	"fmt"
	"net/http"

	"github.com/runar-rkmedia/skiver/handlers"
)

func ExampleStripBeforeLast() {
	fmt.Println(handlers.StripBeforeLast("/api/export/foo", "/"))
	// Output: foo
}

func ExampleExtractParams() {
	r, _ := http.NewRequest(http.MethodGet, "/api/export/foo=3&feature", nil)
	fmt.Println(handlers.ExtractParams(r))
	// Output: map[feature:[] foo:[3]] <nil>
}
func ExampleExtractParams_basic() {
	r, _ := http.NewRequest(http.MethodGet, "/api/export/?foo=3&feature", nil)
	fmt.Println(handlers.ExtractParams(r))
	// Output: map[feature:[] foo:[3]] <nil>
}
