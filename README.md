# wt — git worktree manager

[![CI](https://github.com/glennr/wt/actions/workflows/release.yml/badge.svg)](https://github.com/glennr/wt/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A transparent wrapper around `git worktree` that adds central worktree storage (`~/worktrees/<project>/<branch>`) and project-specific bootstrap/teardown via `--run`.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/glennr/wt/main/install.sh | sh
```

By default, `wt` is installed to `~/.local/bin`. To change the install directory:

```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/glennr/wt/main/install.sh | sh
```

Requires Go 1.26+.

### Uninstall

```bash
rm ~/.local/bin/wt
```

### Other methods

```bash
# Go install
go install github.com/glennr/wt@latest

# From source
git clone https://github.com/glennr/wt.git
cd wt
task install   # builds to ~/.local/bin/wt
```

## Usage

### Create a worktree

```bash
wt add -b feature/login --repo ~/src/myapp
# Creates ~/worktrees/myapp/feature-login

# With post-creation bootstrap
wt add -b feature/login --repo ~/src/myapp \
  --run='make bootstrap-from SRC=$WT_SOURCE_DIR'
```

All flags not recognized by wt (`-b`, `-t`, `--track`, etc.) are passed through to `git worktree add`.

### Remove a worktree

```bash
wt rm feature/login --repo ~/src/myapp

# Remove current worktree (when inside one)
wt rm

# With pre-removal teardown
wt rm feature/login --repo ~/src/myapp --run='make teardown'

# Skip confirmation prompt
wt rm feature/login --repo ~/src/myapp --force
```

### List worktrees

```bash
wt ls                          # all projects
wt ls --repo ~/src/myapp       # one project
```

Output:

```
PROJECT  BRANCH         COMMIT   STATUS  PATH
myapp    feature-login  a1b2c3d  active  /home/user/worktrees/myapp/feature-login
api      new-endpoint   e4f5g6h  active  /home/user/worktrees/api/new-endpoint
```

### Get worktree path

```bash
wt path feature/login --repo ~/src/myapp
# /home/user/worktrees/myapp/feature-login
```

### Navigate to a worktree

```bash
cd $(wt path feature/login)
```

### Run a command in a worktree

```bash
wt exec feature/login --repo ~/src/myapp -- make test
# Runs `make test` with cwd set to the worktree dir

wt exec feature/login -- ls -la
# Uses auto-detected repo
```

The command runs with `WT_*` environment variables set (see below).

### Prune stale references

```bash
wt prune
```

## Repo resolution

All commands auto-detect the source repo from cwd when `--repo` is omitted:

1. `--repo` flag (if provided)
2. cwd inside a git worktree → resolves to main repo
3. cwd inside a git repo → uses that repo
4. Error

**Project name** = basename of the resolved repo path.

## `--run` environment

Commands passed to `--run` execute via `sh -c` with these environment variables:

| Variable | Description |
|---|---|
| `WT_SOURCE_DIR` | Source repo path |
| `WT_BRANCH` | Branch name (original, e.g. `gr/app-123`) |
| `WT_PROJECT` | Project name (e.g. `myapp`) |
| `WT_WORKTREE_DIR` | Absolute worktree path |

Timing: **post-creation** for `add`, **pre-removal** for `rm`. On failure, wt logs the error but does not roll back.

## Directory layout

```
~/worktrees/
├── myapp/
│   ├── feature-login/
│   └── fix-bug/
├── api/
│   └── new-endpoint/
```

Branch names are sanitized for directory use: `/` → `-`.

## Inspiration

Inspired by [incident.io](https://incident.io/)'s worktree workflow, described in [Practical Guide to Git Worktrees](https://incident.io/blog/git-worktrees).

## Development

Requires [Go 1.26+](https://go.dev/dl/) and [Task](https://taskfile.dev/) (task runner). [mise](https://mise.jdx.dev/) is used to manage tool versions (golangci-lint, goreleaser, git-cliff).

```bash
# Setup
mise install        # install tool dependencies
task setup          # configure git hooks (conventional commits)

# Build & test
task build          # build to bin/wt (version from git tags)
task test           # run tests
task lint           # run golangci-lint
task ci             # run test + lint (same as CI)
```

### Commit messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/). A commit-msg hook enforces the format after running `task setup`.

```
feat: add new command
fix: handle edge case in path resolution
docs: update README
chore: bump dependencies
```

### Releasing

Releases are automated via GitHub Actions + [GoReleaser](https://goreleaser.com/). Tags are the source of truth for versions.

```bash
# Create a release (runs CI, generates changelog, tags, and pushes)
task release VERSION=0.2.0

# Test the release build locally (no publish)
task release-local

# Generate/update CHANGELOG.md without releasing
task changelog
```

The `release` task will:
1. Run `task ci` (test + lint)
2. Generate `CHANGELOG.md` via git-cliff
3. Commit the changelog, create an annotated tag `v<VERSION>`
4. Push to `main` with the tag

GitHub Actions then builds cross-platform binaries and publishes a GitHub release.
