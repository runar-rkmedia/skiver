package sourcemap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/utils"
)

func TestJsonKeys(t *testing.T) {
	type args struct {
		filepath string
		content  string
	}
	type linemap = map[string]int
	tests := []struct {
		name        string
		args        args
		wantLineMap linemap
		wantErr     bool
	}{
		{
			"Simple singleline json",
			args{
				"simple.json",
				`{"foo": 1, "bar": "baz"}`,
			},
			linemap{"foo": 1, "bar": 1},
			false,
		},
		{
			"Simple multiline json",
			args{
				"simple.json",
				`{
					"foo": 1, 
					"bar": "baz"
					}`,
			},
			linemap{"foo": 2, "bar": 3},
			false,
		},
		{
			"Simple with object-json",
			args{
				"simple.json",
				`{
					"int-kind": 1, 
					"object-kind": {},
					"text": "baz"
					}`,
			},
			linemap{"int-kind": 2, "text": 4, "object-kind": 3},
			false,
		},
		{
			"Simple nested object-json",
			args{
				"simple.json",
				`{
					"int-kind": 1, 
					"nested": {
					  "inner-text": "lorem",
					  "inner-int": 42
					},
					"text": "baz"
					}`,
			},
			linemap{"int-kind": 2, "text": 7, "nested": 3, "nested.inner-text": 4, "nested.inner-int": 5},
			false,
		},
		{
			"Simple nested object-json with array",
			args{
				"simple.json",
				`{
"int-kind": 1, 
"nested": {
	"inner-text": "lorem",
	"inner-int": 42,
	"list": [    1,
					2,
					  3,


					"foo-text"]
},
"text": "baz"
}`,
			},
			linemap{
				"int-kind":          2,
				"nested":            3,
				"nested.inner-text": 4,
				"nested.inner-int":  5,
				"nested.list":       6,
				"nested.list.0":     6,
				"nested.list.1":     7,
				"nested.list.2":     8,
				"nested.list.3":     11,
				"text":              13,
			},
			false,
		},
		{
			"Multiple nested object-json",
			args{
				"simple.json",
				`{
				"one": {
									"two": {
					"three": {
					"four": {
					  "five": true,
					  "six": false, "seven": null
					}
					}
					}
									}
}`,
			},
			linemap{
				"one":                      2,
				"one.two":                  3,
				"one.two.three":            4,
				"one.two.three.four":       5,
				"one.two.three.four.five":  6,
				"one.two.three.four.six":   7,
				"one.two.three.four.seven": 7,
			},
			false,
		},
		{
			"Multiple nested object-json with arrays",
			args{
				"simple.json",
				`{
				"one": [
					"two",
					{
					  "three": [
					  "four"
					]
					}
					] 
}`,
			},
			linemap{
				"one":           2,
				"one.0":         3,
				"one.1":         4,
				"one.1.three":   5,
				"one.1.three.0": 6,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check that the json is valid
			j := map[string]interface{}{}
			err := json.Unmarshal([]byte(tt.args.content), &j)
			if err != nil {
				t.Fatalf("json is invalid: %v", err)
			}

			tokenizer, err := NewTokenizer(tt.args.filepath, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenizeSourceFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			lines := strings.Split(tt.args.content, "\n")
			tokens := []Token{}
			tokenizer.TokensWithOffsets(func(token Token) bool {
				tokens = append(tokens, token)
				return false
			})
			jkeys := JsonKeys(tokens)
			keys := utils.SortedMapKeys(jkeys)
			lm := make(linemap)
			for _, key := range keys {
				k := jkeys[key]
				lm[key] = k.Start.Line
				line := lines[k.Start.Line-1]
				if !strings.Contains(line, k.Value) {
					t.Errorf("Expcted SpanToken.Value (%s) to exist at specified linenumber(%d) from SpanToken.Start.Line, but it was not found", k.Value, k.Start.Line)
					return
				}

			}
			if err := internal.Compare("map for path to linenumber should match wanted result", lm, tt.wantLineMap); err != nil {
				t.Error(err)
			}
		})
	}
}

func createMockJsonString(size, depth int) string {
	jmap := createMockJson(size, depth)
	b, _ := json.MarshalIndent(jmap, "", "  ")
	// b, _ := json.Marshal(jmap)
	return string(b)
}
func createMockJson(size, depth int) map[string]interface{} {
	jmap := map[string]interface{}{}
	for i := 0; i < size; i++ {
		id := strconv.FormatInt(int64(i), 10)
		o := map[string]interface{}{
			"foo": "bar",
			"baz": "fobar",
		}
		if depth >= 0 {
			o["data"] = createMockJson(size, depth-1)
		}
		jmap[id] = o
	}
	return jmap
}

func BenchmarkJsonPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jp := NewJsonPath()
		jp.Add("foo")
		jp.Add("bar")
		jp.AddArrayElement()
		jp.AddArrayElement()
		jp.Add("bar")
		_ = jp.String()

	}

}
func BenchmarkJsonTokens(b *testing.B) {
	testSize := [][2]int{
		{0, 0},
		{1, 1},
		{10, 2},
		{5, 5},
	}
	for _, size := range testSize {
		jsonStr := createMockJsonString(size[0], size[1])
		b.Run(fmt.Sprintf("json_%dx%d:%s", size[0], size[1], humanize.Bytes(uint64(len(jsonStr)))), func(b *testing.B) {

			tokenizer, err := NewTokenizer("big.json", jsonStr)
			if err != nil {
				b.Errorf("failed to create tokenizer: %v", err)
			}
			tokens := []Token{}
			tokenizer.TokensWithOffsets(func(token Token) bool {
				tokens = append(tokens, token)
				return false
			})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				JsonKeys(tokens)
			}
		})
	}
}
