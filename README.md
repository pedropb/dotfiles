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

The repository is the source of truth. Edit files here, then activate the profile;
do not edit the managed files in `$HOME`.

## Bootstrap

Nix with flakes must be available in the target environment. The flake selects the
current Nix system automatically and supports Linux and macOS on both x86_64 and
ARM64. For Windows, install and activate it inside a WSL 2 distribution; native
Windows is not a Home Manager target.

The profile uses the invoking user's `USER` and `HOME`. First activation installs
the declared packages (`just`, `neovim`, `starship`, and `tmux`) and backs up
conflicting managed files as `*.before-home-manager`.

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
