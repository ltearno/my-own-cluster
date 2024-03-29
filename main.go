package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/ltearno/my-own-cluster/enginejs"
	"github.com/ltearno/my-own-cluster/enginewasm"
	"github.com/ltearno/my-own-cluster/enginewasmer"

	"github.com/ltearno/my-own-cluster/assetsgen"
	"github.com/ltearno/my-own-cluster/common"

	"github.com/ltearno/my-own-cluster/apijwt"

	"github.com/ltearno/my-own-cluster/apicore"
	"github.com/ltearno/my-own-cluster/apigpu"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

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
		trace := verbs[0].GetOptionOr("trace", "false") == "true"
		removeFilters := verbs[0].GetOptionOr("remove-filters", "false") == "true"
		useWasmer := verbs[0].GetOptionOr("wasmer", "false") == "true"
		trace = trace || removeFilters

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

		if trace {
			dumpDB(db)
		}

		{
			// migrate old plug system
			prefix := []byte("/plugs/byspec/")
			iter := db.NewIterator(util.BytesPrefix(prefix), nil)
			for iter.Next() {
				key := string(iter.Key()[len(prefix):])
				value := iter.Value()

				s := strings.Index(key, "/")
				method := key[0:s]
				path := key[s+1:]

				fmt.Printf("migration: %s %s %s\n", method, path, string(value))

				newKey := []byte(fmt.Sprintf("/plug_system/plugs/byspec/%s/%s", method, path))
				hasIt, _ := db.Has(newKey, nil)
				if hasIt {
					fmt.Printf("=> skip, already on new version...\n")
				} else {
					fmt.Printf("=> copy to %s\n", string(newKey))
					db.Put(newKey, value, nil)
				}

				fmt.Printf("=> deleting old key\n")
				db.Delete(iter.Key(), nil)
			}
			iter.Release()
		}

		db.Put([]byte("/database-version"), []byte("1"), nil)

		if removeFilters {
			fmt.Printf("\nremoving all filters because of command line option\n\n")
			db.Delete([]byte("/filters"), nil)
		}

		orchestrator := common.NewOrchestrator(db, trace)

		// register execution engines
		orchestrator.AddExecutionEngine("text/javascript", enginejs.NewJavascriptDuktapeEngine())
		if useWasmer {
			fmt.Printf("use WASMER web assembly engine")
			orchestrator.AddExecutionEngine("application/wasm", enginewasmer.NewWasmWasmerEngine())
		} else {
			fmt.Printf("use WASM3 web assembly engine")
			orchestrator.AddExecutionEngine("application/wasm", enginewasm.NewWasmWasm3Engine())
		}

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
		apiProvider, err = apijwt.NewJWTAPIProvider()
		if err == nil {
			orchestrator.AddAPIProvider("jwt", apiProvider)
		}

		// init core-api
		coreAPILibrary, err := assetsgen.Asset("assets/rest-default-api.js")
		if err == nil {
			systemTags := "{\"category\":\"system-bootstrap\"}"
			orchestrator.RegisterBlobWithName("core-api", "text/javascript", coreAPILibrary)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/blob/register", "core-api", "registerBlob", "", systemTags)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/file/plug", "core-api", "plugFile", "", systemTags)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/function/plug", "core-api", "plugFunction", "", systemTags)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/function/unplug", "core-api", "unplugPath", "", systemTags)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/function/call", "core-api", "callFunction", "", systemTags)
			orchestrator.PlugFunction("POST", "/my-own-cluster/api/filter/plug", "core-api", "plugFilter", "", systemTags)
			orchestrator.PlugFunction("DELETE", "/my-own-cluster/api/filter/plug/!filter-id", "core-api", "unplugFilter", "", systemTags)
			orchestrator.PlugFunction("GET", "/my-own-cluster/api/admin/export-database", "core-api", "exportDatabase", "", systemTags)
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

	case "push":
		CliPushFunction(verbs)

	case "call":
		CliCallFunction(verbs)

	case "export-database":
		CliExportDatabase(verbs)

	case "upload":
		CliUploadFile(verbs)

	case "upload-dir":
		CliUploadDir(verbs)

	case "plug":
		CliPlugFunction(verbs)

	case "unplug":
		CliUnplug(verbs)

	case "plug-filter":
		CliPlugFilter(verbs)

	case "unplug-filter":
		CliUnplugFilter(verbs)

	case "kvm_test":
		TestKVM()

	case "guest-api":
		CliGuestApi(verbs)

	case "version":
		CliVersion(verbs)

	case "remote":
		CliRemote(verbs)

	default:
		fmt.Printf("No argument received !\n")
		printHelp()
	}
}
