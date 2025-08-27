# noji

A CLI to automate developer workflow tasks like creating/updating PRs and updating ticket status, powered by opencode prompts.

## Features

- PR workflows
  - `noji pr create` – guide model to draft a new PR description
  - `noji pr update` – update the current PR description
- Ticket workflows
  - `noji ticket update` – craft an update for your tracker ticket
- Model management
  - `noji models` – list available opencode models
  - `noji use <model>` – select and persist a default model
- Config visibility
  - `noji config path` – show config and prompts locations
- Reviews
  - `noji pr reviews` – list open PRs where your review is requested (uses gh)

## Installation

Build from source:

```sh
make build
# binary at ./bin/noji
```

Or with go (installs to $GOBIN or $GOPATH/bin):

```sh
go install ./cmd/noji
```

Requirements: opencode CLI and GitHub CLI (gh) must be installed and available on PATH.

## Quick start

```sh
# show config paths and ensure first-run setup
./bin/noji config path

# list models and pick one
./bin/noji models
./bin/noji use github-copilot/gpt-5

# create a PR description using your prompt template text
./bin/noji pr create

# update PR description later
./bin/noji pr update

# update your ticket using the ticket prompt
./bin/noji ticket update

# see PRs with reviews requested from you
./bin/noji pr reviews
./bin/noji pr reviews --org your-org --limit 5
```

## Configuration

noji stores configuration and user-editable prompt templates under the OS config directory. You can override the base directory using an environment variable.

- Default location: `${XDG_CONFIG_HOME:-$HOME/.config}/noji`
- Override: set `NOJI_CONFIG_HOME` to any directory (no trailing `/noji` needed).

On every run, noji ensures the following exist (creating them if missing):

- `config.yaml` – stores the selected model
- `prompts/` directory with initial templates:
  - `pr_create.txt`
  - `pr_update.txt`
  - `ticket_update.txt`

User prompt files are never overwritten after creation. Edit them freely to customize instructions for your org/repo.

## Prompts and models

- Prompts are plain text files under the user config prompts dir. Their contents are passed verbatim to the selected opencode model.
- Change the model anytime with:

```sh
noji use <model>
```

- Discover models:

```sh
# create a PR description using your prompt template text
./bin/noji pr create

# update PR description later
./bin/noji pr update
```

This updates the currently open PR for the checked-out branch, reusing the `prompts/pr_update.txt` template.

The command uses the local git history and the `prompts/pr_create.txt` template to draft a PR description. It will open an editor to review before submission.

## Environment variables

- `NOJI_CONFIG_HOME` – overrides the base config directory. Example:

```sh
export NOJI_CONFIG_HOME="$HOME/.config"  # results in $HOME/.config/noji
```

## Troubleshooting

- opencode not found: ensure the `opencode` CLI is installed and on PATH.
- Permissions creating config: run once in a shell with permissions to create `${NOJI_CONFIG_HOME:-$XDG_CONFIG_HOME:-$HOME/.config}/noji`.
- Model not accepted: run `noji models` to verify and `noji use <model>` to set.
- Prompts missing: run any command; the app seeds missing prompt files automatically. Existing files are never overwritten.

## Development

Project structure:

- `cmd/noji/main.go` – Cobra entrypoint
- `internal/commands/*` – subcommands and wiring
- `internal/config/config.go` – config paths and ensure/seed logic
- `internal/opencode/opencode.go` – thin wrapper for opencode CLI
- `prompts/*.txt` – repository templates used to seed user prompts on first run

Common tasks:

```sh
# build
make build

# run the local binary
./bin/noji --help

# lint (if you add linters in the future)
# make lint
```

## Roadmap ideas

- Embed prompt templates via go:embed to guarantee seeding for globally installed binaries
- `noji prompts check` to list present/missing prompt files
- Richer help output with examples
- Better preflight checks for dependencies (opencode/gh/git)

## License

MIT
