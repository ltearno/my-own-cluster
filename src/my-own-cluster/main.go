package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const BASE_SERVER_URL = "https://localhost:8443"

type Verb struct {
	Name    string
	Options map[string]string
}

func ParseArgs(args []string) ([]Verb, error) {
	verbs := []Verb{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			verbs[len(verbs)-1].Options[arg[1:]] = args[i+1]
			i++
		} else {
			verbs = append(verbs, Verb{arg, make(map[string]string)})
		}
	}

	return verbs, nil
}

func main() {
	v, _ := ParseArgs(os.Args)
	fmt.Println(v)

	var printUsage = false
	var help = flag.Bool("help", false, "show this help")
	var port = flag.Int("port", 8443, "webserver listening port")

	flag.Parse()

	if *help {
		printUsage = true
	}

	verbs := flag.Args()

	if len(verbs) == 0 {
		fmt.Println("not enough parameters, use '-help' !")
		printUsage = true
	}

	printHelp := func() {
		fmt.Printf("\nmy-own-cluster usage :\n\n  my-own-cluster [OPTIONS] verbs...\n\nOPTIONS :\n\n")
		flag.PrintDefaults()
		fmt.Printf("\nVERBS :\n\n")
		fmt.Printf("  help\n")
		fmt.Printf("      prints this message\n")
		fmt.Printf("  serve\n")
		fmt.Printf("      start the web server\n")
		fmt.Printf("  push FUNCTION_NAME WASM_FILE\n")
		fmt.Printf("      sends a wasm code to the server\n")
		fmt.Printf("  call FUNCTION_NAME posix")
		fmt.Printf("      calls a function in POSIX mode (through WASI implementation)\n")
		fmt.Printf("  call FUNCTION_NAME direct\n")
		fmt.Printf("      calls a function in direct mode\n")
	}

	if printUsage {
		printHelp()
		return
	}

	// execute the verb
	switch verbs[0] {
	case "serve":
		relativeWorkdir := "."
		if len(verbs) > 1 {
			relativeWorkdir = verbs[1]
		}

		workingDir, err := filepath.Abs(relativeWorkdir)
		if err != nil {
			fmt.Printf("cannot find working directory, abort (%v)\n", err)
			return
		}

		orchestrator := NewOrchestrator()

		StartWebServer(*port, workingDir, orchestrator)
		break

	case "push":
		if len(verbs) != 3 {
			printHelp()
			return
		}

		functionName := verbs[1]
		wasmFileName := verbs[2]

		CliPushFunction(functionName, wasmFileName)
		break

	case "call":
		CliCallFunction(verbs[1:])
		break

	default:
		printHelp()
	}
}
