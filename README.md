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

Private, machine-specific configuration — which user the profile builds for,
Git identities, credential helpers, private [`cln`](tools/cln/README.md)
forges — lives outside the checkout in `~/.config/dotfiles/local.toml`,
maintained by [`dotfiles-local`](tools/dotfiles-local/README.md) and
documented with examples in [`home/local-config.md`](home/local-config.md).

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
just switch
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

The repository is the source of truth. Edit files here, then activate the profile;
do not edit the managed files in `$HOME`.

## Tools

[`tools/cln`](tools/cln/README.md) clones a repository by
provider/namespace/repo shorthand instead of the full HTTPS remote URL, e.g.
`cln gh dotfiles` or `cln octocat/Hello-World`, into `~/src/<host>/<namespace>/<repo>`
so it's immediately reachable with `scd <namespace>/<repo>`. Its providers come from
`dotfiles.cln.defaultProvider`/`dotfiles.cln.providers`, rendered to
`~/.config/cln/config.toml`; the profile sets a `gh` provider for
`github.com`. Add private forges with `dotfiles-local provider add`; see
[`home/local-config.md`](home/local-config.md) for the schema and an example.

`tools/` holds each such standalone program in its own directory, built and
tested independently of this Home Manager profile; see
[`tools/README.md`](tools/README.md).

## Bootstrap

Nix with flakes must be available in the target environment. The flake supports
Linux and macOS on both x86_64 and ARM64. For Windows, install and activate it
inside a WSL 2 distribution; native Windows is not a Home Manager target.

The user, home directory, and system the profile builds for are recorded in
`~/.config/dotfiles/local.toml` instead of being read from the environment,
which is what keeps evaluation pure: no `--impure`, and `nix flake check` works
on a fresh clone. The package inventory is defined in
[`home/packages.nix`](home/packages.nix).

```sh
just bootstrap
```

That prompts for your private configuration, then activates the profile,
preserving conflicting files as `*.before-home-manager`.

After activation:

```sh
just check    # evaluate the flake and this machine's profile
just switch   # deploy repository changes
just local    # show the machine-specific configuration in use
```

WezTerm remains an external GUI installation; Home Manager owns only its
configuration. The macOS-only background blur is enabled only on macOS.
