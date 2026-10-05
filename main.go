package main

import (
	"fmt"
	"os"
	"strings"
)

const usage = `toylang — a toy programming language compiler + VM

usage:
  toylang run <file.toy>          compile and run a program
  toylang run -e "<source>"       run source from the command line
  toylang parse <file.toy>        print the AST (debug)
  toylang compile <file.toy>      print the bytecode (debug)
  toylang version                 print version
  toylang help                    show this help
`

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "run":
		handle(cmdRun(args))
	case "parse":
		handle(cmdParse(args))
	case "compile":
		handle(cmdCompile(args))
	case "version":
		fmt.Println(version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fatal(fmt.Errorf("unknown command: %s", cmd))
	}
}

func handle(err error) {
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "toylang: %v\n", err)
	os.Exit(1)
}

func readSource(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no source provided")
	}
	if args[0] == "-e" {
		if len(args) < 2 {
			return "", fmt.Errorf("-e requires source code")
		}
		return args[1], nil
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseSource(source string) (*Program, error) {
	lex := NewLexer(source)
	tokens, err := lex.Tokenize()
	if err != nil {
		return nil, err
	}
	parser := NewParser(tokens)
	return parser.ParseProgram()
}

func cmdParse(args []string) error {
	source, err := readSource(args)
	if err != nil {
		return err
	}
	program, err := parseSource(source)
	if err != nil {
		return err
	}
	fmt.Println(FormatProgram(program))
	return nil
}

func cmdCompile(args []string) error {
	source, err := readSource(args)
	if err != nil {
		return err
	}
	program, err := parseSource(source)
	if err != nil {
		return err
	}
	compiler := NewCompiler()
	chunk, _, err := compiler.CompileProgram(program)
	if err != nil {
		return err
	}
	fmt.Print(chunk.Disassemble())
	return nil
}

func cmdRun(args []string) (err error) {
	// Recover from panics (e.g. stack overflow from infinite recursion)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	source, err := readSource(args)
	if err != nil {
		return err
	}
	program, err := parseSource(source)
	if err != nil {
		return err
	}
	compiler := NewCompiler()
	chunk, functions, err := compiler.CompileProgram(program)
	if err != nil {
		return err
	}
	vm := NewVM()
	if err := vm.Run(chunk, functions); err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}
	return nil
}

var _ = strings.TrimSpace