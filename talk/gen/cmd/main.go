// talk-gen generates endpoint registration code from Go interfaces.
//
// Usage:
//
//	//go:generate go run go.zoe.im/x/talk/gen/cmd -type=UserService
//
// This will generate a file named <input>_talk.go containing endpoint
// registration code for the specified interface.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.zoe.im/x/talk/gen"
)

func main() {
	var (
		typeName   = flag.String("type", "", "interface type name to generate endpoints for")
		outputFile = flag.String("output", "", "output file name (default: <input>_talk.go)")
	)

	flag.Parse()

	if *typeName == "" {
		fmt.Fprintln(os.Stderr, "error: -type flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Get source file from GOFILE env var (set by go generate) or args
	sourceFile := os.Getenv("GOFILE")
	if sourceFile == "" && flag.NArg() > 0 {
		sourceFile = flag.Arg(0)
	}
	if sourceFile == "" {
		fmt.Fprintln(os.Stderr, "error: no source file specified (run via go generate or pass file as argument)")
		os.Exit(1)
	}

	// Make source file absolute if needed
	if !filepath.IsAbs(sourceFile) {
		dir := os.Getenv("GOPACKAGE")
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
				os.Exit(1)
			}
		}
		sourceFile = filepath.Join(dir, sourceFile)
	}

	g := &gen.Generator{
		TypeName:   *typeName,
		OutputFile: *outputFile,
	}

	if err := g.Generate(sourceFile); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outputName := *outputFile
	if outputName == "" {
		ext := filepath.Ext(sourceFile)
		outputName = sourceFile[:len(sourceFile)-len(ext)] + "_talk" + ext
	}
	fmt.Printf("Generated %s\n", outputName)
}
