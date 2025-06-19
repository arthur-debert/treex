package main

import (
	"fmt"
	"log"
	"os"

	"my-project/internal/server"
	"my-project/internal/parser"
)

func main() {
	fmt.Println("Starting myapp...")
	
	// Parse command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: myapp <command>")
	}
	
	command := os.Args[1]
	
	switch command {
	case "server":
		server.Start()
	case "parse":
		if len(os.Args) < 3 {
			log.Fatal("Usage: myapp parse <file>")
		}
		filename := os.Args[2]
		parser.ParseFile(filename)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
} 