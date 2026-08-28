# `~/.config/dotfiles/local.toml`

Machine- and identity-specific configuration that shouldn't live in a public
dotfiles repository: which user the profile is built for, private employer git
hosts, org names, work email addresses.

It is data, not code, and it lives outside the checkout. The
[`dotfiles-local`](../tools/dotfiles-local/README.md) CLI maintains it; hand
editing is equally supported.

## How it's loaded

The flake declares a `local` input that defaults to the tracked stub in
[`home/local-stub`](local-stub/local.toml), and `just switch` overrides it with
your real configuration directory:

```sh
nix run .#home-manager -- switch --flake .#default \
  --override-input local "path:$(dotfiles-local path)"
```

`flake.nix` then reads it as data and hands it to `home/default.nix`:

```nix
localConfig = builtins.fromTOML (builtins.readFile "${local}/local.toml");
```

Four things follow from this:

- **Evaluation is pure.** No `builtins.getEnv`, no `builtins.currentSystem`, no
  absolute paths — so `just switch` and `just check` need no `--impure`, and
  `nix flake check` is meaningful on a fresh clone.
- **The path is `$XDG_CONFIG_HOME/dotfiles`** (`~/.config/dotfiles` when that
  variable is unset), overridable with `DOTFILES_LOCAL_DIR`. It is deliberately
  outside the repository: it survives a re-clone and can never make the work
  tree dirty or be committed by accident.
- **Edits apply immediately.** The override re-hashes the directory on every
  evaluation, so editing the file and running `just switch` is enough — no
  `git add`, no lock file update.
- **Missing is fine, up to a point.** Every section except `[machine]` is
  optional and falls back to the public defaults in `home/default.nix` (the
  `gh`/GitHub provider, `gh auth git-credential` for github.com). `[machine]`
  is required, and `dotfiles-local ensure` — which `just switch` runs for you —
  creates it from the current host.

## Schema

Keys are snake_case, matching the TOML that `cln` itself reads. Unknown keys
are rejected rather than ignored, so a typo is an error with a line number.

### `[machine]`

Required. Written by `dotfiles-local ensure`; you should rarely touch it.

```toml
[machine]
system = "x86_64-darwin"          # Nix double: {x86_64,aarch64}-{darwin,linux}
username = "example"              # becomes home.username
home_directory = "/Users/example" # becomes home.homeDirectory
```

`ensure` fills in only what is missing, so a deliberately edited value survives.
When a recorded value disagrees with the host it warns; `ensure -f` re-detects.

### `[git.identities.<name>]`

A Git author identity applied only under one or more directory prefixes, via
Git's `includeIf "gitdir:…"`. Activation writes
`~/.config/git/identities/<name>` and one `[includeIf]` block per prefix into
`~/.config/git/local`.

`gitdir` is a list, which matters when a host is reachable at more than one
directory (e.g. a repo tree that predates a renamed provider host, kept working
during a migration):

```toml
[git.identities.work]
gitdir = ["~/src/git.example.com/", "~/src/example.com/"]
name = "Example Author"
email = "author@example.com"
```

```sh
dotfiles-local identity add work \
  --author "Example Author" --email author@example.com \
  --gitdir "~/src/git.example.com/" --gitdir "~/src/example.com/"
dotfiles-local identity rm work
```

### `[git.credential_helpers]`

The Git credential helper command per HTTPS host, keyed by the host exactly as
the remote URL spells it, including a non-default port. The generated per-host
config clears the inherited helper chain first, so no platform helper (e.g.
macOS's keychain helper) answers in its place.

```toml
[git.credential_helpers]
"gitlab.example.com" = "!glab auth git-credential"
```

```sh
dotfiles-local helper set gitlab.example.com '!glab auth git-credential'
dotfiles-local helper rm gitlab.example.com
```

`github.com` is already wired to `!gh auth git-credential` by
`home/default.nix`; an entry here under the same host replaces it. See the root
README's "Git authentication" section for authenticating `gh` and `glab`
themselves — this file only wires the helper, it doesn't log in.

### `[cln]` and `[cln.providers.<alias>]`

A private [`cln`](../tools/cln/README.md) provider — a forge whose name you
don't want in the public repo — and, optionally, the default so a bare
`cln <namespace>/<repo>` reaches it instead of GitHub.

```toml
[cln]
default_provider = "corp"

[cln.providers.corp]
type = "gitlab"                # or "github", for GitHub Enterprise
host = "git.example.com"
default_namespace = "team"     # optional: enables `cln <repo>` with no namespace
```

```sh
dotfiles-local provider add corp --type gitlab --host git.example.com --default
dotfiles-local default-provider gh   # back to the built-in GitHub provider
dotfiles-local provider rm corp
```

`default_provider` must name a configured provider or the built-in `gh`; both
the CLI and a Home Manager assertion reject a dangling alias, which would
otherwise produce a `cln` config pointing at a provider that doesn't exist.

## Full example

```toml
[machine]
system = "x86_64-darwin"
username = "example"
home_directory = "/Users/example"

[git.identities.work]
gitdir = ["~/src/git.example.com/", "~/src/example.com/"]
name = "Example Author"
email = "author@example.com"

[git.credential_helpers]
"git.example.com" = "!glab auth git-credential"

[cln]
default_provider = "corp"

[cln.providers.corp]
type = "gitlab"
host = "git.example.com"
```

Run `just switch` after any change; `dotfiles-local check` validates the file
without building anything.
