package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mathwro/AzCliPermissions/cmd"
)

func main() {
	// Parse command line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		versionShort = flag.Bool("v", false, "Show version information (short)")
		helpShort    = flag.Bool("h", false, "Show help information (short)")
		debugMode    = flag.Bool("debug", false, "Enable debug mode with verbose output")
		debugShort   = flag.Bool("d", false, "Enable debug mode with verbose output (short)")
	)
	
	flag.Parse()

	// Create CLI instance (always uses live Azure API querying)
	cli := cmd.NewCLI()

	// Set debug mode if flag is provided
	if *debugMode || *debugShort {
		cli.SetDebugMode(true)
	}

	// Handle version flag
	if *showVersion || *versionShort {
		fmt.Printf("Azure CLI Permissions Analyzer (azperm) v%s\n", cli.Version())
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp || *helpShort {
		cli.Help()
		os.Exit(0)
	}

	// Run the main CLI logic (always uses live Azure API)
	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
