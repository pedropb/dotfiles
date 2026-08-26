# dotfiles

Portable Home Manager configuration for macOS and Linux, including Linux distributions running under WSL 2.

## Managed configuration

- Git: `~/.gitconfig` with an automatic personal identity

- Zsh: Home Manager-generated `.zshrc` and `.zprofile`, with Oh My Zsh's `git`
  aliases, autosuggestions, `you-should-use`, and syntax highlighting
- Starship: `~/.config/starship.toml`
- WezTerm: `~/.config/wezterm/wezterm.lua`
- Neovim: `~/.config/nvim/init.lua`
- tmux: `~/.config/tmux/tmux.conf` (the XDG path, read by tmux 3.1+). A
  host-provided tmux config is sourced first, then overridden here.

Private, machine-specific configuration — Git identities, credential
helpers, private [`cln`](tools/cln/README.md) forges — belongs in the
git-ignored `home/local.nix`, documented with examples in
[`home/local.nix.md`](home/local.nix.md).

## Git authentication

The profile installs `gh` and `glab`. GitHub HTTPS remotes use
`gh auth git-credential` automatically. Authenticate that CLI separately:

```sh
gh auth login
```

`local.nix` also carries private credential helper entries, one per HTTPS
remote host; see [`home/local.nix.md`](home/local.nix.md) for the option,
its exact-hostname-match rule, and an example. Run `just switch` after
changing one.

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

The repository is the source of truth. Edit files here, then activate the profile;
do not edit the managed files in `$HOME`.

## Tools

[`tools/cln`](tools/cln/README.md) clones a repository by
provider/namespace/repo shorthand instead of the full HTTPS remote URL, e.g.
`cln gh dotfiles` or `cln octocat/Hello-World`, into `~/src/<host>/<namespace>/<repo>`
so it's immediately reachable with `scd <namespace>/<repo>`. Its providers come from
`dotfiles.cln.defaultProvider`/`dotfiles.cln.providers`, rendered to
`~/.config/cln/config.toml`; the profile sets a `gh` provider for
`github.com`. Add private forges in `home/local.nix`; see
[`home/local.nix.md`](home/local.nix.md) for the option and an example.

`tools/` holds each such standalone program in its own directory, built and
tested independently of this Home Manager profile; see
[`tools/README.md`](tools/README.md).

## Bootstrap

Nix with flakes must be available in the target environment. The flake selects the
current Nix system automatically and supports Linux and macOS on both x86_64 and
ARM64. For Windows, install and activate it inside a WSL 2 distribution; native
Windows is not a Home Manager target.

The profile uses the invoking user's `USER` and `HOME`. Its package inventory is
defined in [`home/packages.nix`](home/packages.nix).

```sh
nix run --impure .#home-manager -- switch --impure -b before-home-manager --flake .#default
```

After activation:

```sh
just check    # evaluate the flake
just switch   # deploy repository changes
```

WezTerm remains an external GUI installation; Home Manager owns only its
configuration. The macOS-only background blur is enabled only on macOS.
