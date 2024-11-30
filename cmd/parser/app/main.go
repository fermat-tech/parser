package app

import (
	"flag"
	"fmt"
	"github.com/fermat-tech/parser/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func usageAndExit1() {
	programName := filepath.Base(os.Args[0])
	posArg1 := "<INPUT_FILE>"
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "usage: %v [--help] %s\n", programName, posArg1)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func parse(ts *token.TokenStream, closer func()) error {
	defer closer()
	tok := token.Token{}
	err := error(nil)
	for {
		if tok, err = ts.Get(); err != nil {
			if err != io.EOF {
				if len(tok.Value) > 0 { // Check for token value showing error context
					tok.Value = strings.TrimRight(tok.Value, "\r\n")
					fmt.Fprintf(os.Stderr, "ERROR: %v at: %v\n", err, tok.Value)
				} else {
					fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				}
			}
			return err
		} else {
			switch tok.Type {
			case token.DQSTRING:
				fmt.Printf("Token: DQSTRING: %v\n", tok.Value[1:len(tok.Value)-1])
			case token.SQSTRING:
				fmt.Printf("Token: SQSTRING: %v\n", tok.Value[1:len(tok.Value)-1])
			case token.NUMBER:
				fmt.Printf("Token: NUMBER: %v\n", tok.Value)
			case token.PUNCT:
				fmt.Printf("Token: PUNCT: %v\n", tok.Value)
			case token.DOT:
				fmt.Printf("Token: DOT: %v\n", tok.Value)
			case token.OPAREN:
				fmt.Printf("Token: OPAREN: %v\n", tok.Value)
			case token.CPAREN:
				fmt.Printf("Token: CPAREN: %v\n", tok.Value)
			case token.OBRACKET:
				fmt.Printf("Token: OBRACKET: %v\n", tok.Value)
			case token.CBRACKET:
				fmt.Printf("Token: CBRACKET: %v\n", tok.Value)
			case token.OBRACE:
				fmt.Printf("Token: OBRACE: %v\n", tok.Value)
			case token.CBRACE:
				fmt.Printf("Token: CBRACE: %v\n", tok.Value)
			case token.COMMA:
				fmt.Printf("Token: COMMA: %v\n", tok.Value)
			case token.SEMI:
				fmt.Printf("Token: SEMI: %v\n", tok.Value)
			case token.COLON:
				fmt.Printf("Token: COLON: %v\n", tok.Value)
			case token.CARET:
				fmt.Printf("Token: CARET: %v\n", tok.Value)
			case token.PLUS:
				fmt.Printf("Token: PLUS: %v\n", tok.Value)
			case token.MINUS:
				fmt.Printf("Token: MINUS: %v\n", tok.Value)
			case token.MULT:
				fmt.Printf("Token: MULT: %v\n", tok.Value)
			case token.DIV:
				fmt.Printf("Token: DIV: %v\n", tok.Value)
			case token.COMMAND:
				if tok.Value == ".exit" || tok.Value == ".quit" {
					fmt.Fprintf(os.Stderr, "NOTICE: Exiting the program due to token %q\n", tok.Value)
					return nil
				}
				fmt.Fprintf(os.Stderr, "NOTICE: Invalid COMMAND: %v\n", tok.Value)
			case token.IDENTIFIER:
				fmt.Printf("Token: IDENTIFIER: %v\n", tok.Value)
			default:
				fmt.Printf("Token: INVALID: %v\n", tok.Value)
			}
		}
	}
}

func realMain() int {
	errors := false
	fs := flag.NewFlagSet("main", flag.ExitOnError)
	fs.Usage = usageAndExit1
	fs.Parse(os.Args[1:])
	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", "Missing <INPUT_FILE>")
		usageAndExit1()
	}
	for _, file := range fs.Args() {
		if file == "-" {
			file = "<stdin>"
			fmt.Printf("%v\n", file)
			continue
		}
		ts, closer, err := token.NewFileTokenStream(file)
		if err != nil {
			errors = true
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			continue
		}
		if err := parse(ts, closer); err != nil && err != io.EOF {
			errors = true
		}
	}
	if errors {
		return 2
	}
	return 0
}

func Main() {
	os.Exit(realMain())
}
