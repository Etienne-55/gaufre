package main

import (
	"flag"
	"fmt"
	"os"

	"gaufre/internal/app"
)

func main() {
	popup := flag.Bool("popup", false, "run inside a tmux popup: drop Gaufre's own frame for the popup's")
	flag.Parse()

	if err := app.Run(*popup); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
