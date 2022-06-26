// Package sourcemap provides a simplified sourcemap for data-files.
// Its primary use is to map keypaths to the correct linenumber / offsets within the data-file.
package sourcemap

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"

	"github.com/r3labs/diff/v2"
)

// DiffOfObjects returns a changelog with options set for use with for instance i18n-json.
func DiffOfObjects(a, b interface{}) (diff.Changelog, error) {
	return diff.Diff(a, b, diff.DisableStructValues(), diff.AllowTypeMismatch(true))
}

type Tokenizer struct {
	iterator chroma.Iterator
	tokens   []chroma.Token
	index    int
	FilePath string
	Lexer    chroma.Config
}

func NewTokenizer(filepath string, content string) (Tokenizer, error) {
	lexer := lexers.Match(filepath)
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		return Tokenizer{}, fmt.Errorf("failed to identify content-type for file '%s'", filepath)
	}
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return Tokenizer{}, fmt.Errorf("failed to tokenize content: %w", err)
	}
	t := Tokenizer{
		iterator: iterator,
		index:    0,
		FilePath: filepath,
		Lexer:    *lexer.Config(),
	}
	return t, nil
}

type Token struct {
	chroma.Token
	Start Offset
	End   Offset
}

func (t Token) String() string {
	return fmt.Sprintf("'%s=>%s=>%s' (%s) :L%d", t.Type.Category(), t.Type.SubCategory(), t.Type, t.Value, t.Start.Line)
}

type Offset struct {
	Offset int
	Line   int
}

type SpanToken struct {
	Path []string
	// not ready for use yet
	End *Offset `json:"-"`
	Token
}

func trimOne(s string, cut string) string {
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, cut) {
		s = s[1:]
	}
	if strings.HasSuffix(s, cut) {
		s = s[:len(s)-1]
	}
	return s
}

func IsValueLike(t chroma.Token) bool {
	if t.Type == chroma.KeywordConstant {
		if t.Value == "true" || t.Value == "false" {
			return true
		}
	}
	if t.Type.InCategory(chroma.Text) {
		return true
	}
	if t.Type.InCategory(chroma.Literal) {
		return true
	}
	return false
}

// Returns offsets for each token
func (t *Tokenizer) TokensWithOffsets(consumer func(token Token) bool) {
	var offset int
	linenumber := 1
	for tok := t.iterator(); ; tok = t.iterator() {
		token := Token{
			Token: tok,
		}
		token.Start = Offset{
			Offset: offset,
			Line:   linenumber,
			// TODO: add column
		}
		linenumber += strings.Count(tok.Value, "\n")
		offset += len(tok.Value)
		token.End = Offset{
			Offset: offset,
			Line:   linenumber,
		}
		if consumer(token) {
			return
		}
		if tok == chroma.EOF {
			return
		}
	}

}
func (t Tokenizer) Tokens() []chroma.Token {
	if t.tokens != nil && len(t.tokens) != 0 {
		return t.tokens
	}
	return t.iterator.Tokens()
}
func (t Tokenizer) IsChanged() bool {
	return t.tokens != nil
}
func (t *Tokenizer) SetTokens(tokens []chroma.Token) {
	t.tokens = tokens
}
func (t Tokenizer) Concat() string {
	return chroma.Stringify(t.Tokens()...)
}
