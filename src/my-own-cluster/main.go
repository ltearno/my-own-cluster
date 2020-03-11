package main

import (
	"fmt"
	"my-own-cluster/apicore"
	"my-own-cluster/apigpu"
	"my-own-cluster/assetsgen"
	"my-own-cluster/common"
	"my-own-cluster/enginejs"
	"my-own-cluster/enginewasm"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/syndtr/goleveldb/leveldb"
)

/**

TODO

BE CAREFUL, SOON WE WILL REPLACE NAMES :

- FUNCTION BY MODULE
- START_FUNCTION BY FUNCTION

**/

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

func dumpDB(db *leveldb.DB) {
	fmt.Println("=database_dump=>")
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		fmt.Printf("%s\n", string(iter.Key()))
	}
	iter.Release()
	fmt.Println("<=database_dump_finished=")
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
		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 2)

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

		orchestrator := common.NewOrchestrator(db)

		// register execution engines
		orchestrator.AddExecutionEngine("text/javascript", enginejs.NewJavascriptDuktapeEngine())
		orchestrator.AddExecutionEngine("application/wasm", enginewasm.NewWasmWasm3Engine())

		// add api providers
		apiProvider, err := apicore.NewCoreAPIProvider()
		if err == nil {
			orchestrator.AddAPIProvider("core", apiProvider)
			orchestrator.AddAPIProvider("my-own-cluster", apiProvider)
		}
		apiProvider, err = apigpu.NewGPUAPIProvider()
		if err == nil {
			orchestrator.AddAPIProvider("gpu", apiProvider)
		}

		// init core-api
		coreAPILibrary, err := assetsgen.Asset("assets/rest-default-api.js")
		if err == nil {
			orchestrator.RegisterBlobWithName("core-api", "text/javascript", coreAPILibrary)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/blob/register", "core-api", "registerBlob")
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/file/plug", "core-api", "plugFile")
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/function/plug", "core-api", "plugFunction")
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/function/call", "core-api", "callFunction")
		} else {
			fmt.Printf("[error] cannot load rest-default-api.js, things may go bad quickly...\n")
		}

		port := 8443
		if portOption, ok := verbs[0].Options["port"]; ok {
			port, err = strconv.Atoi(portOption)
			if err != nil {
				fmt.Printf("wrong port '%s', should be a number\n", portOption)
				return
			}
		}

		trace := verbs[0].GetOptionOr("trace", "false") == "true"

		if trace {
			dumpDB(db)
		}

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			StartWebServer(port, workingDir, orchestrator, trace)
			fmt.Printf("\nweb server terminated abruptly, exiting\n")
			done <- true
		}()

		go func() {
			sig := <-sigs
			fmt.Printf("\nreceived signal %v, exiting\n", sig)
			done <- true
		}()

		<-done
		fmt.Printf("bye\n")

		break

	case "push":
		CliPushFunction(verbs)
		break

	case "call":
		CliCallFunction(verbs)
		break

	case "upload":
		CliUploadFile(verbs)
		break

	case "upload-dir":
		CliUploadDir(verbs)
		break

	case "plug":
		CliPlugFunction(verbs)
		break

	case "kvm_test":
		TestKVM()
		break

	case "guest-api":
		CliGuestApi(verbs)
		break

	case "version":
		CliVersion(verbs)
		break

	default:
		fmt.Printf("No argument received !\n")
		printHelp()
	}
}
