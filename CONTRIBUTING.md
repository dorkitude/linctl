# Contributing to linctl

Thanks for contributing! This repo aims to keep changes simple, focused, and tested.

## Development

- Requirements: Go 1.22+, `gh` CLI (optional), `jq` (for examples), `golangci-lint` (optional).
- Useful targets:
  - `make deps` — install/tidy deps
  - `make build` — local build
  - `make test` — smoke tests (read-only commands)
  - `make lint` — lint if you have golangci-lint
  - `make fmt` — go fmt

## Release Checklist

Follow this checklist to cut a new release and update Homebrew:

1) Prepare
- Ensure README and help text match behavior.
- Update `CHANGELOG.md`:
  - Move relevant entries from `Unreleased` into the new `vX.Y.Z` section.
  - Add the release date.
  - Keep entries user-facing; prefer command names, flags, and behavior changes
    over internal implementation details.
- Run `make test` to verify smoke tests pass.
- Draft release notes from `CHANGELOG.md` (highlights, fixes, breaking changes).

2) Tag and Release (vX.Y.Z)
- Create tag and push:
  ```bash
  git tag vX.Y.Z -a -m "vX.Y.Z: short summary"
  git push origin vX.Y.Z
  ```
- Create GitHub release (with notes):
  ```bash
  gh release create vX.Y.Z \
    --title "linctl vX.Y.Z" \
    --notes "<highlights/fixes>"
  ```

3) Homebrew Tap Bump (auto)
- This repo has a GitHub Action that auto-opens a PR to the tap on release publish.
- Required secret: `HOMEBREW_TAP_TOKEN` (fine‑grained PAT with contents:write on `dorkitude/homebrew-linctl`).
  - Add in GitHub: repo Settings → Secrets and variables → Actions → New repository secret.

4) Homebrew Tap Bump (manual fallback)
If the action is disabled or no secret is configured:
```bash
TAG=vX.Y.Z
TARBALL=https://github.com/dorkitude/linctl/archive/refs/tags/${TAG}.tar.gz
curl -sL "$TARBALL" -o /tmp/linctl.tgz
SHA=$(shasum -a 256 /tmp/linctl.tgz | awk '{print $1}')

git clone https://github.com/dorkitude/homebrew-linctl.git
cd homebrew-linctl
git checkout -b bump-linctl-${TAG#v}
sed -i.bak -E "s|url \"[^\"]+\"|url \"$TARBALL\"|g" Formula/linctl.rb
sed -i.bak -E "s|sha256 \"[0-9a-f]+\"|sha256 \"$SHA\"|g" Formula/linctl.rb
rm -f Formula/linctl.rb.bak
git commit -am "linctl: bump to ${TAG}"
git push -u origin HEAD
gh pr create --title "linctl: bump to ${TAG}" --body "Update formula to ${TAG}." --base master --head bump-linctl-${TAG#v}
```

5) Validate
- After the tap PR merges:
  ```bash
  brew update && brew upgrade linctl
  linctl --version
  linctl docs | head -n 5
  ```
- Run a quick smoke test against your Linear workspace if possible.

6) Housekeeping
- Close any issues tied to the release.
- Start a new milestone if applicable.
- Add a fresh `Unreleased` section to `CHANGELOG.md` if the release process
  consumed it.
