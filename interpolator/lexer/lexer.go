package lexer

type Lexer struct {
	Input          string
	position       int
	readPosition   int
	length         int
	ch             byte // To support utf-8-tokens, we should change to rune. Literals should be fine.
	tokenMap       map[string]TokenKind
	maxTokenLength int
	Tokens         []Token
}

func NewLexer(input string, tokenMap map[string]TokenKind) *Lexer {
	l := Lexer{Input: input, length: len(input), tokenMap: tokenMap}
	for k := range tokenMap {
		if len(k) > l.maxTokenLength {
			l.maxTokenLength = len(k)
		}
	}
	l.readChar()
	return &l
}

func (l *Lexer) NewInput(input string) {
	l.Input = input
	l.position = 0
	l.readPosition = 0
	l.length = len(l.Input)
	l.ch = 0
	l.readChar()
}

func (l *Lexer) readChar() {
	if l.readPosition >= l.length {
		l.ch = 0
	} else {
		l.ch = l.Input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}
func (l *Lexer) peekCharAt(i int) byte {

	if i >= l.length {
		return 0
	} else {
		return l.Input[i]
	}
}
func (l *Lexer) peekChar() byte {
	return l.peekCharAt(l.readPosition)
}

type Token struct {
	Kind    TokenKind
	Literal string
	Start   int
	End     int
}
type TokenKind string

const (
	TokenEOF     TokenKind = "EOF"
	TokenPrefix            = "TokenPrefixInterpolation"
	TokenSuffix            = "TokenSuffixInterpolation"
	TokenLiteral           = "TokenLiteral"
	// Can follow TokenPrefix, or TokenNestingPrefix
	TokenFormatSeperator  = "(Format)/Nesting seperator" // These two Seperators are often, but not always the same token...
	TokenNestingSeperator = "Format/(Nesting) seperator"
	TokenNestingPrefix    = "NestingPrefix"
	TokenNestingSuffix    = "NestingSuffix"
)

func newToken(kind TokenKind, ch string, start, end int) Token {
	return Token{kind, string(ch), start, end}
}

func (l *Lexer) FindAllTokens() []Token {
	lastIndex := 0
	for {
		tok := l.peekToken(l.position)
		if tok != nil {
			if lastIndex != tok.Start {
				literal := newToken(TokenLiteral, l.Input[lastIndex:tok.Start], lastIndex, tok.Start)
				l.Tokens = append(l.Tokens, literal)
			}
			lastIndex = tok.End
			l.Tokens = append(l.Tokens, *tok)
			l.position = tok.End - 1
			l.readPosition = tok.End
			l.readChar()
			if tok.Kind == TokenEOF {
				return l.Tokens
			}
		} else {
			l.readChar()
		}
	}
}
func (l *Lexer) peekToken(index int) *Token {
	var tok Token
	firstCh := l.peekCharAt(index)
	switch firstCh {
	case 0:
		tok = newToken(TokenEOF, "", l.position, l.position)
		return &tok
	}
	// A bit naive, but our tokens are probably not too long
	// TODO: the tokenMap could be a tree, which would reduce lookups, but its a small dataset.
	maxReadAhead := 0
	for k, kind := range l.tokenMap {
		if k[0] != firstCh {
			continue
		}
		for i := 1; i < len(k); i++ {
			if i > maxReadAhead {
				maxReadAhead = i
			}
			ch := l.peekChar()
			if ch != k[i] {
				continue
			}
		}
		startPos := l.position
		// TODO: readAhead, do stuf
		tok = newToken(kind, k, startPos, startPos+len(k))
		return &tok
	}
	return nil

}
