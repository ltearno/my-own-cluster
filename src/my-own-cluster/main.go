package main

import (
	"flag"
	"fmt"
	"path/filepath"
)

func main() {
	fmt.Printf("\nwelcome to my-own-cluster\n\n")

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
		fmt.Printf("\nVERBS :\n\n  serve\n        start the web server\n")
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

	default:
		printHelp()
	}
}
