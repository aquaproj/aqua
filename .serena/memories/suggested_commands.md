# Development Commands for aqua

## Essential Commands via cmdx

### Testing
- `cmdx test` or `cmdx t` - Run all tests with race detection and coverage
- `cmdx coverage <target>` or `cmdx c <target>` - Run coverage test for specific target
- `go test ./... -race -covermode=atomic` - Direct test command

### Linting & Code Quality
- `cmdx lint` or `cmdx l` - Run golangci-lint
- `cmdx vet` or `cmdx v` - Run go vet
- `golangci-lint run` - Direct lint command

### Building
- `cmdx build` - Build aqua binary to dist/aqua
- `cmdx install` or `cmdx i` - Build and install aqua locally with version info
- `go build -o dist/aqua ./cmd/aqua` - Direct build command

### Development & Running
- `cmdx run [args]` - Run aqua via go run with arguments
- `cmdx docker` - Build and run in a clean Docker container for testing

### Code Generation
- `cmdx wire` or `cmdx w` - Run google/wire for dependency injection
- `cmdx js` - Generate JSON Schema
- `cmdx validate-js` - Validate aqua.yaml and registry.yaml with JSON Schema

## Git Commands (Darwin/macOS)
- `git status` - Check current status
- `git diff` - View unstaged changes
- `git add <file>` - Stage changes
- `git commit -m "message"` - Commit changes
- `git push` - Push to remote
- `git pull` - Pull from remote
- `git log --oneline -n 10` - View recent commits

## File System Commands (Darwin/macOS)
- `ls -la` - List all files with details
- `cd <directory>` - Change directory
- `pwd` - Print working directory
- `find . -name "*.go"` - Find Go files
- `grep -r "pattern" .` - Search for pattern in files
- `cat <file>` - Display file contents

## Go Development Commands
- `go mod download` - Download module dependencies
- `go mod tidy` - Clean up go.mod and go.sum
- `go get <package>` - Add or update dependency
- `go generate ./...` - Run code generation
- `go fmt ./...` - Format code

## JSON Schema Validation
- `ajv --spec=draft2020 -s json-schema/aqua-yaml.json -d aqua.yaml` - Validate aqua.yaml
- `ajv --spec=draft2020 -s json-schema/registry.json -d tests/main/registry.yaml` - Validate registry
