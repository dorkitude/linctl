# Changelog

All notable user-facing changes to `linctl` are documented here.

Entries through v0.1.10 were reconstructed from GitHub release notes, annotated
tags, PR merge commits, and tag-to-tag commit history.

## [Unreleased]

### Docs

- Added `CHANGELOG.md` and release guidance for maintaining it.

## [v0.1.10] - 2026-06-25

### Added

- Added `--estimate` / `-e` to `linctl issue create` and `linctl issue update`.
- Added command tests verifying estimate values are sent in Linear GraphQL
  mutation inputs.

### Docs

- Documented estimate usage in README quick-start examples and command
  reference.

## [v0.1.9] - 2026-05-23

### Added

- Added first-class issue attachment discovery and download workflows:
  `linctl issue attachment list`, `linctl issue attachment download`, and
  `linctl issue get --download-attachments`.
- Added optional `pass(1)` credential storage via `LINCTL_PASS_NAME`, while
  preserving the default `~/.linctl-auth.json` behavior.
- Added dynamic `linctl mcp` commands backed by Linear GraphQL schema
  introspection and a cached tool registry: `mcp sync`, `mcp tools`, and
  `mcp call`.
- Added `linctl project update --content` for updating project document content.
- Added a Nix flake and documented `nix profile install github:dorkitude/linctl`.

### Fixed

- Avoided sending Linear auth headers to arbitrary external attachment URLs.
- Skipped non-downloadable links during attachment `--all` downloads.
- Avoided overwriting existing files during directory attachment downloads.

### Docs

- Added release-process guidance requiring package-manager verification for
  version bumps.

## [v0.1.8] - 2026-04-03

### Fixed

- Updated Personal API Key setup guidance to Linear's current Security & Access
  URL.
- Kept README, auth prompt, and test environment examples aligned with the new
  URL.

## [v0.1.7] - 2026-03-16

### Added

- Added `linctl issue relation` commands for listing, adding, and removing issue
  relations.
- Supported relation types `--blocks`, `--blocked-by`, `--related`,
  `--duplicate`, and `--similar`.
- Added relation command aliases including `ls`, `create`, `rm`, and `delete`.

### Fixed

- Made relation labels direction-aware so `blocks` and `blocked by` are displayed
  consistently.
- Made JSON relation output shape more consistent for scripting.

## [v0.1.6] - 2026-03-02

### Changed

- Renamed team workflow "status" commands to "state" terminology:
  `linctl team statuses` became `linctl team state list`, and
  `linctl team status-update` became `linctl team state update`.
- Added `--state` / `-s` to `linctl issue create`.

### Docs

- Updated README and SKILL command references for `team state` commands.
- Added release-process docs requiring Homebrew verification for version bumps.

## [v0.1.5] - 2026-03-01

### Added

- Added `linctl graphql` as a raw GraphQL escape hatch using existing linctl
  authentication.
- Supported GraphQL input via positional query, `--query`, `--file`, and stdin.
- Added variables support via `--variables` and `--variables-file`.

### Docs

- Added README examples for raw GraphQL workflows.
- Refreshed the agent playbook and Linear API reference.

## [v0.1.4] - 2026-03-01

### Added

- Added team workflow status management commands: `linctl team statuses` and
  `linctl team status-update`.
- Added API support for the `workflowStateUpdate` mutation.

### CI

- Hardened the Homebrew tap bump workflow.
- Added manual dispatch support for the Homebrew tap bump workflow.

### Docs

- Updated README and SKILL docs for team status commands.
- Removed a hardcoded status color example.

## [v0.1.3] - 2026-03-01

### Added

- Added `linctl label create --is-group` for creating group labels.
- Added command/API coverage for group label creation.

### Docs

- Documented label group creation in README and refreshed command-surface
  guidance.

## [v0.1.2] - 2026-02-28

### Added

- Added `linctl issue search`.
- Added project CRUD commands for create, update, archive/delete, and permanent
  delete.
- Added issue project and milestone linking.
- Added parent issue updates.
- Added `LINCTL_API_KEY` authentication override.
- Added cycle filtering/display and issue URL/GitHub PR attachment creation.
- Added full comment CRUD commands.
- Added label CRUD commands and issue label assignment.
- Added agent session commands and issue delegation support.
- Added Homebrew tap auto-bump workflow and contributing docs.

### Fixed

- Fixed project list team filtering.
- Fixed nil user/actor rendering for comments and plaintext issue output.

## [v0.1.1] - 2025-09-09

### Added

- Added Homebrew formula packaging.
- Added richer issue listing and detail output, including sorting, time-based
  filters, and include-completed behavior.
- Added comment, project, team, and user command areas.
- Added built-in `linctl docs`.
- Added smoke testing.

### Changed

- Focused authentication documentation and behavior on Personal API Key auth.
- Synchronized README and help examples with the expanded command surface.

### Fixed

- Fixed Linear attachment `source` unmarshalling.

## [v0.1.0] - 2025-07-13

### Added

- Initial Go/Cobra CLI structure.
- Added Personal API Key authentication commands and credential storage.
- Added initial issue commands and Linear GraphQL API client support.
- Added table, plaintext, and JSON output helpers.
- Added initial README, Linear API reference snapshots, and development
  Makefile.

### Changed

- Refined root command header rendering.

[Unreleased]: https://github.com/dorkitude/linctl/compare/v0.1.10...HEAD
[v0.1.10]: https://github.com/dorkitude/linctl/compare/v0.1.9...v0.1.10
[v0.1.9]: https://github.com/dorkitude/linctl/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/dorkitude/linctl/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/dorkitude/linctl/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/dorkitude/linctl/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/dorkitude/linctl/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/dorkitude/linctl/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/dorkitude/linctl/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/dorkitude/linctl/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/dorkitude/linctl/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/dorkitude/linctl/releases/tag/v0.1.0
