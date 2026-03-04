package main

import (
	"fmt"
	"os"
)

func main() {
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

	printUsage(usage)
}
