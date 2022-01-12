package lexer

import (
	"testing"

	"github.com/runar-rkmedia/skiver/internal"
)

func TestLexer(t *testing.T) {
	defaultMap := map[string]TokenKind{
		"{{":  TokenPrefix,
		"}}":  TokenSuffix,
		",":   TokenFormatSeperator,
		"$t(": TokenNestingPrefix,
		")":   TokenNestingSuffix,
	}
	maxLoops := 10
	tests := []struct {
		name    string
		fields  string
		expects []Token
	}{

		{
			"Simple string",
			"foo bar {{count}} baz",
			[]Token{

				{Kind: TokenLiteral, Literal: "foo bar ", Start: 0, End: 8},
				{Kind: TokenPrefix, Literal: "{{", Start: 8, End: 10},
				{Kind: TokenLiteral, Literal: "count", Start: 10, End: 15},
				{Kind: TokenSuffix, Literal: "}}", Start: 15, End: 17},
				{Kind: TokenLiteral, Literal: " baz", Start: 17, End: 21},
				{Kind: TokenEOF, Literal: "", Start: 21, End: 21},
			},
		},
		{
			"With format",
			"foo bar {{ count , option }} baz",
			[]Token{
				{Kind: TokenLiteral, Literal: "foo bar ", Start: 0, End: 8},
				{Kind: TokenPrefix, Literal: "{{", Start: 8, End: 10},
				{Kind: TokenLiteral, Literal: " count ", Start: 10, End: 17},
				{Kind: TokenFormatSeperator, Literal: ",", Start: 17, End: 18},
				{Kind: TokenLiteral, Literal: " option ", Start: 18, End: 26},
				{Kind: TokenSuffix, Literal: "}}", Start: 26, End: 28},
				{Kind: TokenLiteral, Literal: " baz", Start: 28, End: 32},
				{Kind: TokenEOF, Literal: "", Start: 32, End: 32},
			},
		},
		{
			"With format and Nesting",
			"foo bar {{ count , option }} $t( gib ) baz",
			[]Token{
				{Kind: TokenLiteral, Literal: "foo bar ", Start: 0, End: 8},
				{Kind: TokenPrefix, Literal: "{{", Start: 8, End: 10},
				{Kind: TokenLiteral, Literal: " count ", Start: 10, End: 17},
				{Kind: TokenFormatSeperator, Literal: ",", Start: 17, End: 18},
				{Kind: TokenLiteral, Literal: " option ", Start: 18, End: 26},
				{Kind: TokenSuffix, Literal: "}}", Start: 26, End: 28},
				{Kind: TokenLiteral, Literal: " ", Start: 28, End: 29},
				{Kind: TokenNestingPrefix, Literal: "$t(", Start: 29, End: 32},
				{Kind: TokenLiteral, Literal: " gib ", Start: 32, End: 37},
				{Kind: TokenNestingSuffix, Literal: ")", Start: 37, End: 38},
				{Kind: TokenLiteral, Literal: " baz", Start: 38, End: 42},
				{Kind: TokenEOF, Literal: "", Start: 42, End: 42},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.fields, defaultMap)
			for i := 0; l.ch != 0; i++ {
				l.FindAllTokens()
				// t.Log(i, string(l.ch), tok)
				if i >= maxLoops {
					t.Errorf("max-loops-reached %#v", l)
					return
				}
			}
			got := l.Tokens

			if err := internal.Compare("result", got, tt.expects, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				Yaml:    false,
				JSON:    true,
			}); err != nil {
				t.Log("input", tt.fields)
				t.Error(err)
			}
			eof := got[len(got)-1]
			if eof.End != len(tt.fields) {
				t.Errorf("Expected length (%d) to equal EOF.End %d", len(tt.fields), eof.End)
			}
			if eof.Kind != TokenEOF {
				t.Errorf("Expeced the last token to be EOF but was: %s", eof.Kind)
			}
		})
	}
}
