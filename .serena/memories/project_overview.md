# aqua - Declarative CLI Version Manager

## Project Purpose
aqua is a declarative CLI version manager written in Go. It allows teams to:
- Switch tool versions per project
- Unify tool versions and how to install in teams and CI
- Continuous update by Renovate
- Provide lazy install functionality
- Build an ecosystem through its Registry
- Ensure security
- Make CLI tool management easy to use

## Tech Stack
- **Language**: Go 1.24.6
- **Build Tool**: Go build system
- **Task Runner**: cmdx (configuration in cmdx.yaml)
- **Testing**: Go test with race detection and coverage
- **Linting**: golangci-lint (configured in .golangci.yml)
- **JSON Schema**: For validating aqua.yaml and registry.yaml
- **CI/CD**: GitHub Actions (in .github/)
- **Release**: goreleaser (configured in .goreleaser.yml)
- **Dependencies**: Managed via go.mod/go.sum

## Key Dependencies
- github.com/Masterminds/sprig/v3 - Template functions
- github.com/goccy/go-yaml - YAML processing
- github.com/expr-lang/expr - Expression evaluation
- github.com/sirupsen/logrus - Logging
- github.com/urfave/cli/v3 - CLI framework

## Users
aqua is used by various companies including:
- Recruit Co., Ltd.
- Masterpoint Consulting
- Retty
- Mercari, Inc.
- Gunosy Inc.
- DeNA Co., Ltd.
- CADDi Inc.
- And many more listed in the README

## License
MIT License
