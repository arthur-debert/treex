# Release Guide

This document explains how to set up and use the automated release process for `treex`.

## Prerequisites

### 1. GitHub Personal Access Token (PAT)

You need a GitHub Personal Access Token with `repo` scope to allow the release workflow to push formula updates to your Homebrew tap repository.

**Create a PAT:**

1. Go to [GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Give it a descriptive name like "treex-homebrew-tap"
4. Set expiration (recommend 1 year)
5. Select the `repo` scope (this gives full repository access)
6. Click "Generate token"
7. **Copy the token immediately** (you won't be able to see it again)

**Add the token to your repository:**

1. Go to your `treex` repository on GitHub
2. Navigate to Settings > Secrets and variables > Actions
3. Click "New repository secret"
4. Name: `HOMEBREW_TAP_TOKEN`
5. Value: Paste your PAT
6. Click "Add secret"

### 2. Homebrew Tap Repository

You already have this set up at <https://github.com/arthur-debert/homebrew-tools>.

## Release Process

The release process is fully automated through GitHub Actions and GoReleaser. Here's how it works:

### Automatic Release (Recommended)

1. **Make your changes** on the `main` branch
2. **Commit and push** your changes
3. **Create and push a version tag**:

   ```bash
   git tag -a v0.1.0 -m "Release version 0.1.0"
   git push origin v0.1.0
   ```

4. **That's it!** The GitHub Action will automatically:
   - Build binaries for Linux, macOS, and Windows (x86_64 and ARM64)
   - Create archives with documentation, man pages, and shell completions
   - Generate checksums
   - Create a GitHub Release with all assets
   - Generate a Homebrew formula
   - Push the formula to your `homebrew-tools` tap repository

### Manual Testing (Before Release)

You can test the release process without creating a tag:

```bash
# Install GoReleaser if you haven't already
go install github.com/goreleaser/goreleaser@latest

# Test the release configuration (creates snapshot build)
goreleaser release --snapshot --clean

# This creates builds in the dist/ directory for testing
```

## Release Configuration

The release process is configured in two files:

### `.goreleaser.yml`

- Defines build targets (OS/architecture combinations)
- Configures archive creation with completions and man pages
- Sets up Homebrew formula generation
- Specifies which files to include in releases

### `.github/workflows/release.yml`

- Triggers on version tags (`v*.*.*`)
- Sets up Go environment
- Runs GoReleaser with proper permissions

## Versioning

Follow [Semantic Versioning](https://semver.org/):

- `v1.0.0` - Major version (breaking changes)
- `v0.2.0` - Minor version (new features, backward compatible)
- `v0.1.1` - Patch version (bug fixes)

Examples:

```bash
git tag -a v0.1.0 -m "Initial release"
git tag -a v0.2.0 -m "Add filtering features"
git tag -a v0.2.1 -m "Fix depth limit bug"
```

## What Gets Released

Each release includes:

### GitHub Release Assets

- `treex_Darwin_x86_64.tar.gz` - macOS Intel binary + docs
- `treex_Darwin_arm64.tar.gz` - macOS Apple Silicon binary + docs
- `treex_Linux_x86_64.tar.gz` - Linux x86_64 binary + docs
- `treex_Linux_arm64.tar.gz` - Linux ARM64 binary + docs
- `treex_Windows_x86_64.zip` - Windows binary + docs
- `checksums.txt` - SHA256 checksums for all archives

### Each archive contains

- `treex` - The binary
- `README.md` - Project documentation
- `LICENSE` - License file
- `docs/` - All documentation files
- `completions/` - Shell completion scripts
  - `treex.bash` - Bash completion
  - `_treex` - Zsh completion
  - `treex.fish` - Fish completion
- `man/man1/treex.1` - Man page

### Homebrew Formula

- Automatically generated `treex.rb` formula
- Pushed to `arthur-debert/homebrew-tools` repository
- Includes proper URLs and checksums for all platforms
- Sets up binary, man page, and completion installation

## Troubleshooting

### Release Failed

Check the GitHub Actions logs:

1. Go to your repository > Actions tab
2. Click on the failed release workflow
3. Expand the "Run GoReleaser" step to see detailed logs

Common issues:

- **Invalid PAT**: Check that `HOMEBREW_TAP_TOKEN` is set correctly
- **PAT expired**: Generate a new token and update the secret
- **Build errors**: Check Go version compatibility and dependencies

### Formula Not Updated

If the Homebrew formula wasn't updated:

1. Check that the `HOMEBREW_TAP_TOKEN` has `repo` scope
2. Verify the token has access to `arthur-debert/homebrew-tools`
3. Check the GoReleaser logs for tap-related errors

### Manual Formula Update

If automation fails, you can manually update the formula:

1. Download the checksums from the GitHub release
2. Update `Formula/treex.rb` in your tap repository
3. Update URLs and SHA256 checksums for each platform

## Security Notes

- The `HOMEBREW_TAP_TOKEN` has write access to your tap repository
- Keep the token secure and rotate it regularly
- The token is only used in GitHub Actions, never exposed in logs
- Consider using fine-grained PATs (beta) for better security once they support actions
