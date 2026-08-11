# dotfiles

Portable Home Manager configuration for macOS.

## Managed configuration

- Git: `~/.gitconfig` with an automatic personal identity
- Zsh: `.zshrc`, `.zprofile`
- Starship: `~/.config/starship.toml`
- WezTerm: `~/.config/wezterm/wezterm.lua`
- Neovim: `~/.config/nvim/init.lua`
- tmux: `~/.tmux.conf`

The repository is the source of truth. Edit files here, then activate the profile;
do not edit the managed files in `$HOME`.

## Bootstrap

Nix with flakes must be available. The profile uses the invoking user's `USER`
and `HOME`, so it can deploy to any macOS account. First activation installs the
declared packages (`just`, `neovim`, `starship`, and `tmux`) and backs up
conflicting managed files as `*.before-home-manager`.

```sh
nix run --impure .#home-manager -- switch --impure -b before-home-manager --flake .#default
```

After activation:

```sh
just check    # evaluate the flake
just switch   # deploy repository changes
```

WezTerm remains an external macOS GUI installation; Home Manager owns only its
configuration.
