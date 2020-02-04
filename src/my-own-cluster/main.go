package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const BASE_SERVER_URL = "https://localhost:8443"

type Verb struct {
	Name    string
	Options map[string]string
}

func (v *Verb) GetOptionOr(optionName string, defaultValue string) string {
	value, ok := v.Options[optionName]
	if !ok {
		value = defaultValue
	}
	return value
}

func (v *Verb) HasOption(optionName string) bool {
	_, ok := v.Options[optionName]
	return ok
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

func printHelp() {
	fmt.Printf("\nmy-own-cluster usage :\n\n")
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

func main() {
	verbs, err := ParseArgs(os.Args)
	if err != nil {
		printHelp()
		return
	}

	// remove process file name
	verbs = verbs[1:]

	if len(verbs) == 0 || verbs[0].Name == "help" {
		fmt.Println("not enough parameters, use '-help' !")
		printHelp()
		return
	}

	// execute the verb
	switch verbs[0].Name {
	case "serve":
		relativeWorkdir := "."
		if len(verbs) > 1 {
			relativeWorkdir = verbs[1].Name
		}

		workingDir, err := filepath.Abs(relativeWorkdir)
		if err != nil {
			fmt.Printf("cannot find working directory, abort (%v)\n", err)
			return
		}

		db, err := leveldb.OpenFile(filepath.Join(workingDir, "my-own-cluster-database-provisional"), nil)
		if err != nil {
			fmt.Printf("cannot find open database (%v)\n", err)
			return
		}
		defer db.Close()

		orchestrator := NewOrchestrator(db)

		port := 8443
		if portOption, ok := verbs[0].Options["port"]; ok {
			port, err = strconv.Atoi(portOption)
			if err != nil {
				fmt.Printf("wrong port '%s', should be a number\n", portOption)
				return
			}
		}

		StartWebServer(port, workingDir, orchestrator)
		break

	case "push":
		if len(verbs) != 3 {
			printHelp()
			return
		}

		functionName := verbs[1].Name
		wasmFileName := verbs[2].Name

		CliPushFunction(functionName, wasmFileName)
		break

	case "call":
		CliCallFunction(verbs)
		break

	case "upload":
		CliUploadFile(verbs)
		break

	default:
		fmt.Printf("No argument received !\n")
		printHelp()
	}
}
