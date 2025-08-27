# noji

A CLI to automate developer workflow tasks like creating/updating PRs and updating ticket status, powered by opencode prompts.

## Commands and examples

- General
  - `noji --help`
  - `noji current` – print the currently selected model

- Models
  - `noji models` – list available opencode models
    - Example:
      ```sh
      noji models
      ```
  - `noji use <model>` – select and persist a default model
    - Example:
      ```sh
      noji use github-copilot/gpt-5
      ```

- Pull requests
  - `noji pr create` – create a PR description using your prompt template
    - Examples:
      ```sh
      noji pr create
      ```
  - `noji pr update` – update the current PR description
    - Examples:
      ```sh
      noji pr update
      ```
   - `noji pr reviews` – list open PRs where your review is requested (uses gh)
     - Flags:
       - `--org <org>` filter by GitHub organization
       - `--limit <n>` limit number of results (0=all)
       - `--json` output raw JSON
       - `--infer-orgs` infer orgs (default true when --org not provided)
       - `--no-bots` exclude PRs from bot authors (default true)
       - `--bots` only PRs from bot authors (overrides --no-bots)
       - `--links auto|always|never` control clickable hyperlink output (default auto)
     - Examples:
       ```sh
       noji pr reviews
       noji pr reviews --org your-org --limit 5
       noji pr reviews --json
       noji pr reviews --bots --limit 20
       ```
   - `noji pr comments` – list your PRs with human comments; optionally classify severity and priority
     - Flags:
       - `--repo OWNER/REPO` limit to a repo
       - `--state open|closed|all` PR state (default open)
       - `--drafts` include draft PRs (default true)
       - `--no-bots` exclude bot comments (default true)
       - `--limit <n>` limit number of PRs (0=all)
       - `--since YYYY-MM-DD` only PRs updated on/after date
       - `--json` output JSON
       - `--classify` enable opencode-based severity classification and derived Priority
       - `--links auto|always|never` control clickable hyperlink output (default auto)
     - Notes:
       - Output shows only the raw PR URL line (no separate "Open PR" clickable line). Individual comment arrows (↗) remain clickable when links are enabled.
     - Examples:
       ```sh
       noji pr comments
       noji pr comments --repo owner/repo --state open --no-bots --limit 20
       noji pr comments --since 2025-07-28 --json
       noji pr comments --classify
       ```

- Tickets
  - `noji ticket update` – craft an update for your tracker ticket
    - Example:
      ```sh
      noji ticket update
      ```

- Config
  - `noji config path` – show config and prompts locations
    - Example:
      ```sh
      noji config path
      ```

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
noji config path

# list models and pick one
noji models
noji use github-copilot/gpt-5

# create a PR description using your prompt template text
noji pr create

# update PR description later
noji pr update

# update your ticket using the ticket prompt
noji ticket update

# see PRs with reviews requested from you
noji pr reviews
noji pr reviews --org your-org --limit 5
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

## Roadmap ideas

- Embed prompt templates via go:embed to guarantee seeding for globally installed binaries
- `noji prompts check` to list present/missing prompt files
- Richer help output with examples
- Better preflight checks for dependencies (opencode/gh/git)

## License

MIT
