package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bozz33/sublimeadmin/generator"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "make:resource":
		makeResource(os.Args[2:])
	case "make:page":
		makePage(os.Args[2:])
	case "make:widget":
		makeWidget(os.Args[2:])
	case "make:enum":
		makeEnum(os.Args[2:])
	case "make:action":
		makeAction(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("SublimeAdmin CLI v%s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func makeResource(args []string) {
	fs := flag.NewFlagSet("make:resource", flag.ExitOnError)
	output := fs.String("output", ".", "Output directory")
	force := fs.Bool("force", false, "Overwrite existing files")
	dryRun := fs.Bool("dry-run", false, "Show what would be generated without writing")
	verbose := fs.Bool("verbose", false, "Verbose output")
	_ = fs.Parse(args)

	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "Usage: sublimego make:resource <Name> [flags]")
		fmt.Fprintln(os.Stderr, "Example: sublimego make:resource User --output=./")
		os.Exit(1)
	}

	gen, err := generator.New(&generator.Options{
		Force:     *force,
		DryRun:    *dryRun,
		Verbose:   *verbose,
		OutputDir: *output,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generator error: %v\n", err)
		os.Exit(1)
	}

	if err := generator.GenerateResource(gen, name, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating resource: %v\n", err)
		os.Exit(1)
	}

	data := generator.NewResourceData(name)
	fmt.Printf("\nGenerated %s resource\n", name)
	fmt.Printf("   Slug: %s\n", data.Slug)
	fmt.Printf("   Package: %s\n", data.PackageName)
	fmt.Printf("   Type: %s\n", data.TypeName)
}

func makePage(args []string) {
	fs := flag.NewFlagSet("make:page", flag.ExitOnError)
	group := fs.String("group", "", "Navigation group")
	icon := fs.String("icon", "", "Material icon name")
	output := fs.String("output", ".", "Output directory")
	force := fs.Bool("force", false, "Overwrite existing files")
	verbose := fs.Bool("verbose", false, "Verbose output")
	_ = fs.Parse(args)

	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "Usage: sublimego make:page <Name> [flags]")
		fmt.Fprintln(os.Stderr, "Example: sublimego make:page Settings --output=./")
		os.Exit(1)
	}

	gen, err := generator.New(&generator.Options{
		Force:     *force,
		Verbose:   *verbose,
		OutputDir: *output,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generator error: %v\n", err)
		os.Exit(1)
	}

	if err := generator.GeneratePageWithOptions(gen, name, *output, *group, *icon, 100); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating page: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated page: %s\n", name)
}

func makeWidget(args []string) {
	fs := flag.NewFlagSet("make:widget", flag.ExitOnError)
	output := fs.String("output", ".", "Output directory")
	force := fs.Bool("force", false, "Overwrite existing files")
	verbose := fs.Bool("verbose", false, "Verbose output")
	_ = fs.Parse(args)

	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "Usage: sublimego make:widget <Name> [flags]")
		fmt.Fprintln(os.Stderr, "Example: sublimego make:widget RevenueChart --output=./")
		os.Exit(1)
	}

	gen, err := generator.New(&generator.Options{
		Force:     *force,
		Verbose:   *verbose,
		OutputDir: *output,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generator error: %v\n", err)
		os.Exit(1)
	}

	if err := generator.GenerateWidget(gen, name, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating widget: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated widget: %s\n", name)
}

func makeEnum(args []string) {
	fs := flag.NewFlagSet("make:enum", flag.ExitOnError)
	output := fs.String("output", ".", "Output directory")
	force := fs.Bool("force", false, "Overwrite existing files")
	verbose := fs.Bool("verbose", false, "Verbose output")
	_ = fs.Parse(args)

	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "Usage: sublimego make:enum <Name> [flags]")
		fmt.Fprintln(os.Stderr, "Example: sublimego make:enum OrderStatus --output=./")
		os.Exit(1)
	}

	gen, err := generator.New(&generator.Options{
		Force:     *force,
		Verbose:   *verbose,
		OutputDir: *output,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generator error: %v\n", err)
		os.Exit(1)
	}

	if err := generator.GenerateEnum(gen, name, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating enum: %v\n", err)
		os.Exit(1)
	}
}

func makeAction(args []string) {
	fs := flag.NewFlagSet("make:action", flag.ExitOnError)
	output := fs.String("output", ".", "Output directory")
	force := fs.Bool("force", false, "Overwrite existing files")
	verbose := fs.Bool("verbose", false, "Verbose output")
	_ = fs.Parse(args)

	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "Usage: sublimego make:action <Name> [flags]")
		fmt.Fprintln(os.Stderr, "Example: sublimego make:action ArchivePost --output=./")
		os.Exit(1)
	}

	gen, err := generator.New(&generator.Options{
		Force:     *force,
		Verbose:   *verbose,
		OutputDir: *output,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generator error: %v\n", err)
		os.Exit(1)
	}

	if err := generator.GenerateAction(gen, name, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating action: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`SublimeAdmin CLI v%s
A code generator for the SublimeAdmin Go framework.

Usage:
  sublimego <command> [arguments] [flags]

Commands:
  make:resource <Name>   Generate a new resource (table + form + CRUD)
  make:page <Name>       Generate a custom page
  make:widget <Name>     Generate a dashboard widget
  make:enum <Name>       Generate a typed enum (HasLabel, HasColor, HasIcon)
  make:action <Name>     Generate a custom action handler

Global Flags:
  --output <dir>         Output directory (default: current dir)
  --force                Overwrite existing files
  --dry-run              Show what would be generated (no writes)
  --verbose              Verbose output

Examples:
  sublimego make:resource User --output=./
  sublimego make:resource Product --output=./
  sublimego make:page Settings --output=./
  sublimego make:widget RevenueChart --output=./
  sublimego make:enum OrderStatus --output=./
  sublimego make:action ArchivePost --output=./

`, version)
}
