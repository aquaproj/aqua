# Task Completion Checklist for aqua

When completing any coding task in the aqua project, ensure you:

## 1. Code Quality Checks
- [ ] Run linting: `cmdx lint` or `golangci-lint run`
- [ ] Run go vet: `cmdx vet` or `go vet ./...`
- [ ] Format code: `go fmt ./...` (if needed)
- [ ] Ensure imports are organized (goimports is run by golangci-lint)

## 2. Testing
- [ ] Run tests: `cmdx test` or `go test ./... -race -covermode=atomic`
- [ ] Check test coverage if significant changes: `cmdx coverage <target>`
- [ ] Ensure all tests pass with race detection enabled

## 3. Build Verification
- [ ] Verify build succeeds: `cmdx build` or `go build -o dist/aqua ./cmd/aqua`
- [ ] If changing dependencies, run `go mod tidy`

## 4. Code Generation (if applicable)
- [ ] Run wire if dependency injection changed: `cmdx wire`
- [ ] Generate JSON schema if config structures changed: `cmdx js`
- [ ] Validate configs if schema changed: `cmdx validate-js`

## 5. Documentation
- [ ] Update relevant documentation if API/behavior changed
- [ ] Ensure exported functions have appropriate documentation

## 6. Pre-commit Checklist
- [ ] Review changes with `git diff`
- [ ] Stage only intended changes
- [ ] Write clear, descriptive commit message
- [ ] Follow existing commit message patterns in the repository

## Quick Command Sequence
For most changes, run this sequence:
```bash
# Test and verify
cmdx test
cmdx lint
cmdx vet

# Build to ensure compilation
cmdx build

# Review changes before committing
git diff
git status
```

## Note
The project uses cmdx as a task runner. All common development commands are defined in `cmdx.yaml`. Use `cmdx` to see available commands.