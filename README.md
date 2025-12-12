# Generative CLI

A proof-of-concept CLI generator that introspects the `go-scalingo` SDK and automatically generates Cobra commands for all available endpoints.

## Concept

Similar to how `sqlc generate` works for SQL, this tool parses the Go SDK interfaces and generates CLI commands automatically. The manifest is the source of truth: run the generator CLI to record new SDK methods, then regenerate the code from that manifest (without overwriting your edits).

There are two binaries:

- **Generator CLI**: `./cmd/generator` → keeps the manifest up to date and regenerates code.
- **Generated CLI**: `./cmd/runtime` → the actual CLI built from the generated command set.

## Architecture

```
generative-cli/
├── cmd/
│   ├── generator/        # Entry point for generator CLI (go run ./cmd/generator)
│   └── runtime/          # Entry point for generated CLI (go run ./cmd/runtime)
├── generatorcli/         # Generator commands (update-manifest, generate)
├── runtimecli/           # Root command that wires all generated commands
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
└── manifest.toml         # Registry of known SDK methods
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

A TOML file tracks which SDK methods have been discovered and whether they should be generated. Parameters are stored with names so you can tweak flag names manually without being overwritten.

```toml
version = 1
sdk_version = ""

[services]
  [services.AppsService]

    [[services.AppsService.methods]]
      name = "AppsList"
      params = []
      returns = "[]*App"
      generated = true

    [[services.AppsService.methods]]
      name = "AppsShow"
      returns = "*App"
      generated = true

      [[services.AppsService.methods.params]]
        name = "app"
        type = "string"
```

#### Manifest Types

The manifest uses the following Go structures:

```go
type Manifest struct {
    Version    int                        `toml:"version"`
    SDKVersion string                     `toml:"sdk_version"`
    Services   map[string]ManifestService `toml:"services"`
}

type ManifestService struct {
    Methods []ManifestMethod `toml:"methods"`
}

type ManifestMethod struct {
    Name      string          `toml:"name"`
    Params    []ManifestParam `toml:"params"`
    Returns   string          `toml:"returns"`
    Generated bool            `toml:"generated"`
}

type ManifestParam struct {
    Name string `toml:"name,omitempty"`
    Type string `toml:"type"`
}
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

The project uses Go's standard `go generate` mechanism:

```bash
# Regenerate CLI from SDK (updates manifest + generates code)
go generate ./...

# Output:
# Parsing SDK at: vendor/github.com/Scalingo/go-scalingo/v8
# Found 35 services
# Manifest already up to date
# Generating commands for 134 methods across 35 services
# Generation complete!
```

You can also run the generator commands manually:

```bash
# Update manifest with any new SDK methods
go run ./cmd/generator update-manifest

# Override SDK location if you want to scan a different checkout
go run ./cmd/generator update-manifest --sdk-path=/path/to/go-scalingo

# Generate commands from manifest.toml
go run ./cmd/generator generate
```

### Update SDK and Regenerate

When the upstream SDK is updated:

```bash
# Update vendored SDK
go get github.com/Scalingo/go-scalingo/v8@latest
go mod vendor

# Regenerate (will add new methods to manifest)
go generate ./...
```

### Customize manifest entries

- Flip `generated = false` on any method to skip codegen for it.
- Adjust `params` names/types to tweak flag names (e.g., rename `app-i-d` to `app-id`).
- Change `returns` to influence renderer selection (`[]Type` → table, `*Type` → detail, empty → success).
`generate` never rewrites the manifest, so your edits stay intact; only `update-manifest` appends missing methods.

### Use Generated Commands

```bash
# Build the generated CLI (or use go run ./cmd/runtime)
go build -o scalingo-gen ./cmd/runtime

# List apps
./scalingo-gen apps list

# Show app details
./scalingo-gen apps show --app-name my-app

# List regions
./scalingo-gen regions list

# JSON output
./scalingo-gen apps list --output json
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
