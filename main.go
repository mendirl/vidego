package main

import (
	"log"
	"vidego/pkg/commands"
)

func main() {
	err := commands.Execute()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
