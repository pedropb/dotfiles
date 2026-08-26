# `home/local.nix`

An optional, git-ignored Home Manager module for machine- and
identity-specific configuration that shouldn't live in a public dotfiles
repository: private employer git hosts, org names, work email addresses.

## How it's loaded

`home/default.nix` imports it conditionally:

```nix
localModule = builtins.toPath "${homeDirectory}/src/github.com/pedropb/dotfiles/home/local.nix";
imports = lib.optional (builtins.pathExists localModule) localModule;
```

Two things follow from this:

- **The path is fixed.** It must exist at `home/local.nix` inside this exact
  checkout (`~/src/github.com/pedropb/dotfiles`). There's no override for a
  different location.
- **It's read impurely, by absolute path, not through the flake's tracked
  source.** Editing it and running `just switch` picks the change up
  immediately — no `git add` needed, and it's never picked up by `git status`
  showing the flake as dirty. This is the opposite of every other file in
  this repository, which `nix build`/`switch` only sees once staged (dirty
  trees are supported, but only for indexed content).
- Missing entirely is fine: the module is simply skipped, and every option
  below falls back to its public default (e.g. `dotfiles.cln`'s built-in
  `gh`/GitHub provider).

If it doesn't exist yet, create it with:

```nix
{ ... }:
{
}
```

## Options

These are the same `dotfiles.*` options `home/default.nix` declares —
`local.nix` just sets values for them, the same as any other Home Manager
module would. Run `just switch` after any change.

### `dotfiles.git.conditionalIdentities.<name>`

A Git author identity (`name`, `email`) applied only under one or more
directory prefixes (`gitdir`), via Git's `includeIf "gitdir:…"`. Activation
writes `~/.config/git/identities/<name>` and one `[includeIf]` block per
prefix into `~/.config/git/local`.

`gitdir` accepts a single prefix or a list, when a host is reachable at more
than one directory (e.g. a repo tree that predates a later renamed
provider host, kept working during a migration):

```nix
{ ... }:
{
  dotfiles.git.conditionalIdentities.work = {
    gitdir = [ "~/src/git.example.com/" "~/src/example.com/" ];
    name = "Example Author";
    email = "author@example.com";
  };
}
```

### `dotfiles.git.credentialHelpers.<host>`

The Git credential helper command for HTTPS remotes on `host`. `host` must
exactly match the remote URL's hostname, including a non-default port. The
generated per-host config clears the inherited helper chain first, so no
platform helper (e.g. macOS's keychain helper) answers in its place.

```nix
{ ... }:
{
  dotfiles.git.credentialHelpers."gitlab.example.com" =
    "!glab auth git-credential";
}
```

See the root README's "Git authentication" section for authenticating `gh`
and `glab` themselves — `local.nix` only wires the credential helper, it
doesn't log in.

### `dotfiles.cln.defaultProvider` / `dotfiles.cln.providers.<alias>`

Add a private [`cln`](../tools/cln/README.md) provider — a git.com-alike
forge whose name you don't want in the public repo — and, optionally, make it
the default so a bare `cln <namespace>/<repo>` (no provider prefix) reaches
it instead of GitHub:

```nix
{ ... }:
{
  dotfiles.cln.defaultProvider = "corp";
  dotfiles.cln.providers.corp = {
    type = "gitlab";              # or "github" (GitHub Enterprise)
    host = "git.example.com";
    # defaultNamespace = "team";  # optional: enables `cln <repo>` with no namespace
  };
}
```

## Full example

Combining all three, for a private GitLab instance reachable at a different
hostname than an existing local checkout tree:

```nix
{ ... }:
{
  dotfiles.git.conditionalIdentities.work = {
    gitdir = [ "~/src/git.example.com/" "~/src/example.com/" ];
    name = "Example Author";
    email = "author@example.com";
  };

  dotfiles.git.credentialHelpers."git.example.com" = "!glab auth git-credential";

  dotfiles.cln.defaultProvider = "corp";
  dotfiles.cln.providers.corp = {
    type = "gitlab";
    host = "git.example.com";
  };
}
```
