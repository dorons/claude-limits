package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	jsonFlag := flag.Bool("json", false, "output usage data as JSON")
	flag.Parse()

	token, err := readCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credentials: %v\n", err)
		os.Exit(1)
	}

	usage, err := fetchUsage(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching usage: %v\n", err)
		os.Exit(1)
	}

	if *jsonFlag {
		printUsageJSON(usage)
	} else {
		printUsage(usage)
	}
}
