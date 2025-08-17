# aqua Codebase Structure

## Root Directory Structure
```
.
├── cmd/                    # Command-line applications
│   ├── aqua/              # Main aqua CLI application
│   └── gen-jsonschema/    # JSON schema generator
├── pkg/                   # Core packages and business logic
├── scripts/               # Build and utility scripts
├── tests/                 # Integration and e2e tests
├── json-schema/           # JSON schema definitions
├── logo/                  # Project logo assets
├── website/               # Documentation website
├── .github/               # GitHub Actions workflows
├── cmdx.yaml              # Task runner configuration
├── go.mod & go.sum        # Go module dependencies
├── .golangci.yml          # Linting configuration
├── .goreleaser.yml        # Release configuration
└── README.md              # Project documentation
```

## pkg/ Directory (Core Logic)
Key packages include:
- **cli/**: CLI command implementations and utilities
- **config/**: Configuration handling and parsing
- **config-finder/**: Logic to find config files
- **config-reader/**: Reading and parsing config files
- **controller/**: Main business logic controllers
- **domain/**: Core domain models and types
- **download/**: Package download functionality
- **install-registry/**: Registry installation logic
- **installpackage/**: Package installation logic
- **checksum/**: Checksum validation and management
- **runtime/**: Runtime environment handling
- **github/**: GitHub API interactions
- **slsa/**: SLSA provenance verification
- **cosign/**: Cosign signature verification
- **keyring/**: Keyring/credentials management
- **errors/**: Custom error types
- **testutil/**: Testing utilities
- **template/**: Template processing
- **expr/**: Expression evaluation
- **fuzzyfinder/**: Fuzzy search functionality

## Controller Pattern
Controllers in `pkg/controller/` handle main operations:
- Each operation has its own sub-package
- Controllers coordinate between different components
- Business logic is encapsulated in controllers

## Configuration
- Main config in `aqua.yaml` at project roots
- Registry config in `registry.yaml`
- JSON schemas for validation in `json-schema/`
- Global config paths supported

## Testing Structure
- Unit tests alongside source files (*_test.go)
- Integration tests in `tests/` directory
- Mock files with `mock_` prefix or `mock.go` suffix
- Test utilities in `pkg/testutil/`

## CLI Structure
- Main entry point: `cmd/aqua/main.go`
- CLI framework: urfave/cli/v3
- Commands organized in `pkg/cli/` package
- Parameter handling through `util.Param` struct

## Build and Release
- cmdx.yaml defines all development tasks
- .goreleaser.yml handles releases
- Docker support via Dockerfile
- GitHub Actions for CI/CD
