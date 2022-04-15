package parser

import (
	"encoding/json"
	"fmt"

	"github.com/runar-rkmedia/skiver/interpolator/lexer"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	curIndex  int
}

type Node struct {
	Token lexer.Token
	Left  *Node `json:",omitempty"`
	Right *Node `json:",omitempty"`
}

type NodeError struct {
	Parent *NodeError
	Node
	Message string
}

func (ne NodeError) Error() string {
	if ne.Parent != nil {
		b, _ := json.Marshal(ne.Node)
		return fmt.Errorf("%w Child: %s %s at position %d-%d, %s",
			ne.Parent,
			ne.Message,
			ne.Node.Token.Kind,
			ne.Node.Token.Start,
			ne.Node.Token.End,
			string(b),
		).Error()

	}
	b, _ := json.Marshal(ne.Node)
	return fmt.Sprintf("%s NodeError: %s at position %d-%d, %s", ne.Token.Kind, ne.Message, ne.Token.Start, ne.Token.End, string(b))
}

func (node Node) err(message string) error {
	return NodeError{nil, node, message}
}
func (node Node) parentErr(parent NodeError, message string) error {
	return NodeError{&parent, node, message}
}

type NodeKind string

func NewParser(tokenMap map[string]lexer.TokenKind) *Parser {
	l := lexer.NewLexer("", tokenMap)
	return &Parser{l: l}
}

func (p *Parser) Parse(s string) (Ast, error) {
	p.l.NewInput(s)
	p.l.Tokens = []lexer.Token{}
	p.l.FindAllTokens()
	p.curIndex = 0
	p.curToken = lexer.Token{}
	p.peekToken = lexer.Token{}
	// Set curToekn to the first token, and peakToken to the second one
	// We are now ready to read these.
	p.nextToken()
	p.nextToken()
	return p.parse()
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	if p.curIndex == len(p.l.Tokens) {
		return
	}
	p.peekToken = p.l.Tokens[p.curIndex]
	p.curIndex++
}

type Ast struct {
	Nodes []Node
}

func stringRangeWithInjectedChar(s string, start, mid, end int, injection string) string {
	if start < 0 {
		start = 0
	}
	if end >= len(s) {
		end = len(s) - 1
	}
	before := s[start:mid]
	after := s[mid:end]
	return before + injection + after
}

func (p *Parser) parse() (ast Ast, err error) {
	ast = Ast{[]Node{}}

	i := 0
	for p.curToken.Kind != lexer.TokenEOF {
		if i > 100 {
			return ast, Node{Token: p.curToken}.err("Expected EOF, gave up")
		}
		node := Node{Token: p.curToken}
		switch p.curToken.Kind {
		case lexer.TokenLiteral:
			ast.Nodes = append(ast.Nodes, node)
		case lexer.TokenPrefix:
			node, err = p.parseInterpolation(node)
			if err != nil {
				return ast, err
			}
			ast.Nodes = append(ast.Nodes, node)
		case lexer.TokenNestingPrefix:
			node, err = p.parseNesting(node)
			if err != nil {
				return ast, err
			}
			ast.Nodes = append(ast.Nodes, node)
		default:
			if len(ast.Nodes) == 0 {
				return ast, node.err("unexpected token-kind")
			}
			last := ast.Nodes[len(ast.Nodes)-1]
			b := stringRangeWithInjectedChar(p.l.Input, node.Token.Start-10, node.Token.Start, node.Token.End+10, " ¦ ")
			return ast, node.err(
				fmt.Sprintf(
					"'%[1]s' unexpected token-kind %[2]s '%[3]s' following %[4]s '%[5]s'. ",
					b,                  // 1
					node.Token.Kind,    // 2
					node.Token.Literal, // 3
					last.Token.Kind,    // 4
					last.Token.Literal, // 5
				))
		}

		i++
		p.nextToken()
	}

	return ast, nil

}

func (p *Parser) parseInterpolation(node Node) (Node, error) {

	node.Left = &Node{Token: p.peekToken}
	if node.Left.Token.Kind != lexer.TokenLiteral {
		nerr := NodeError{nil, node, fmt.Sprintf("Expected first node to be a %s", p.printToken(lexer.TokenLiteral))}
		return node, node.Left.parentErr(nerr, "expected TokenLiteral")
	}
	p.nextToken()
	switch p.peekToken.Kind {
	case lexer.TokenFormatSeperator:
		node.Right = &Node{Token: p.peekToken}
		p.nextToken()
		if p.peekToken.Kind != lexer.TokenLiteral {
			nerr := NodeError{nil, node, fmt.Sprintf("Expected option to be a %s", p.printToken(lexer.TokenLiteral))}
			return node, node.Left.parentErr(nerr, "expected TokenLiteral")
		}
		p.nextToken()
	}
	switch p.peekToken.Kind {
	case lexer.TokenSuffix:
		node.Token.End = p.peekToken.End
		p.nextToken()
	}
	return node, nil
}

func (p *Parser) errTokenCannotFollow(node Node) NodeError {
	return NodeError{nil, node, fmt.Sprintf("%s cannot follow %s", p.printToken(p.peekToken.Kind), p.printToken(p.curToken.Kind))}
}
func (p *Parser) parseNesting(node Node) (Node, error) {
	node.Left = &Node{Token: p.peekToken}
	if node.Left.Token.Kind != lexer.TokenLiteral {
		nerr := NodeError{nil, node, fmt.Sprintf("Expected first node to be a %s", p.printToken(lexer.TokenLiteral))}
		return node, node.Left.parentErr(nerr, fmt.Sprintf("expected %s", p.printToken(lexer.TokenLiteral)))
	}
	p.nextToken()
	if p.peekToken.Kind == lexer.TokenNestingSeperator || p.peekToken.Kind == lexer.TokenFormatSeperator {
		p.nextToken()
		switch p.peekToken.Kind {
		// TODO: we must change the tokemMap to a tokenSlice and allow identical identfiers, then we must check if there are identtical tokens
		case lexer.TokenNestingSeperator, lexer.TokenFormatSeperator:
			nerr := p.errTokenCannotFollow(node)
			return node, node.Left.parentErr(nerr, fmt.Sprintf("expected %s", p.printToken(lexer.TokenLiteral)))
		case lexer.TokenNestingSuffix:
			nerr := p.errTokenCannotFollow(node)
			return node, node.Left.parentErr(nerr, fmt.Sprintf("expected %s", p.printToken(lexer.TokenLiteral)))
		default:
			var args Node
			args.Token.Kind = lexer.TokenArgument
			args.Token.Start = p.peekToken.Start
		tokenLoop:
			for {

				switch p.peekToken.Kind {
				case lexer.TokenNestingSuffix:
					// handled below
					break tokenLoop
				case lexer.TokenEOF:
					nerr := NodeError{nil, args, fmt.Sprintf("Expected %s", p.printToken(lexer.TokenNestingSuffix))}
					return node, node.Left.parentErr(nerr, fmt.Sprintf("Unexpected %s", p.printToken(lexer.TokenEOF)))
				default:
					// swallow all tokens into a single token
					args.Token.Literal += p.peekToken.Literal
					args.Token.End = p.peekToken.End
					p.nextToken()
				}
				if args.Token.Literal != "" {
					node.Right = &args
				}
			}

		}
	}
	switch p.peekToken.Kind {
	case lexer.TokenNestingSuffix:
		node.Token.End = p.peekToken.End
		p.nextToken()
	}
	return node, nil
}

func (p *Parser) printToken(token lexer.TokenKind) string {
	literal := p.l.TokenMapLookup(token)
	if literal == "" {
		return fmt.Sprintf("'%s'", token)
	}
	return fmt.Sprintf("'%s' (%s)", token, p.l.TokenMapLookup(token))
}
