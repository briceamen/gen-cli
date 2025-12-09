# Generative CLI

A proof-of-concept CLI generator that introspects the `go-scalingo` SDK and automatically generates Cobra commands for all available endpoints.

## Concept

Similar to how `sqlc generate` works for SQL, this tool parses the Go SDK interfaces and generates CLI commands automatically. When the SDK is updated with new endpoints, running `generative-cli generate` detects and scaffolds the missing commands.

## Architecture

```
generative-cli/
├── cmd/
│   ├── root.go           # Main Cobra root command
│   └── generate.go       # `generate` subcommand
├── generator/
│   ├── parser.go         # Go AST parser for SDK interfaces
│   ├── manifest.go       # TOML manifest management
│   ├── differ.go         # Diff SDK vs manifest to find new methods
│   ├── codegen.go        # Go code generation for Cobra commands
│   └── specgen.go        # TOML spec generation
├── render/
│   ├── styles.go         # Lipgloss styles
│   ├── table.go          # Table renderer (terminal-adaptive)
│   ├── detail.go         # Detail view renderer
│   ├── error.go          # Error/success formatters
│   └── registry.go       # Type → Renderer mapping
├── config/
│   └── config.go         # Auth config (reads ~/.config/scalingo/auth)
├── generated/
│   └── commands/         # Auto-generated command files
├── manifest.toml         # Registry of known SDK methods
└── main.go
```

## How It Works

### 1. SDK Parsing (`generator/parser.go`)

Uses Go's `go/ast` package to parse all `*Service` interfaces from `go-scalingo`:

```go
// Extracts method signatures like:
type AppsService interface {
    AppsList(ctx context.Context) ([]*App, error)
    AppsShow(ctx context.Context, app string) (*App, error)
    // ...
}
```

### 2. Manifest Tracking (`manifest.toml`)

A TOML file tracks which SDK methods have been discovered:

```toml
version = "1"
sdk_version = "v8"

[services.AppsService]
methods = ["AppsList", "AppsShow", "AppsCreate", ...]
```

### 3. Code Generation (`generator/codegen.go`)

Generates Cobra commands with:
- Automatic flag inference from method parameters
- SDK client initialization
- Renderer wiring based on return type

### 4. Rendering (`render/`)

Convention-based type mapping:

| Return Type | Renderer |
|-------------|----------|
| `[]*Type` or `[]Type` | Table view |
| `*Type` | Detail view (key-value) |
| `error` only | Success message |

Tables automatically adapt to terminal width using `lipgloss/table`.

## Usage

### Generate Commands

```bash
# Parse SDK and generate commands for new methods
./generative-cli generate --sdk-path=/path/to/go-scalingo

# Output:
# Parsing SDK at /path/to/go-scalingo...
# Found 35 services with 133 methods
# New methods to generate: 5
# Generated commands written to generated/commands/
```

### Use Generated Commands

```bash
# List apps
./generative-cli apps list

# Show app details
./generative-cli apps show --app-name my-app

# List regions
./generative-cli regions list

# JSON output
./generative-cli apps list --output json
```

## Authentication

The CLI reads authentication from:

1. `SCALINGO_API_TOKEN` environment variable (if set)
2. `~/.config/scalingo/auth` JSON file (same as the official Scalingo CLI)

This means if you're already logged in with `scalingo login`, this CLI will use the same credentials.

## Current Status

This is a PoC demonstrating the approach. Currently implemented:

- [x] SDK parser (AST-based interface extraction)
- [x] Manifest system (TOML-based method tracking)
- [x] Code generator (Cobra command scaffolding)
- [x] Render components (Lipgloss tables, detail views)
- [x] Auth integration (reads existing Scalingo CLI auth)
- [x] Working implementations for: `apps list`, `apps show`, `regions list`, `stacks list`

Most generated commands are stubs that need implementation wiring.

## Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling and tables
- [Viper](https://github.com/spf13/viper) - Configuration management
- [go-scalingo](https://github.com/Scalingo/go-scalingo) - Scalingo SDK

## Next Steps

- Wire up more command implementations
- Add interactive prompts for required parameters
- Support complex parameter types (JSON input for structs)
- Add `--columns` flag to select which columns to display
- Implement pagination for large result sets
