# GitHub Copilot Instructions for aqua

This document contains GitHub Copilot-specific instructions. For general project guidelines, see [AGENTS.md](../AGENTS.md).

## Core Guidelines

Refer to [AGENTS.md](../AGENTS.md) for all project conventions, including:
- Language, commit messages, code style
- Project structure and package responsibilities
- Testing and validation commands
- Error handling patterns

## Copilot-Specific Instructions

### Code Suggestions

When suggesting code completions:
- Prioritize consistency with existing code patterns in the file
- Complete imports based on packages already used in the project
- Suggest idiomatic Go patterns

### Context Awareness

- Use surrounding code context to infer appropriate variable names
- Match the indentation and formatting style of the current file
- Suggest appropriate error messages based on function context

### Autocomplete Behavior

- For test files, automatically suggest table-driven test patterns
- For CLI commands, follow the `urfave/cli/v3` patterns used throughout the project
- For error handling, wrap errors with context using `fmt.Errorf` with `%w`

## Quick Reference

See [AGENTS.md](../AGENTS.md) for project-specific patterns and commands.
