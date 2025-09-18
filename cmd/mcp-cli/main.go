package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	scan "github.com/your-org/sast-sca-mcp/internal/scan"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]
	switch cmd {
	case "semgrep":
		runSemgrepCLI(os.Args[2:])
	case "grype":
		runGrypeCLI(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
	}
}

func runSemgrepCLI(args []string) {
	fs := flag.NewFlagSet("semgrep", flag.ExitOnError)
	target := fs.String("target", "", "path to the project directory")
	config := fs.String("config", "auto", "Semgrep configuration (rule set URI or file)")
	timeoutFlag := fs.Duration("timeout", 10*time.Minute, "timeout (e.g. 5m, 300s). Use 0 for no timeout")
	fs.Parse(args)

	ensureTarget(fs.Name(), target)

	resolved, err := scan.ResolveDirectory(*target)
	if err != nil {
		exitErr(err)
	}

	output, err := scan.RunSemgrep(context.Background(), resolved, *config, *timeoutFlag)
	if len(output) > 0 {
		os.Stdout.Write(output)
		if output[len(output)-1] != '\n' {
			fmt.Fprintln(os.Stdout)
		}
	}
	if err != nil {
		exitErr(err)
	}
}

func runGrypeCLI(args []string) {
	fs := flag.NewFlagSet("grype", flag.ExitOnError)
	target := fs.String("target", "", "path to the project directory")
	timeoutFlag := fs.Duration("timeout", 5*time.Minute, "timeout (e.g. 5m, 300s). Use 0 for no timeout")
	fs.Parse(args)

	ensureTarget(fs.Name(), target)

	resolved, err := scan.ResolveDirectory(*target)
	if err != nil {
		exitErr(err)
	}

	output, err := scan.RunGrype(context.Background(), resolved, *timeoutFlag)
	if len(output) > 0 {
		os.Stdout.Write(output)
		if output[len(output)-1] != '\n' {
			fmt.Fprintln(os.Stdout)
		}
	}
	if err != nil {
		exitErr(err)
	}
}

func ensureTarget(cmd string, target *string) {
	if target == nil || *target == "" {
		fmt.Fprintf(os.Stderr, "%s: --target is required\n\n", cmd)
		usage()
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, `MCP security scan CLI

Usage:
  mcp-cli <command> [flags]

Commands:
  semgrep   Run Semgrep SAST scan
  grype     Run Grype SCA scan

Global flags:
  -h, --help   Show this help message

Use "mcp-cli <command> -h" to see command-specific flags.
`)
	os.Exit(2)
}
