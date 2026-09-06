# dotfiles

Portable Home Manager configuration for macOS and Linux, including Linux
distributions running under WSL 2.

## Install

The flake supports x86_64 and ARM64 macOS and Linux. Native Windows is not a
Home Manager target; install and activate it inside a WSL 2 distribution.

### 1. Install Nix

Install Nix with the [official installer](https://nixos.org/download/), then
open a new shell so its environment is available.

On macOS:

```sh
curl --proto '=https' --tlsv1.2 -L https://nixos.org/nix/install | sh
```

On Linux or WSL 2 with systemd:

```sh
curl --proto '=https' --tlsv1.2 -L https://nixos.org/nix/install | sh -s -- --daemon
```

For Linux without systemd, use the official single-user installer instead:

```sh
curl --proto '=https' --tlsv1.2 -L https://nixos.org/nix/install | sh -s -- --no-daemon
```

This repository uses flakes. Enable Nix's `nix-command` and `flakes` features
if the installer did not already enable them:

```sh
mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/nix"
printf '%s\n' 'experimental-features = nix-command flakes' \
  >> "${XDG_CONFIG_HOME:-$HOME/.config}/nix/nix.conf"
```

### 2. Clone and activate

```sh
git clone https://github.com/pedropb/dotfiles.git ~/src/github.com/pedropb/dotfiles
cd ~/src/github.com/pedropb/dotfiles
./bin/bootstrap.sh
```

`bootstrap.sh` interactively creates the private machine configuration, then
activates the profile. Files that conflict with managed files are preserved as
`*.before-home-manager`.

### 3. Update the profile

After activation, Home Manager provides `just`. Run these commands from the
checkout:

```sh
just check    # evaluate the flake and this machine's profile
just switch   # deploy repository changes
just local    # show the machine-specific configuration in use
```

The repository is the source of truth. Edit files here, then run `just switch`;
do not edit managed files in `$HOME`.

## Managed configuration

- Git: `~/.gitconfig` with an automatic personal identity
- Zsh: Home Manager-generated `.zshrc` and `.zprofile`, with Oh My Zsh's `git`
  aliases, autosuggestions, `you-should-use`, and syntax highlighting
- Starship: `~/.config/starship.toml`
- WezTerm: `~/.config/wezterm/wezterm.lua`
- Neovim: `~/.config/nvim/init.lua`
- tmux: `~/.config/tmux/tmux.conf` (the XDG path, read by tmux 3.1+). A
  host-provided tmux config is sourced first, then overridden here.

WezTerm remains an external GUI installation; Home Manager owns only its
configuration. The macOS-only background blur is enabled only on macOS.

## Private machine configuration

Private, machine-specific configuration—the profile's user, home directory,
Git identities, credential helpers, and private
[`cln`](tools/cln/README.md) forges—lives outside the checkout at
`~/.config/dotfiles/local.toml`. It is maintained by
[`dotfiles-local`](tools/dotfiles-local/README.md) and documented with examples
in [`home/local-config.md`](home/local-config.md).

The package inventory is defined in [`home/packages.nix`](home/packages.nix).

## Git authentication

The profile installs `gh` and `glab`. GitHub HTTPS remotes use
`gh auth git-credential` automatically. Authenticate that CLI separately:

```sh
gh auth login
```

`local.toml` also carries private credential helper entries, one per HTTPS
remote host:

```sh
dotfiles-local helper set gitlab.example.com '!glab auth git-credential'
./bin/switch.sh
```

See [`home/local-config.md`](home/local-config.md) for the key and its
exact-hostname-match rule.

For a self-managed GitLab instance, create a personal access token with `api`
and `write_repository` scopes, then store it in the operating-system keyring:

```sh
read -rs GLAB_PAT
printf '%s\n' "$GLAB_PAT" |
  glab auth login --hostname gitlab.example.com --stdin --use-keyring
unset GLAB_PAT
```

`--stdin` avoids glab's interactive Git-configuration prompt; Home Manager
owns that configuration. Do not leave `GITLAB_TOKEN`, `GITLAB_ACCESS_TOKEN`,
or `OAUTH_TOKEN` exported: glab treats each as a token for every configured
GitLab host.

## Tools

[`tools/cln`](tools/cln/README.md) clones a repository by
provider/namespace/repo shorthand instead of the full HTTPS remote URL, e.g.
`cln gh dotfiles` or `cln octocat/Hello-World`, into
`~/src/<host>/<namespace>/<repo>` so it is immediately reachable with
`scd <namespace>/<repo>`. Its providers come from
`dotfiles.cln.defaultProvider`/`dotfiles.cln.providers`, rendered to
`~/.config/cln/config.toml`; the profile sets a `gh` provider for `github.com`.
Add private forges with `dotfiles-local provider add`; see
[`home/local-config.md`](home/local-config.md) for the schema and an example.

`tools/` holds each such standalone program in its own directory, built and
tested independently of this Home Manager profile; see
[`tools/README.md`](tools/README.md).
