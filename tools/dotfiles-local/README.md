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
  init                        first-run wizard: full-screen editor, prefilled from this host
  edit                        full-screen editor for an existing local.toml

inspect
  path                        print the config directory
  show                        print the current local.toml
  check                       validate local.toml without changing it

edit a single field (scriptable; init and edit ask for all of these too)
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
`init` and `edit` are the only interactive commands, and every prompt they ask
has a flag-driven equivalent above — nothing is reachable only by wizard.

`init` and `edit` open the same full-screen editor: a picker over every
section — `[machine]` fields, each git identity, credential helper, and `cln`
provider, plus the default provider — with a live, colorized preview of the
whole file above it, so you always see the document you're about to write,
not just the field you're touching. Selecting an existing entry offers
Edit/Remove/Back; selecting an "+ Add …" row opens a form for a new one.
Saving asks for confirmation before overwriting an existing file, and the
previous content survives at `local.toml.bak`, win or lose. `init` differs
from `edit` only in prefilling `[machine]` from this host on first run — the
same thing `ensure` does, just before instead of after.

Before the first activation the binary is not yet on `PATH`; run it through the
flake instead:

```sh
nix run .#dotfiles-local -- init
```

## Design notes

- **The config directory is outside the checkout** — `$DOTFILES_LOCAL_DIR`,
  else `$XDG_CONFIG_HOME/dotfiles`, else `~/.config/dotfiles`. It survives
  re-clones and cannot be committed by accident.
- **Writes are validated, atomic, `0600`, and backed up.** An edit that would
  fail the Home Manager module never reaches disk, so this tool cannot be the
  cause of a broken `just switch`; a write that succeeds first copies whatever
  was there to `local.toml.bak`, so a wizard run or a typo'd flag command is
  always one copy away from undone.
- **Output is a pure function of content.** Map keys are sorted and sections
  emptied by a removal are pruned, so no command produces a spurious diff.
- **Unknown keys are rejected** on read. A misspelled key is a setting that
  silently does nothing, which is exactly the failure mode of the
  Nix-module-shaped local config this replaced.
- **Hand editing is a first-class path.** The file is read back faithfully;
  only a command that rewrites it normalizes layout and drops comments.

`dotfiles-local` has two kinds of third-party dependency, for the same
reason: turning a mistake into a clear message instead of either silently
doing the wrong thing or being unpleasant enough to use that hand-editing
wins by default. `github.com/pelletier/go-toml/v2` parses and writes the
file — a real parser turns a typo into a message with a line and column.
`github.com/charmbracelet/{bubbletea,huh,lipgloss}` drive `init` and `edit` —
arrow-key pickers and validated forms instead of typing `y`/`n` and free-text
answers, in color, full-screen, redrawn in place instead of scrolling by.
Unlike [`cln`](../cln/README.md)'s zero dependencies, both are load-bearing:
hand-rolling either would either be wrong in a way nobody notices until they
hit it, or worse to use than editing the TOML by hand.

## Tests

```sh
just test-local          # or: go test ./...
```

`config_test.go` covers encoding determinism, strict decoding, validation, and
machine detection; `cli_test.go` drives every subcommand end-to-end against a
temporary config directory; `editor_test.go` covers the editor's huh-free
logic (the section list and the dirty-check). The full-screen editor itself
needs a real terminal, so — like `init` before it — it isn't unit tested;
`nix build .#dotfiles-local` and a manual run are how it gets exercised.
