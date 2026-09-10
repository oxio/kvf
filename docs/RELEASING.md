# Release Process

Releases are created automatically by GitHub Actions. Pushing a tag that
matches the rules below triggers the [Release workflow](.github/workflows/release.yml),
which runs the tests, builds binaries for all supported platforms, and
publishes a GitHub release with the binaries attached.

## Release Rules

| Rule | Value |
|------|-------|
| Branch | Tags must point to a commit on the `v2` branch |
| Tag format | `vMAJOR.MINOR.PATCH`, optionally suffixed with `-alpha.N` or `-beta.N` |
| Examples | `v2.1.0`, `v2.1.0-alpha.1`, `v2.1.0-beta.3` |

Tags that do not match the format (e.g. `v2.1`, `2.1.0`, `v2.1.0-rc.1`), or
that point to a commit outside the `v2` branch, fail the workflow's
validation job with an explanatory error — no release is created.

Tags with an `-alpha.N` or `-beta.N` suffix are published as
**pre-releases**.

## Supported Platforms

Binaries are built for all platforms supported by `task build:all`:

| OS | Architecture | Binary name |
|----|--------------|-------------|
| macOS | amd64 | `kvf-darwin-amd64` |
| macOS | arm64 | `kvf-darwin-arm64` |
| Linux | amd64 | `kvf-linux-amd64` |
| Linux | arm64 | `kvf-linux-arm64` |
| Linux | arm (armv7) | `kvf-linux-arm` |
| Windows | amd64 | `kvf-windows-amd64.exe` |
| Windows | arm64 | `kvf-windows-arm64.exe` |

## Step-by-Step Release Process

```bash
# 1. Make sure you are on the v2 branch with everything pushed
git checkout v2
git status
git log origin/v2..HEAD   # should print nothing

# 2. Run tests locally (optional sanity check)
task test

# 3. Create an annotated tag
git tag -a v2.1.0 -m "Release v2.1.0"

# 4. Push the tag — this triggers the release workflow
git push origin v2.1.0
```

For a pre-release, use a suffix:

```bash
git tag -a v2.1.0-beta.1 -m "Beta release v2.1.0-beta.1"
git push origin v2.1.0-beta.1
```

## What the Workflow Does

1. **Validate** — checks the tag format and that the tagged commit is on the
   `v2` branch.
2. **Build** — cross-compiles the binaries for all supported platforms in a
   matrix, embedding the version via ldflags so `kvf --version` reports the
   tag.
3. **Release** — creates the GitHub release with `--generate-notes` and
   attaches all binaries.

## Post-Release Verification

After the workflow finishes, verify:

1. **Workflow run**: https://github.com/oxio/kvf/actions
2. **Release and binaries**: https://github.com/oxio/kvf/releases — all 7
   binaries should be attached.
3. **Installation test**:

```bash
curl -sL https://raw.githubusercontent.com/oxio/kvf/main/install.sh | bash
kvf --version
```

## Troubleshooting

### Validation failed: tag format

Delete and re-create the tag with a valid name:

```bash
git tag -d v2.1
git push origin --delete v2.1
git tag -a v2.1.0 -m "Release v2.1.0"
git push origin v2.1.0
```

### Validation failed: tag not on the `v2` branch

The tagged commit must be reachable from `v2`. Either move the tag to a
commit on `v2`, or merge/cherry-pick the commit into `v2` first:

```bash
git checkout v2
git merge <commit-or-branch>
git push origin v2
# then re-tag the merge commit (or the original commit, now reachable from v2)
```

### Release assets are missing or wrong

Delete the release and re-run the workflow:

```bash
gh release delete v2.1.0 --yes
git tag -d v2.1.0
git push origin --delete v2.1.0
# re-create and push the tag to trigger a fresh run
```

### Version shows as "dev"

This only affects manually built binaries. Release binaries always embed the
tag. For local builds, use `task build`, which embeds `git describe` output.

## Quick Reference

| Action | Command |
|--------|---------|
| Run tests | `task test` |
| Build locally (current platform) | `task build` |
| Build all platforms locally | `task build:all` |
| Create tag | `git tag -a v2.1.0 -m "Release v2.1.0"` |
| Push tag (triggers release) | `git push origin v2.1.0` |
| List releases | `gh release list` |
| Clean build artifacts | `task clean` |
