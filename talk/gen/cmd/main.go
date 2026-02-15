// talk-gen generates endpoint registration code from Go interfaces.
//
// Usage:
//
//	//go:generate go run go.zoe.im/x/talk/gen/cmd -type=UserService
//	//go:generate go run go.zoe.im/x/talk/gen/cmd -type=userService -annotations
//
// This will generate a file named <input>_talk.go containing endpoint
// registration code for the specified interface.
//
// With -annotations flag, it generates TalkAnnotations() method from
// source code comments containing @talk directives.
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
		typeName    = flag.String("type", "", "type name to generate for")
		outputFile  = flag.String("output", "", "output file name")
		annotations = flag.Bool("annotations", false, "generate TalkAnnotations() from comments")
	)

	flag.Parse()

	if *typeName == "" {
		fmt.Fprintln(os.Stderr, "error: -type flag is required")
		flag.Usage()
		os.Exit(1)
	}

	sourceFile := os.Getenv("GOFILE")
	if sourceFile == "" && flag.NArg() > 0 {
		sourceFile = flag.Arg(0)
	}
	if sourceFile == "" {
		fmt.Fprintln(os.Stderr, "error: no source file specified (run via go generate or pass file as argument)")
		os.Exit(1)
	}

	if !filepath.IsAbs(sourceFile) {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
			os.Exit(1)
		}
		sourceFile = filepath.Join(dir, sourceFile)
	}

	if *annotations {
		if err := gen.GenerateAnnotations(sourceFile, *typeName, *outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		outputName := *outputFile
		if outputName == "" {
			ext := filepath.Ext(sourceFile)
			outputName = sourceFile[:len(sourceFile)-len(ext)] + "_talk_annotations" + ext
		}
		fmt.Printf("Generated %s\n", outputName)
		return
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
