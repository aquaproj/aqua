# Claude-Specific Guidelines for aqua

This document contains Claude-specific guidelines. For general project guidelines, see [AI_GUIDE.md](AI_GUIDE.md).

## Core Guidelines

All general project guidelines are documented in [AI_GUIDE.md](AI_GUIDE.md). Please refer to that document for:
- Language conventions
- Commit message format
- Code validation and testing
- Project structure
- Code style guidelines
- Common tasks and workflows

## Claude-Specific Notes

### Context Window Management

- Be mindful of context window limits when reading large files
- Use file offset and limit parameters when appropriate
- Summarize lengthy outputs to conserve context

### Tool Usage

- Prefer batch operations when possible to improve efficiency
- Use the Task tool for complex multi-step operations
- Always validate changes with `cmdx v` and `cmdx t`

### Communication Style

- Keep responses concise and to the point
- Focus on actionable information
- Avoid unnecessary explanations unless requested

## Quick Reference

For quick access to common commands and guidelines, see [AI_GUIDE.md](AI_GUIDE.md#important-commands).
