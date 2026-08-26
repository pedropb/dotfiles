# cln

Clone a git repository by provider/namespace/repo shorthand instead of typing
the full HTTPS remote URL.

```sh
cln gh dotfiles              # -> git clone https://github.com/pedropb/dotfiles
cln octocat/Hello-World      # -> git clone https://github.com/octocat/Hello-World
cln team/billing             # -> fuzzy-matches "billing" against the default
                              #    provider's "team" namespace, e.g. to
                              #    team/billing-service
cln providers                # list configured providers
```

`cln [-n] [provider] <[namespace/]repo> [dest]`. `provider` is only read as a
provider when it names one configured below; otherwise a 2-argument call is
`repo dest`, matching `git clone <repo> [dir]`. `-n`/`--dry-run` prints the
resolved `git clone` command instead of running it.

Without an explicit `dest`, `cln` clones into
`~/src/<host>/<namespace>/<repo>` — the layout `config/zsh/scd.zsh` expects
under `$HOME/src` — creating any missing parent directories first. A repo
cloned with `cln` is immediately reachable with `scd <namespace>/<repo>`; an
explicit `dest` argument always overrides this.

## Configuration

`cln` reads `$XDG_CONFIG_HOME/cln/config.toml` (`~/.config/cln/config.toml`
by default). With no file at all it falls back to a single `gh` provider for
`github.com` with no default namespace, so `cln <owner>/<repo>` always works
out of the box.

```toml
default_provider = "gh"

[providers.gh]
type = "github"
host = "github.com"
default_namespace = "pedropb"

[providers.corp]
type = "gitlab"
host = "git.example.com"
```

- `default_provider`: alias used when a command omits one.
- `[providers.<alias>]`: one block per backend.
  - `type`: `"github"` or `"gitlab"` — picks how `cln` builds URLs and lists
    repositories (see below).
  - `host`: the hostname repositories are cloned from.
  - `default_namespace` (optional): namespace used when a repository is
    given without one, e.g. your personal account or a team you clone from
    daily.

In this repository, `home/default.nix` renders this file from the
`dotfiles.cln.defaultProvider`/`dotfiles.cln.providers` Nix options. Add
private, non-public forges to the git-ignored `home/local.nix` rather than
here — see [`home/local.nix.md`](../../home/local.nix.md) for the option
and an example.

The parser only understands this subset of TOML (flat root keys plus one
level of `[providers.<alias>]` tables of string values) — see the doc
comment on `parseConfig` in `config.go`. That keeps `cln` dependency-free;
the file itself is still valid TOML.

## Fuzzy matching

When a repository name doesn't need to be exact, `cln` lists the target
namespace's repositories and ranks candidates in tiers: exact match, prefix,
substring, then subsequence (the same family of match `fzf` uses). It picks
the sole winner in the best non-empty tier. Multiple winners in that tier are
reported as an ambiguous-repository error rather than guessed at — the same
behavior `config/zsh/scd.zsh` uses for ambiguous directory matches.

This match logic is a few dozen lines in `match.go`, not a dependency:
`cln` intentionally has zero third-party Go modules (`go.mod` has no
`require`s), which keeps building and packaging it (see below) trivial and
fully offline.

If a namespace's repository listing can't be retrieved — the provider CLI
isn't installed or authenticated, there's no network, etc. — or the listing
is larger than `maxCatalogSize` (300), fuzzy matching is skipped entirely and
the repository name is used exactly as given; `git clone` then reports an
unknown repository itself if it doesn't exist. Listings are cached under
`$XDG_CACHE_HOME/cln` for an hour; a failed refresh falls back to whatever is
cached, however stale, before giving up on fuzzy matching.

## Authentication

`cln` never handles credentials itself. The actual `git clone` is
authenticated by your global git credential helper — in this repository,
`home/default.nix` and `home/local.nix` already wire `gh auth git-credential`
/ `glab auth git-credential` per host (see the root README's
"Git authentication" section).

Listing a namespace's repositories for fuzzy matching does need an
authenticated call, so `cln` shells out to the provider's own CLI, reusing
whatever credentials it already has stored:

| `type`   | lists repositories via                                     |
| -------- | ------------------------------------------------------------ |
| `github` | `gh repo list <namespace> --json name --jq '.[].name'`       |
| `gitlab` | `glab api "groups/<namespace>/projects" --hostname <host>`   |

### Extending to another provider

Adding a backend (Bitbucket, a self-hosted Gitea, ...) needs no plugin
system, just:

1. A `fetchXCatalog(p Provider, namespace string) ([]string, error)` in
   `provider.go`, following the existing two: shell out to that backend's
   authenticated CLI, return repository slugs, and fail (don't guess) past
   `maxCatalogSize`.
2. A case for it in `fetchCatalog`'s switch statement.
3. Its type string added to the check in `Config.validate` (`config.go`).

`cloneURL` and the namespace-splitting in `resolve.go` are already generic
(`https://<host>/<namespace>/<repo>`); only nested namespaces are
type-gated, since GitHub owners can't nest.

## Build, test, install

```sh
go build ./...
go test ./...
```

`package.nix` is a `buildGoModule` derivation (`vendorHash = null`, since
there are no third-party dependencies to vendor); `home/packages.nix` adds it
to the profile. Outside this repository, `go install
github.com/pedropb/dotfiles/tools/cln@latest`, or build the module directly,
work the same way.
