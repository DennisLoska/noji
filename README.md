# noji

A CLI to automate developer workflow tasks like creating/updating PRs and updating ticket status, powered by opencode prompts.


## Installation

```sh
# install (embeds version/commit/date from git)
make install

# ensure your install path is on PATH
export PATH="$(go env GOBIN 2>/dev/null || echo "$(go env GOPATH)/bin"):$PATH"

# verify
noji -v
```

Install a specific version
```sh
git fetch --tags
git checkout v0.1.0
make install
noji -v
```

Local build (no install)
```sh
make build
./bin/noji -v
```

Requirements: opencode CLI and GitHub CLI (gh) must be installed and on PATH.


## Quick start

```sh
# show config paths and ensure first-run setup
noji config path

# list models and pick one
noji models
noji use github-copilot/gpt-5

# create or update a PR description
noji pr create
noji pr update

# update your ticket using the ticket prompt
noji ticket update
noji ticket edit $TICKET_ID

# see PRs with reviews requested from you
noji pr reviews --limit 5
```

## Configuration

noji stores configuration and user-editable prompt templates under the OS config directory. You can override the base directory using an environment variable.

- Default location: `${XDG_CONFIG_HOME:-$HOME/.config}/noji`
- Override: set `NOJI_CONFIG_HOME` to any directory (no trailing `/noji` needed).

On every run, noji ensures the following exist (creating them if missing):

- `config.yaml` – stores the selected model
- `prompts/` directory with initial templates:

User prompt files are never overwritten after creation. Edit them freely to customize instructions for your org/repo.

## Prompts and models

- Prompts are plain text files under the user config prompts dir. Their contents are passed verbatim to the selected opencode model.
- Change the model anytime with:

```sh
noji use <model>
```

- Discover models:

```sh
noji models
```

The PR commands use the local git history and your prompt templates to draft or update the PR description.

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
noji --help
```


## License

MIT
