# Code Style and Conventions for aqua

## Go Code Style
- **Go Version**: 1.24.6
- **Package Structure**: Domain-driven design with packages organized by functionality
- **Import Organization**: Standard library, third-party, then local packages

## Naming Conventions
- **Packages**: Lowercase, single word when possible (e.g., `config`, `controller`, `domain`)
- **Files**: Lowercase with underscores for multi-word (e.g., `update_checksum.go`)
- **Types/Structs**: PascalCase (e.g., `Controller`, `Config`)
- **Methods/Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE for configuration constants
- **Interfaces**: Often end with "-er" suffix (e.g., `Reader`, `Finder`)

## Code Organization
- Main package in `cmd/aqua/`
- Core logic in `pkg/` directory
- Each package has focused responsibility
- Mock files named with `mock_` prefix or `mock.go` suffix
- Test files follow `*_test.go` convention

## Error Handling
- Named error returns (e.g., `namedErr error`)
- Error wrapping with context
- Use of `logrus` for structured logging
- Custom error types in `pkg/errors/`

## Comments and Documentation
- Minimal inline comments unless complex logic
- Package-level documentation where needed
- Function/method documentation for exported items
- No unnecessary comments per project preference

## Testing
- Unit tests alongside code files
- Race detection enabled (`-race` flag)
- Coverage testing with atomic mode
- Test helpers in `pkg/testutil/`

## Linting Configuration (golangci-lint)
- All linters enabled by default with specific exclusions
- Key disabled linters: depguard, godot, lll, nlreturn, wsl
- Formatters: gci, gofumpt, goimports
- Generated files and mocks have relaxed rules

## Dependency Injection
- Use of google/wire for dependency injection
- Constructor functions typically named `New()` or `NewController()`

## Interfaces
- Define interfaces where implementations will vary
- Keep interfaces small and focused
- Define interfaces close to where they're used

## File System Operations
- Use `afero.Fs` interface for testability
- Consistent use of `os.Stat()` for file existence checks

## Logging
- Structured logging with logrus
- Log entries with fields for context
- Error logging with `logerr` package
