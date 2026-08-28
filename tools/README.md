# tools

Standalone programs developed alongside this dotfiles profile but versioned
and tested independently of it — each is its own module with its own build,
tests, and Nix package, wired into the profile from `home/packages.nix`.

| tool                                         | what                                                       |
| -------------------------------------------- | ---------------------------------------------------------- |
| [`cln`](cln/README.md)                       | clone a repo by provider/namespace/repo shorthand          |
| [`dotfiles-local`](dotfiles-local/README.md) | maintain the private, machine-specific half of the profile |

Adding another tool: give it its own directory here with its own module/build
file and a `package.nix`, then reference that `package.nix` from
`home/packages.nix` the way `cln`'s is.
