package main

import (
	"fmt"
	"os"
	"strings"

	"larklang.io/lark/pkg/ast"
	"larklang.io/lark/pkg/parser"
)

func exit(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		exit("no input file")
	}

	filename := os.Args[1]
	text, err := os.ReadFile(filename)
	if err != nil {
		exit(err.Error())
	}

	parsed := parser.Parse(text)

	if len(parsed.Errors) > 0 {
		for _, err := range parsed.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", filename, err.Pos.Line+1, err.Pos.Column+1, err.Message)
			if err.Pos.Line < len(parsed.Lines) {
				line := parsed.Lines[err.Pos.Line]
				fmt.Fprintf(os.Stderr, "  %s\n", line)
				fmt.Fprint(os.Stderr, strings.Repeat(" ", err.Pos.Column+2)+"^\n")
			}
		}
	} else {
		ast.Print(parsed.File)
	}
}
