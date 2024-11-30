package token

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type Type int

const (
	INVALID Type = iota
	IDENTIFIER
	NUMBER
	DQSTRING
	SQSTRING
	PUNCT
	COMMAND
	DOT
	OPAREN
	CPAREN
	OBRACKET
	CBRACKET
	OBRACE
	CBRACE
	COMMA
	SEMI
	COLON
	CARET
	PLUS
	MINUS
	MULT
	DIV
)

type charStream struct {
	reader       *bufio.Reader
	lookaheadBuf []rune
}

func newCharStream(r *bufio.Reader) *charStream {
	return &charStream{
		reader:       r,
		lookaheadBuf: []rune{},
	}
}

func (c *charStream) peek() (rune, error) {
	if len(c.lookaheadBuf) > 0 {
		return c.lookaheadBuf[0], nil
	}
	chr, _, err := c.reader.ReadRune()
	if err != nil {
		return rune(0), err
	}
	c.lookaheadBuf = append([]rune{chr}, c.lookaheadBuf...)
	return chr, nil
}

func (c *charStream) get() (rune, error) {
	if len(c.lookaheadBuf) > 0 {
		chr := c.lookaheadBuf[0]
		c.lookaheadBuf = c.lookaheadBuf[1:]
		return chr, nil
	}
	chr, _, err := c.reader.ReadRune()
	if err != nil && err != io.EOF {
		return rune(0), err
	}
	return chr, nil
}

type Token struct {
	Value string
	Type  Type
}

type TokenStream struct {
	fileName     string
	chrStream    *charStream
	lookaheadBuf []string
}

func NewFileTokenStream(file string) (*TokenStream, func(), error) {
	ts := TokenStream{fileName: file}
	reader, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	bioReader := bufio.NewReader(reader)
	ts.chrStream = newCharStream(bioReader)
	return &ts,
		func() {
			//fmt.Fprintf(os.Stderr, "DEBUG: Closing %s\n", file)
			reader.Close()
		},
		nil
}

func (t *TokenStream) String() string {
	return fmt.Sprintf("TokenStream: fileName: %s", t.fileName)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v' || r == '\f' || r == '\r' || r == '\n'
}

func (t *TokenStream) Get() (Token, error) {
	chr := rune(0)
	err := error(nil)
	token := strings.Builder{}

	// Skip whitespace
	for {
		chr, err = t.chrStream.peek()
		if err != nil {
			return Token{Value: "", Type: INVALID}, err
		}
		if unicode.IsSpace(chr) {
			t.chrStream.get()
			continue
		}
		break
	}

	// Get IDENTIFIER
	if unicode.IsLetter(chr) || chr == '_' {
		t.chrStream.get()
		token.WriteRune(chr)
		for {
			chr, err = t.chrStream.peek()
			if err != nil {
				if err != io.EOF {
					return Token{Value: "", Type: INVALID}, err
				}
				break
			}
			if chr == rune(0) {
				break
			}
			if unicode.IsLetter(chr) || unicode.IsDigit(chr) || chr == '_' {
				t.chrStream.get()
				token.WriteRune(chr)
				continue
			}
			break
		}
		return Token{Value: token.String(), Type: IDENTIFIER}, nil
	}

	if chr == '.' {
		t.chrStream.get()
		token.WriteRune(chr)
		return Token{Value: token.String(), Type: DOT}, nil
	}

	// Get COMMAND
	//	if chr == '.' {
	//		t.chrStream.get()
	//		token.WriteRune(chr)
	//		for {
	//			chr, err = t.chrStream.peek()
	//			if err != nil {
	//				return Token{Value: "", Type: INVALID}, err
	//			}
	//			if chr == rune(0) {
	//				break
	//			}
	//			if unicode.IsLetter(chr) || unicode.IsDigit(chr) || chr == '_' {
	//				t.chrStream.get()
	//				token.WriteRune(chr)
	//				continue
	//			}
	//			break
	//		}
	//		return Token{Value: token.String(), Type: COMMAND}, nil
	//	}

	// Get DQSTRING
	if chr == '"' {
		t.chrStream.get()
		token.WriteRune(chr)
		inEscape := false
		inString := true
		for {
			chr, err = t.chrStream.peek()
			if err != nil {
				if err != io.EOF {
					return Token{Value: "", Type: INVALID}, err
				}
				break
			}
			if chr == rune(0) {
				break
			}
			if !inEscape && chr == '\\' {
				t.chrStream.get()
				inEscape = true
				continue
			}
			if inEscape {
				inEscape = false
				t.chrStream.get()
				if chr == 'n' {
					token.WriteRune('\n')
				} else if chr == 't' {
					token.WriteRune('\t')
				} else {
					token.WriteRune(chr)
				}
				continue
			}
			if chr == '"' {
				inString = false
				t.chrStream.get()
				token.WriteRune(chr)
				break
			}
			t.chrStream.get()
			token.WriteRune(chr)
		}
		if inString {
			return Token{Value: token.String(), Type: INVALID}, errors.New("missing closing double quote")
		}
		return Token{Value: token.String(), Type: DQSTRING}, nil
	}

	// Get SQSTRING
	if chr == '\'' {
		t.chrStream.get()
		token.WriteRune(chr)
		inEscape := false
		inString := true
		for {
			chr, err = t.chrStream.peek()
			if err != nil {
				if err != io.EOF {
					return Token{Value: "", Type: INVALID}, err
				}
				break
			}
			if chr == rune(0) {
				break
			}
			if !inEscape && chr == '\\' {
				t.chrStream.get()
				inEscape = true
				continue
			}
			if inEscape {
				inEscape = false
				t.chrStream.get()
				if !(chr == '\'' || chr == '\\') {
					token.WriteRune('\\')
				}
				token.WriteRune(chr)
				continue
			}
			if chr == '\'' {
				inString = false
				t.chrStream.get()
				token.WriteRune(chr)
				break
			}
			t.chrStream.get()
			token.WriteRune(chr)
		}
		if inString {
			return Token{Value: token.String(), Type: INVALID}, errors.New("missing closing single quote")
		}
		return Token{Value: token.String(), Type: SQSTRING}, nil
	}

	// Get NUMBER
	if unicode.IsDigit(chr) {
		t.chrStream.get()
		token.WriteRune(chr)
		for {
			chr, err = t.chrStream.peek()
			if err != nil {
				if err != io.EOF {
					return Token{Value: "", Type: INVALID}, err
				}
				break
			}
			if chr == rune(0) {
				break
			}
			if !unicode.IsDigit(chr) {
				break
			}
			t.chrStream.get()
			token.WriteRune(chr)
		}
		return Token{Value: token.String(), Type: NUMBER}, nil
	}

	if err != nil {
		return Token{Value: "", Type: INVALID}, err
	}
	if chr == rune(0) {
		return Token{Value: "", Type: INVALID}, nil
	}
	// Return individual character as a string
	t.chrStream.get()
	switch chr {
	case '(':
		return Token{Value: string(chr), Type: OPAREN}, nil
	case ')':
		return Token{Value: string(chr), Type: CPAREN}, nil
	case '[':
		return Token{Value: string(chr), Type: OBRACKET}, nil
	case ']':
		return Token{Value: string(chr), Type: CBRACKET}, nil
	case '{':
		return Token{Value: string(chr), Type: OBRACE}, nil
	case '}':
		return Token{Value: string(chr), Type: CBRACE}, nil
	case ',':
		return Token{Value: string(chr), Type: COMMA}, nil
	case ';':
		return Token{Value: string(chr), Type: SEMI}, nil
	case ':':
		return Token{Value: string(chr), Type: COLON}, nil
	case '^':
		return Token{Value: string(chr), Type: CARET}, nil
	case '+':
		return Token{Value: string(chr), Type: PLUS}, nil
	case '-':
		return Token{Value: string(chr), Type: MINUS}, nil
	case '*':
		return Token{Value: string(chr), Type: MULT}, nil
	case '/':
		return Token{Value: string(chr), Type: DIV}, nil
	}
	return Token{Value: string(chr), Type: PUNCT}, nil
}
