package importexport

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/utils"
)

func ExampleFlattenStringMap() {
	data := map[string]interface{}{
		"General": map[string]interface{}{
			"Create": "Create it",
			"Add":    "Add it",
		},
	}

	described := map[string]string{}

	err := FlattenStringMap("", data, described, false)
	sorted := utils.SortedMapKeys(described)
	fmt.Println(err)
	for _, k := range sorted {
		fmt.Println(k, described[k])
	}

	// Output:
	// <nil>
	// General.Add Add it
	// General.Create Create it
}
