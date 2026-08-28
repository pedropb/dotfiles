# dotfiles-local

Maintains `~/.config/dotfiles/local.toml`: the machine-specific and private
half of this repository's Home Manager profile — which user and home directory
to build for, work git identities, per-host credential helpers, and private
`cln` providers.

Its reason to exist is purity. The profile used to read `USER`, `HOME`, and a
git-ignored `home/local.nix` at evaluation time, which forced `--impure` on
every `nix` invocation and made a fresh-clone `nix flake check` impossible.
`dotfiles-local` reads the environment **once**, writes it down, and the flake
consumes the result as data through its `local` input. The impurity moves from
every evaluation to a single recorded, reviewable file.

Schema and examples: [`home/local-config.md`](../../home/local-config.md).

## Commands

```
setup
  ensure [-f]                 record this host in [machine]; print the config directory
  init                        interactive setup for identities, helpers, and providers

inspect
  path                        print the config directory
  show                        print the current local.toml
  check                       validate local.toml without changing it

edit
  identity add <name> --author <name> --email <email> --gitdir <prefix> [--gitdir <prefix>]
  identity rm <name>
  helper set <host> <command>
  helper rm <host>
  provider add <alias> --type <github|gitlab> --host <host> [--namespace <ns>] [--default]
  provider rm <alias>
  default-provider [<alias>]
```

`ensure` is what `just switch` calls: it is idempotent, never prompts, prints
nothing but the config directory on stdout, and is safe over ssh and in CI.
`init` is the only interactive command, and every prompt it asks has a
flag-driven equivalent above — nothing is reachable only by wizard.

Before the first activation the binary is not yet on `PATH`; run it through the
flake instead:

```sh
nix run .#dotfiles-local -- init
```

## Design notes

- **The config directory is outside the checkout** — `$DOTFILES_LOCAL_DIR`,
  else `$XDG_CONFIG_HOME/dotfiles`, else `~/.config/dotfiles`. It survives
  re-clones and cannot be committed by accident.
- **Writes are validated, atomic, and `0600`.** An edit that would fail the
  Home Manager module never reaches disk, so this tool cannot be the cause of a
  broken `just switch`.
- **Output is a pure function of content.** Map keys are sorted and sections
  emptied by a removal are pruned, so no command produces a spurious diff.
- **Unknown keys are rejected** on read. A misspelled key is a setting that
  silently does nothing, which is exactly the failure mode of the
  Nix-module-shaped local config this replaced.
- **Hand editing is a first-class path.** The file is read back faithfully;
  only a command that rewrites it normalizes layout and drops comments.

Unlike [`cln`](../cln/README.md), this tool has one third-party dependency,
`github.com/pelletier/go-toml/v2`. Hand-rolling the writer would be easy, but
the file is meant to be hand-edited too, and a real parser is what turns a typo
into a message with a line and column.

## Tests

```sh
just test-local          # or: go test ./...
```

`config_test.go` covers encoding determinism, strict decoding, validation, and
machine detection; `cli_test.go` drives every subcommand end-to-end against a
temporary config directory.
