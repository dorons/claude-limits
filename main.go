package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	jsonFlag       := flag.Bool("json", false, "output usage data as JSON")
	statuslineFlag := flag.Bool("statusline", false, "output compact colorized usage for statusline")
	flag.Parse()

	token, err := readCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credentials: %v\n", err)
		os.Exit(1)
	}

	usage, err := fetchUsageCached(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching usage: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *jsonFlag:
		printUsageJSON(usage)
	case *statuslineFlag:
		printStatusline(usage)
	default:
		printUsage(usage)
	}
}
