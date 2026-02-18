package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/codeassociates/occam2go/codegen"
	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/modgen"
	"github.com/codeassociates/occam2go/parser"
	"github.com/codeassociates/occam2go/preproc"
)

const version = "0.1.0"

// multiFlag allows a flag to be specified multiple times (e.g. -I path1 -I path2).
type multiFlag []string

func (f *multiFlag) String() string { return strings.Join(*f, ", ") }
func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	// Check for subcommand before parsing flags
	if len(os.Args) >= 2 && os.Args[1] == "gen-module" {
		genModuleCmd(os.Args[2:])
		return
	}

	showVersion := flag.Bool("version", false, "Print version and exit")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	var includePaths multiFlag
	flag.Var(&includePaths, "I", "Include search path (repeatable)")
	var defines multiFlag
	flag.Var(&defines, "D", "Predefined symbol (repeatable)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "occam2go - An Occam to Go transpiler\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input.occ>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s gen-module [-o output] <SConscript>\n\n", os.Args[0])
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

	// Build defines map
	defs := map[string]string{}
	for _, d := range defines {
		if idx := strings.Index(d, "="); idx >= 0 {
			defs[d[:idx]] = d[idx+1:]
		} else {
			defs[d] = ""
		}
	}

	// Preprocess
	pp := preproc.New(
		preproc.WithIncludePaths(includePaths),
		preproc.WithDefines(defs),
	)
	expanded, err := pp.ProcessFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Preprocessor error: %s\n", err)
		os.Exit(1)
	}
	if len(pp.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Preprocessor warnings:\n")
		for _, e := range pp.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
	}

	// Lex
	l := lexer.New(expanded)

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

func genModuleCmd(args []string) {
	fs := flag.NewFlagSet("gen-module", flag.ExitOnError)
	outputFile := fs.String("o", "", "Output file (default: stdout)")
	moduleName := fs.String("name", "", "Module guard name (default: derived from library name)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: occam2go gen-module [-o output] [-name GUARD] <SConscript>\n")
		os.Exit(1)
	}

	sconscriptFile := fs.Arg(0)
	data, err := os.ReadFile(sconscriptFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SConscript: %s\n", err)
		os.Exit(1)
	}

	libs := modgen.ParseSConscript(string(data))
	if len(libs) == 0 {
		fmt.Fprintf(os.Stderr, "No OccamLibrary found in %s\n", sconscriptFile)
		os.Exit(1)
	}

	// Use first library by default
	lib := libs[0]

	// Derive module name from library name if not specified
	guard := *moduleName
	if guard == "" {
		// course.lib → COURSE.MODULE
		name := lib.Name
		if idx := strings.LastIndex(name, "."); idx >= 0 {
			name = name[:idx]
		}
		guard = strings.ToUpper(name) + ".MODULE"
	}

	output := modgen.GenerateModule(lib, guard)

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
