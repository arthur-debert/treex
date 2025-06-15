# Installation Guide

## From GitHub Releases (Recommended)

The easiest way to install `treex` is to download the latest pre-compiled binary from the [GitHub Releases](https://github.com/arthur-debert/treex/releases) page.

### 1. Download the Binary

1. Navigate to the [latest release](https://github.com/arthur-debert/treex/releases/latest)
2. Under the **Assets** section, download the archive for your operating system and architecture:
   - **macOS (Intel)**: `treex_Darwin_x86_64.tar.gz`
   - **macOS (Apple Silicon)**: `treex_Darwin_arm64.tar.gz`
   - **Linux (x86_64)**: `treex_Linux_x86_64.tar.gz`
   - **Linux (ARM64)**: `treex_Linux_arm64.tar.gz`
   - **Windows**: `treex_Windows_x86_64.zip`

3. Extract the archive:

   ```bash
   tar -xzf treex_Darwin_arm64.tar.gz  # macOS/Linux
   # or
   unzip treex_Windows_x86_64.zip      # Windows
   ```

4. Move the `treex` binary to a directory in your system's `PATH`:

   ```bash
   # macOS or Linux
   sudo mv ./treex /usr/local/bin/treex
   
   # Or for user-only installation
   mv ./treex ~/.local/bin/treex
   ```

### 2. Verify Installation

```bash
treex --version
```

## From Homebrew (macOS/Linux)

If you're on macOS or Linux, you can install `treex` using Homebrew:

```bash
# Add the tap
brew tap arthur-debert/homebrew-tools

# Install treex
brew install treex
```

This automatically installs the binary, man page, and shell completions.

## Shell Completion

`treex` provides auto-completion scripts for Bash, Zsh, and Fish shells.

### Bash

**Option 1: Source on demand**
Add to your `~/.bash_profile` or `~/.bashrc`:

```bash
if command -v treex &> /dev/null; then
  source <(treex completion bash)
fi
```

**Option 2: System-wide installation**

```bash
# Linux
sudo treex completion bash > /etc/bash_completion.d/treex

# macOS (with Homebrew)
treex completion bash > $(brew --prefix)/etc/bash_completion.d/treex
```

### Zsh

**Option 1: Source on demand**
Add to your `~/.zshrc`:

```zsh
if command -v treex &> /dev/null; then
  source <(treex completion zsh)
fi
```

Make sure `compinit` is enabled in your `.zshrc`.

**Option 2: Install to completion directory**

```zsh
treex completion zsh > "${fpath[1]}/_treex"
```

### Fish

```fish
treex completion fish > ~/.config/fish/completions/treex.fish
```

## Man Pages

If you installed via Homebrew, man pages are automatically available:

```bash
man treex
```

For manual installations, you can generate the man page:

```bash
treex man --path /usr/local/share/man/man1/
```

## From Source

If you have Go 1.22+ installed, you can build `treex` from source:

```bash
git clone https://github.com/arthur-debert/treex.git
cd treex
go build -o treex ./cmd/treex
sudo mv treex /usr/local/bin/
```

## Verification

After installation, verify that `treex` is working:

```bash
# Check version
treex --version

# Test on current directory
treex .

# View help
treex --help
```

## Uninstallation

### Homebrew

```bash
brew uninstall treex
brew untap arthur-debert/homebrew-tools
```

### Manual Installation

```bash
sudo rm /usr/local/bin/treex
sudo rm /usr/local/share/man/man1/treex.1  # if installed
rm -f ~/.local/share/bash-completion/completions/treex  # bash
rm -f ~/.local/share/zsh/site-functions/_treex  # zsh
rm -f ~/.config/fish/completions/treex.fish  # fish
```

## Troubleshooting

### Command Not Found

- Ensure `/usr/local/bin` (or your chosen directory) is in your `PATH`
- Check that the binary has execute permissions: `chmod +x /usr/local/bin/treex`

### Shell Completion Not Working

- Restart your shell or run `source ~/.bashrc` / `source ~/.zshrc`
- Verify completion is installed: `complete -p treex` (bash) or `which _treex` (zsh)

### Permission Denied

- Use `sudo` when installing to system directories
- Consider installing to `~/.local/bin` instead for user-only installation
