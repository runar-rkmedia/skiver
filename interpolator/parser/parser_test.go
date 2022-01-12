package parser

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/interpolator/lexer"
)

func TestParser(t *testing.T) {
	defaultMap := map[string]lexer.TokenKind{
		"{{":  lexer.TokenPrefix,
		"}}":  lexer.TokenSuffix,
		",":   lexer.TokenFormatSeperator,
		"$t(": lexer.TokenNestingPrefix,
		")":   lexer.TokenNestingSuffix,
	}
	// maxLoops := 10
	tests := []struct {
		name    string
		fields  string
		expects Ast
		wantErr bool
	}{

		{
			"Single interpolation",
			"foo bar {{count}} baz",
			Ast{
				Nodes: []Node{
					{
						Token: lexer.Token{Start: 0, End: 8, Kind: lexer.TokenLiteral, Literal: "foo bar "},
					},
					{
						Token: lexer.Token{Start: 8, End: 17, Kind: lexer.TokenPrefix, Literal: "{{"},
						Left: &Node{
							Token: lexer.Token{Start: 10, End: 15, Kind: lexer.TokenLiteral, Literal: "count"},
						},
					},
					{
						Token: lexer.Token{Start: 17, End: 21, Kind: lexer.TokenLiteral, Literal: " baz"},
					},
				},
			},
			false,
		},
		{
			"Single nested",
			"foo bar $t(abc.asd) baz",
			Ast{
				Nodes: []Node{
					{
						Token: lexer.Token{Start: 0, End: 8, Kind: lexer.TokenLiteral, Literal: "foo bar "},
					},
					{
						Token: lexer.Token{Start: 8, End: 19, Kind: lexer.TokenNestingPrefix, Literal: "$t("},
						Left: &Node{
							Token: lexer.Token{Start: 11, End: 18, Kind: lexer.TokenLiteral, Literal: "abc.asd"},
						},
					},
					{
						Token: lexer.Token{Start: 19, End: 23, Kind: lexer.TokenLiteral, Literal: " baz"},
					},
				},
			},
			false,
		},
		{
			"Single nested",
			"foo bar {{ colour }}$t(abc.asd) baz",
			Ast{
				Nodes: []Node{
					{
						Token: lexer.Token{Start: 0, End: 8, Kind: lexer.TokenLiteral, Literal: "foo bar "},
					},
					{
						Token: lexer.Token{Start: 8, End: 20, Kind: lexer.TokenPrefix, Literal: "{{"},
						Left: &Node{
							Token: lexer.Token{Start: 10, End: 18, Kind: lexer.TokenLiteral, Literal: " colour "},
						},
					},
					{
						Token: lexer.Token{Start: 20, End: 31, Kind: lexer.TokenNestingPrefix, Literal: "$t("},
						Left: &Node{
							Token: lexer.Token{Start: 23, End: 30, Kind: lexer.TokenLiteral, Literal: "abc.asd"},
						},
					},
					{
						Token: lexer.Token{Start: 31, End: 35, Kind: lexer.TokenLiteral, Literal: " baz"},
					},
				},
			},
			false,
		},

		// {
		// 	"Multiple nestings, interpolations, with eobjects",
		// 	`They have $t(girls, {\"count\": {{girls}} }) and $t(boys, {\"count\": {{boys}} })`,
		// 	Ast{
		// 		Nodes: []Node{
		// 			{
		// 				Token: lexer.Token{Start: 0, End: 8, Kind: lexer.TokenLiteral, Literal: "foo bar "},
		// 			},
		// 			{
		// 				Token: lexer.Token{Start: 8, End: 20, Kind: lexer.TokenPrefix, Literal: "{{"},
		// 				Left: &Node{
		// 					Token: lexer.Token{Start: 10, End: 18, Kind: lexer.TokenLiteral, Literal: " colour "},
		// 				},
		// 			},
		// 			{
		// 				Token: lexer.Token{Start: 20, End: 31, Kind: lexer.TokenNestingPrefix, Literal: "$t("},
		// 				Left: &Node{
		// 					Token: lexer.Token{Start: 23, End: 30, Kind: lexer.TokenLiteral, Literal: "abc.asd"},
		// 				},
		// 			},
		// 			{
		// 				Token: lexer.Token{Start: 31, End: 35, Kind: lexer.TokenLiteral, Literal: " baz"},
		// 			},
		// 		},
		// 	},
		// 	false,
		// },
		// TODO: add more tests
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewParser(defaultMap)
			got, err := l.Parse(tt.fields)

			if !tt.wantErr {
				testza.AssertNoError(t, err)
			} else if err == nil {
				t.Error("expected error, but none was returned")
			}

			if err := internal.Compare("result", got, tt.expects, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				// Yaml:    false,
				JSON: true,
				// TOML:    true,
			}); err != nil {
				t.Log("input", tt.fields)
				for _, v := range l.l.Tokens {
					t.Logf("%02d-%02d: '%s'%s\t'%s'", v.Start, v.End, v.Literal, strings.Repeat(" ", 32-(len(v.Literal)%32)), v.Kind)
				}
				// y, _ := yaml.Marshal(l.l.Tokens)
				// t.Log(string(y))
				t.Error(err)
			}
		})
	}
}

// var t = `
// Nodes:
//   - Kind: Literal
// 	  Value: foo bar
// 	- Kind: Interpolation
// 	  Nodes:
// 		  - Kind: Literal
// 			  Value: count
// 			- Kind: NestingOption
//   - Kind: Literal
// 	  Value: " "
//   - Kind: Nesting
// 	  Nodes:
// 		  - Kind: Literal
// 			  Value: gib
//   - Kind: Literal
// 	  Value: " baz"
// `
