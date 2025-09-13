# AI Assistant Guidelines for aqua

This document contains common guidelines for AI assistants working on the aqua project.
Individual AI-specific documents (like CLAUDE.md, CLINE.md) should reference this guide.

## Language

This project uses **English** for all code comments, documentation, and communication.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Common Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools
- `ci`: Changes to CI configuration files and scripts

### Examples

```
feat: add GitHub token management via keyring
fix: handle empty configuration file correctly
docs: add function documentation to controller package
chore(deps): update dependency aquaproj/aqua-registry to v4.403.0
```

## Code Validation

After making code changes, **always run** the following commands to validate and test:

### Linting (golangci-lint)

```sh
cmdx l
# or
cmdx lint
```

This command runs `golangci-lint run` with all configured linters. The configuration is in `.golangci.yml`.

### Validation (go vet)

```sh
cmdx v
# or
cmdx vet
```

This command runs `go vet ./...` to check for common Go mistakes.

### Testing

```sh
cmdx t
# or
cmdx test
```

This command runs all tests in the project with race detection enabled.

All commands should pass before committing changes.

## Project Structure

```
.
├── cmd/                    # Command-line applications
│   ├── aqua/              # Main aqua CLI application
│   └── gen-jsonschema/    # JSON schema generator
├── pkg/                   # Core packages and business logic
│   ├── cli/               # CLI commands and routing
│   ├── config/            # Configuration handling
│   ├── controller/        # Main business logic controllers
│   ├── domain/            # Core domain models
│   ├── download/          # Package download functionality
│   ├── install-registry/  # Registry installation
│   ├── checksum/          # Checksum validation
│   └── ...                # Other packages
├── scripts/               # Build and utility scripts
├── tests/                 # Integration and e2e tests
├── json-schema/           # JSON schema definitions
├── cmdx.yaml              # Task runner configuration
├── go.mod & go.sum        # Go module dependencies
├── .golangci.yml          # Linting configuration
└── .goreleaser.yml        # Release configuration
```

## Package Responsibilities

### pkg/cli

Command-line interface layer that handles command parsing, flag processing, and routing to appropriate subcommands. Uses urfave/cli/v3 framework.

### pkg/config

Configuration management including:
- Reading and parsing aqua.yaml configuration files
- Config validation and merging
- Environment variable handling
- Global and local config support

### pkg/controller

Main business logic controllers for aqua operations:
- `install`: Package installation logic
- `update`: Package update functionality
- `updatechecksum`: Checksum update operations
- `exec`: Command execution wrapper
- `which`: Path resolution for installed tools
- Each controller coordinates between different components

### pkg/domain

Core domain models and types:
- Package definitions
- Registry structures
- Configuration models
- Common interfaces

### pkg/github

GitHub API client integration:
- Release artifact downloading
- GitHub authentication management
- API rate limiting handling
- Release and tag fetching

### pkg/download

Package download functionality:
- HTTP client for downloading packages
- Progress tracking
- Retry logic
- Cache management

### pkg/checksum

Checksum validation and management:
- SHA256/SHA512 verification
- Checksum file parsing
- Security verification

### pkg/install-registry

Registry installation and management:
- Registry downloading
- Registry validation
- Local registry cache


## Testing

- Run all tests with race detection: `cmdx t` or `go test ./... -race -covermode=atomic`
- Run specific package tests: `go test ./pkg/controller/install -race`
- Run coverage test: `cmdx coverage <target>` or `bash scripts/coverage.sh <target>`
- Coverage reports are saved to `.coverage/` directory

## Dependencies

This project uses:

- [aqua](https://aquaproj.github.io/) for tool version management
- [cmdx](https://github.com/suzuki-shunsuke/cmdx) for task runner

## Code Style Guidelines

1. Follow standard Go conventions
2. Use meaningful variable and function names
3. Add comments for exported functions and types
4. Keep functions focused and small
5. Handle errors explicitly
6. Use context for cancellation and timeouts
7. Always end files with a newline character

## Pull Request Process

1. Create a feature branch from `main`
2. Make changes and ensure all checks pass:
   - `cmdx lint` - Run linting
   - `cmdx vet` - Run go vet
   - `cmdx test` - Run tests
3. Write clear commit messages following Conventional Commits
4. Create PR with descriptive title and body
5. Wait for CI checks to pass
6. Request review if needed

## Important Commands

```sh
# Run linting (golangci-lint)
cmdx l
# or
cmdx lint

# Validate code (go vet)
cmdx v
# or
cmdx vet

# Run tests with race detection
cmdx t
# or
cmdx test

# Run coverage test for specific target
cmdx c <target>
# or
cmdx coverage <target>

# Build the project
cmdx build
# or directly:
go build -o dist/aqua ./cmd/aqua

# Build and install locally
cmdx i
# or
cmdx install

# Generate JSON schema
cmdx js

# Validate aqua.yaml and registry.yaml with JSON Schema
cmdx validate-js

# Run wire for dependency injection
cmdx w
# or
cmdx wire

# Run aqua locally via go run
cmdx run [args]
# or directly:
go run ./cmd/aqua [args]

# Test in clean Docker container
cmdx docker

# Format and tidy dependencies
go fmt ./...
go mod tidy
```
