# AI Assistant Guidelines for pinact

This document contains common guidelines for AI assistants working on the pinact project.
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

### Validation (go vet)

```sh
cmdx v
```

This command runs `go vet ./...` to check for common Go mistakes.

### Testing

```sh
cmdx t
```

This command runs all tests in the project.

Both commands should pass before committing changes.

## Project Structure

```
```

## Package Responsibilities

### pkg/cli

Command-line interface layer that handles command parsing, flag processing, and routing to appropriate subcommands.

### pkg/config


### pkg/controller

### pkg/github

GitHub API client integration and authentication management.


## Testing

- Run all tests: `cmdx t` or `go test ./...`
- Run specific package tests: `go test ./pkg/controller/install`

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
2. Make changes and ensure `cmdx v` and `cmdx t` pass
3. Write clear commit messages following Conventional Commits
4. Create PR with descriptive title and body
5. Wait for CI checks to pass
6. Request review if needed

## Important Commands

```sh
# Validate code (go vet)
cmdx v

# Run tests
cmdx t

# Build the project
go build ./cmd/aqua

# Generate JSON schema
cmdx js

# Run aqua locally
go run ./cmd/aqua
```
