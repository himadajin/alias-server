# AGENTS.md

## Project Overview

This repository contains a small local URL alias server written in Go.
It reads a JSON config file, accepts simple one-segment aliases,
and redirects matching local requests to configured target URLs.

## Specification

### Configuration Files

Config files are JSON files that define a default port and a map of link names
to target URLs. The schema is:

```json
{
  "defaultPort": 15555,
  "links": {
    "go": "https://go.dev/"
  }
}
```

Schema rules:

- `defaultPort`
  - TCP port number.
  - May be omitted only when the CLI `-port` option is provided.
- `links` maps link names to absolute target URLs.
  - Link names are single path segments.
  - Link names may use only lowercase ASCII letters, digits, and hyphens.

CLI `-port` takes precedence over `defaultPort`.
Query parameters, wildcard matching, nested paths, and URL encoding are not supported.

Sample config: `example/config.json`.

### HTTP Behavior

- Unsupported paths and unknown links return `404`.
- Unsupported HTTP methods return `405`.

## Development Rules

- Keep the implementation simple and dependency-free.
- Do not add new dependencies unless the user explicitly asks for them.
- Keep the app as a small root-level `package main`; do not introduce `cmd/` or `internal/` unless the project grows enough to justify it.

## Agent Workflow

Use `*.local.json` for local config file names, such as `config.local.json`.
After changes, run formatting and tests.
For startup, CLI, or build-output changes, also verify the build.

Development tasks are defined in `Taskfile.yml`:

- Use `task format` when formatting Go files.
- Use `task test` when verifying tests.
- Use `task build` when creating a local binary.
- Use `task run` when starting the server for local development.
- Use `task clean` when removing local build outputs.

`bin/` is the output directory used by development commands.
