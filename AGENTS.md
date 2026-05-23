# AGENTS.md

## Purpose
Ship safe, small, tested changes to `linctl` with docs that match behavior.

## Operating Rules
- Keep patches minimal and focused; avoid drive-by refactors.
- Prefer `rg`/`go test`/existing scripts; do not invent parallel tooling.
- Never rewrite user-authored history (`reset --hard`, force-push) unless explicitly requested.
- If behavior changes, update docs in the same PR (`README.md`, `SKILL.md`,  as needed).
- If a change updates release versioning, do not consider it complete until Homebrew tap bump status and Nix installability are verified.

## Quality Bar
- Run: `go test ./...` before finalizing code changes.
- For CLI flag/command changes, add or update command tests in `cmd/*_test.go`.
- For GraphQL query/mutation changes, add or update API tests in `pkg/api/*_test.go`.
- If tests are skipped, state exactly what was skipped and why.

## Docs Freshness
- Treat `master_api_ref.md` as drift-prone; verify against current official Linear docs before editing claims (auth, scopes, rate limits, headers, deprecations).
  - When you identify drift around your feature area, edit `master_api_ref.md` in the same PR as your feature.
- Prefer exact terms from upstream docs; avoid hardcoded limits unless sourced.
- Keep examples runnable with current CLI.

## Fast Checks
- `linctl --version`
- `linctl --help`
- `go test ./...`

## Release + Package Managers (Owners Only)
- Owners cut releases from `master` using semver tags `vX.Y.Z`.
- Required for auto Homebrew bump: repo secret `HOMEBREW_TAP_TOKEN` (write access to `dorkitude/homebrew-linctl`).
- Version bump policy: release/version bumps must include Homebrew and Nix verification in the same workstream.
  - Homebrew verification means auto workflow success plus merged tap PR, or a merged manual tap bump PR.
  - Nix verification means the flake builds from the release tag, or the blocker is explicitly documented when Nix is unavailable in the release environment.
- Flow:
  1. `git push origin master`
  2. `git tag -a vX.Y.Z -m "vX.Y.Z: <summary>" && git push origin vX.Y.Z`
  3. `gh release create vX.Y.Z --title "linctl vX.Y.Z" --notes "..."`
  4. Verify `Bump Homebrew Tap` workflow succeeds.
  5. Merge bump PR in `dorkitude/homebrew-linctl` (close stale bump PRs).
  6. Confirm formula points to new tag+sha and `brew upgrade linctl` resolves.
  7. Confirm Nix installability from the tag:
     `nix build github:dorkitude/linctl/vX.Y.Z`
     `nix profile install github:dorkitude/linctl/vX.Y.Z`
     `linctl --version`
