package sourcemap

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
)

// TODO: don't use chroma for this... it allocates and it is quite slow as it has a very different purpose than this.
// TODO: create variants of this to support toml and yaml.

// JsonKeys creates a map of all the key-paths within a json-file pointing to the line-number/offset of the json-file
// This is mostly used to be able to use git blame on the file on the client-side to get a timestamp for the change,
// to make it easier to autoresolve change-conflicts.
// Asumptions for the input-file:
// - Valid Json
// - Each json-value is on a single line. I don't think JSON supports values spanning multiple lines
func JsonKeys(tokens []Token) map[string]SpanToken {
	keyMap := map[string]SpanToken{}
	jsonKeys(keyMap, tokens, []string{}, 0, -1)
	return keyMap

}

func SourceMapperSupports(contentType string) bool {
	switch contentType {
	case "json", "application/json", "application/json; charset=utf-8":
		return true
	}
	return false
}
func MapToSourceFromTokens(contentType string, tokens []Token) (map[string]SpanToken, error) {
	if !SourceMapperSupports(contentType) {
		return nil, fmt.Errorf("contentType not supported for SourceMapper: %s", contentType)
	}
	return JsonKeys(tokens), nil
}
func MapToSource(contentType string, content string) (map[string]SpanToken, error) {
	if !SourceMapperSupports(contentType) {
		return nil, fmt.Errorf("contentType not supported for SourceMapper: %s", contentType)
	}
	tokenizer, err := NewTokenizer("import.json", content)
	if err != nil {
		return nil, err
	}
	var tokens []Token
	tokenizer.TokensWithOffsets(func(token Token) bool {
		tokens = append(tokens, token)
		return false
	})

	return MapToSourceFromTokens(contentType, tokens)
}

func jsonKeys(keyMap map[string]SpanToken, tokens []Token, prefix []string, index int, arrayIndex int) int {
	// keyMap := map[string]SpanToken{}
	i := index
	currentPath := NewJsonPath()
	for _, v := range prefix {
		currentPath.Add(v)
	}
	for ; i < len(tokens); i++ {
		tok := (tokens)[i]
		trimmed := strings.TrimSpace(tok.Value)
		switch {
		case trimmed == "":
			continue
		case tok.Type == chroma.NameTag:
			currentPath.SetQueue(trimOne(tok.Value, `"`))
			key := currentPath.String()
			keyMap[key] = SpanToken{
				Token: tok,
				Path:  currentPath.path,
			}
		case arrayIndex >= 0 && (IsValueLike(tok.Token)):
			currentPath.AddArrayElement()
			key := currentPath.String() //  + "." + strconv.FormatInt(int64(arrayIndex), 10)
			arrayIndex++
			keyMap[key] = SpanToken{
				Token: tok,
				Path:  currentPath.path,
			}
		case tok.Type == chroma.Punctuation:
			switch tok.Value {
			case "[":
				if arrayIndex >= 0 {
					currentPath.AddArrayElement()
					key := currentPath.String() //  + "." + strconv.FormatInt(int64(arrayIndex), 10)
					arrayIndex++
					keyMap[key] = SpanToken{
						Token: tok,
						Path:  currentPath.path,
					}
				}
				newPath := currentPath.path
				if currentPath.Queue != "" {
					newPath = append(newPath, currentPath.Queue)
				}
				parsed := jsonKeys(keyMap, tokens, newPath, i+1, 0)
				// for k, v := range toks {
				// 	keyMap[k] = v
				// }
				i += parsed + 1
			case "{":
				if arrayIndex >= 0 {
					currentPath.AddArrayElement()
					key := currentPath.String() //  + "." + strconv.FormatInt(int64(arrayIndex), 10)
					arrayIndex++
					keyMap[key] = SpanToken{
						Token: tok,
						Path:  currentPath.path,
					}
				}
				newPath := currentPath.path
				if currentPath.Queue != "" {
					newPath = append(newPath, currentPath.Queue)
				}
				parsed := jsonKeys(keyMap, tokens, newPath, i+1, -1)
				i += parsed + 1
			case "}", "]":
				return i - index
			}
		}

	}

	return len(tokens) - index

}
