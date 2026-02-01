package main

import (
	"log"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/cmd/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
