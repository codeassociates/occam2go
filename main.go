package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codeassociates/occam2go/codegen"
	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
)

const version = "0.1.0"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	outputFile := flag.String("o", "", "Output file (default: stdout)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "occam2go - An Occam to Go transpiler\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input.occ>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("occam2go version %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputFile := args[0]

	// Read input file
	input, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(input))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors:\n")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	// Generate Go code
	gen := codegen.New()
	output := gen.Generate(program)

	// Write output
	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(output)
	}
}
