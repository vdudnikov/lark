package scanner_test

import (
	"fmt"

	"larklang.io/lark/pkg/scanner"
)

func ExampleScanner_Scan() {
	// src is the input that we want to tokenize.
	src := []byte("const foo = 1 + bar")

	// New scanner
	s := scanner.New(src, nil)

	// Repeated calls to Scan yield the token sequence found in the input.
	for !s.Done() {
		token := s.Scan()
		fmt.Printf("%d:%d %s %s\n", token.Pos.Line+1, token.Pos.Column+1, token.Kind, token.Value)
	}

	// output:
	// 1:1 const const
	// 1:7 IDENTIFIER foo
	// 1:11 = =
	// 1:13 INTEGER 1
	// 1:15 + +
	// 1:17 IDENTIFIER bar
	// 1:20 ENDMARKER endmarker
}
