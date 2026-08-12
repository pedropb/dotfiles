# dotfiles

Portable Home Manager configuration for macOS and Linux, including Linux distributions running under WSL 2.

## Managed configuration

- Git: `~/.gitconfig` with an automatic personal identity

- Zsh: Home Manager-generated `.zshrc` and `.zprofile`, with Oh My Zsh's `git`
  aliases, autosuggestions, `you-should-use`, and syntax highlighting
- Starship: `~/.config/starship.toml`
- WezTerm: `~/.config/wezterm/wezterm.lua`
- Neovim: `~/.config/nvim/init.lua`
- tmux: `~/.tmux.conf`

Private path-specific Git identities belong in the ignored `home/local.nix`. It
can add `dotfiles.git.conditionalIdentities.<key>` with `gitdir`, `name`, and
`email`; activation writes the matching identity to `~/.config/git/identities`
and includes it only for that Git directory prefix.

```nix
{ ... }:
{
  dotfiles.git.conditionalIdentities.example = {
    gitdir = "~/src/example.com/";
    name = "Example Author";
    email = "author@example.com";
  };
}
```

The repository is the source of truth. Edit files here, then activate the profile;
do not edit the managed files in `$HOME`.

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
