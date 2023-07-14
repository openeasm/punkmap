package main

import (
	"easm_punkmap/services"
	"github.com/jessevdk/go-flags"
)

func main() {
	var scanner services.Scanner
	_, _ = flags.Parse(&scanner)
	scanner.Start()
}
