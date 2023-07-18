package main

import (
	"easm_punkmap/services"
	"github.com/jessevdk/go-flags"
)

func main() {
	var scanner services.Scanner
	var _, err = flags.Parse(&scanner)

	if err == nil {
		scanner.Start()
	}
}
