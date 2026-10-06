#!/usr/bin/env bash
#
# Report the size of the activated Home Manager profile and attribute it to the
# dependencies this repo declares in home/packages.nix.
#
# Sizes are uncompressed NAR sizes. A binary cache serves them compressed, so an
# actual download is typically 40-60% of the figures here.
#
# Attribution is reference-closed: a path counts as a dependency's EXCLUSIVE
# size only when no other package installed in the profile reaches it. That is
# the size dropping (or moving out of Nix) the dependency would reclaim; paths
# something else still needs are reported as SHARED instead.

set -euo pipefail

usage() {
  cat <<'EOF'
Usage: profile-size.sh [--json]

Report the size of the activated Home Manager profile, attributed to the
dependencies declared in home/packages.nix.

  --json   machine-readable output for further analysis
  -h       this help

Columns: FULL is a dependency's whole runtime closure, EXCLUSIVE is the part
nothing else in the profile needs (what dropping it reclaims), SHARED is the
remainder. Sizes are uncompressed NAR bytes.
EOF
}

mode=table
case "${1:-}" in
  "") ;;
  --json) mode=json ;;
  -h | --help)
    usage
    exit 0
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
if [ "$#" -gt 1 ]; then
  usage >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

if [ ! -e "$HOME/.nix-profile" ]; then
  echo "profile-size: no Home Manager profile at ~/.nix-profile; run 'just switch' first" >&2
  exit 1
fi
profile="$(readlink -f "$HOME/.nix-profile")"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Evaluate the same way bin/check.sh does, so the dependency list and the pkgs
# behind it are the ones the activated profile was built from.
local_dir="$(nix run .#dotfiles-local -- ensure)"
eval_opts=(--override-input local "path:$local_dir")
system="$(nix eval --raw .#homeConfigurations.default.config.nixpkgs.system "${eval_opts[@]}")"

# name<TAB>outPath for both the declared inventory and everything the profile
# installs, so the difference is visible as well as the dependencies themselves.
emit='xs: builtins.concatStringsSep "\n" (map (x: "${x.pname or x.name}\t${x.outPath}") xs)'
nix eval --raw ".#legacyPackages.${system}.repoPackages" "${eval_opts[@]}" --apply "$emit" >"$tmp/declared.tsv"
nix eval --raw .#homeConfigurations.default.config.home.packages "${eval_opts[@]}" --apply "$emit" >"$tmp/installed.tsv"

# One closure per package. Duplicate entries collapse, declared entries win, and
# a package that is not in the store yet is reported rather than built.
: >"$tmp/index.tsv"
: >"$tmp/roots.seen"
n=0
skipped=0
for list in declared installed; do
  declared=$([ "$list" = declared ] && echo 1 || echo 0)
  while IFS="$(printf '\t')" read -r name root || [ -n "${name:-}" ]; do
    [ -n "${root:-}" ] || continue
    if grep -Fxq -- "$root" "$tmp/roots.seen"; then continue; fi
    printf '%s\n' "$root" >>"$tmp/roots.seen"
    if nix path-info --recursive "$root" >"$tmp/p.candidate" 2>/dev/null; then
      n=$((n + 1))
      mv "$tmp/p.candidate" "$tmp/p.$n"
      printf '%s\t%s\t%s\t%s\n' "$n" "$declared" "$name" "$root" >>"$tmp/index.tsv"
    else
      rm -f "$tmp/p.candidate"
      skipped=$((skipped + 1))
      echo "profile-size: $name ($root) is not in the store; skipping" >&2
    fi
  done <"$tmp/$list.tsv"
done

: >"$tmp/membership.tsv"
i=1
while [ "$i" -le "$n" ]; do
  while IFS= read -r path; do printf '%s\t%s\n' "$i" "$path"; done <"$tmp/p.$i" >>"$tmp/membership.tsv"
  i=$((i + 1))
done

nix path-info --recursive "$profile" | sort -u >"$tmp/profile.paths"
cat "$tmp/profile.paths" "$tmp"/p.* | sort -u >"$tmp/all.paths"
xargs -n 200 nix path-info --size <"$tmp/all.paths" >"$tmp/sizes.tsv"

awk -F'\t' -v mode="$mode" -v szf="$tmp/sizes.tsv" -v idxf="$tmp/index.tsv" \
  -v prf="$tmp/profile.paths" -v memberf="$tmp/membership.tsv" \
  -v profile="$profile" -v host="$system" -v skipped="$skipped" '
function human(bytes,   unit, i) {
  split("B KiB MiB GiB TiB", unit, " ")
  i = 1
  while (bytes >= 1024 && i < 5) { bytes /= 1024; i++ }
  return sprintf(i == 1 ? "%d %s" : "%.1f %s", bytes, unit[i])
}
function row(i) {
  printf "%-22s %11s %12s %11s %7d %6.1f%%\n", name[i], human(full[i]), human(exclusive[i]),
    human(full[i] - exclusive[i]), paths[i], (total ? 100 * exclusive[i] / total : 0)
}
BEGIN {
  while ((getline line < szf) > 0) { split(line, a, "[ \t]+"); size[a[1]] = a[2] + 0 }
  while ((getline line < idxf) > 0) {
    nsrc++
    split(line, a, "\t")
    id[nsrc] = a[1]; declared[a[1]] = a[2] + 0; name[a[1]] = a[3]; root[a[1]] = a[4]
  }
  while ((getline line < prf) > 0) { total += size[line]; nprofile++ }
}
{
  src = $1; path = $2
  if (!seen[src, path]++) { reaches[path]++; full[src] += size[path]; paths[src]++ }
}
END {
  # Second pass over the membership list: a path that only one package reaches
  # is exclusive to that package.
  close(memberf)
  while ((getline line < memberf) > 0) {
    split(line, a, "\t")
    if (reaches[a[2]] == 1) exclusive[a[1]] += size[a[2]]
  }
  for (path in reaches) {
    covered += size[path]; ncovered++
    if (reaches[path] > 1) shared += size[path]
  }
  # Profile content no package reaches: the profile directory itself, and nix-env
  # bookkeeping such as env-manifest.nix.
  bookkeeping = total - covered
  nbookkeeping = nprofile - ncovered

  nd = 0; no = 0
  for (k = 1; k <= nsrc; k++) {
    i = id[k]
    if (declared[i]) { nd++; did[nd] = i } else { no++; oid[no] = i }
  }
  for (k = 1; k <= nd; k++)
    for (j = k + 1; j <= nd; j++)
      if (exclusive[did[j]] > exclusive[did[k]]) {
        swap = did[k]; did[k] = did[j]; did[j] = swap
      }
  for (k = 1; k <= no; k++)
    for (j = k + 1; j <= no; j++)
      if (exclusive[oid[j]] > exclusive[oid[k]]) {
        swap = oid[k]; oid[k] = oid[j]; oid[j] = swap
      }
  for (k = 1; k <= nd; k++) {
    d_full += full[did[k]]; d_excl += exclusive[did[k]]; d_paths += paths[did[k]]
  }
  for (k = 1; k <= no; k++) {
    o_full += full[oid[k]]; o_excl += exclusive[oid[k]]; o_paths += paths[oid[k]]
  }

  if (mode == "json") {
    printf "{\n"
    printf "  \"profile\": {\"path\": \"%s\", \"system\": \"%s\", \"total_bytes\": %d, \"paths\": %d, \"bookkeeping_bytes\": %d, \"bookkeeping_paths\": %d},\n",
      profile, host, total, nprofile, bookkeeping, nbookkeeping
    printf "  \"declared\": {\"count\": %d, \"skipped\": %d, \"closure_bytes\": %d, \"exclusive_bytes\": %d, \"paths\": %d},\n",
      nd, skipped, d_full, d_excl, d_paths
    printf "  \"module_added\": {\"count\": %d, \"closure_bytes\": %d, \"exclusive_bytes\": %d, \"paths\": %d},\n",
      no, o_full, o_excl, o_paths
    printf "  \"shared_bytes\": %d,\n", shared
    printf "  \"dependencies\": [\n"
    for (k = 1; k <= nd + no; k++) {
      i = (k <= nd ? did[k] : oid[k - nd])
      printf "    {\"name\": \"%s\", \"root\": \"%s\", \"declared\": %s, \"closure_bytes\": %d, \"exclusive_bytes\": %d, \"shared_bytes\": %d, \"paths\": %d}%s\n",
        name[i], root[i], (declared[i] ? "true" : "false"), full[i], exclusive[i],
        full[i] - exclusive[i], paths[i], (k < nd + no ? "," : "")
    }
    printf "  ]\n}\n"
    exit
  }

  printf "\nprofile  %s\n", profile
  printf "total    %s in %d store paths (%s)\n\n", human(total), nprofile, host

  printf "declared in home/packages.nix\n"
  printf "%-22s %11s %12s %11s %7s %7s\n", "DEPENDENCY", "FULL", "EXCLUSIVE", "SHARED", "PATHS", "%PROF"
  for (k = 1; k <= nd; k++) row(did[k])
  printf "%-22s %11s %12s %11s %7d %6.1f%%\n", "declared total", human(d_full), human(d_excl),
    human(d_full - d_excl), d_paths, (total ? 100 * d_excl / total : 0)

  if (no) {
    printf "\ninstalled by home-manager modules, not declared in the repo\n"
    printf "%-22s %11s %12s %11s %7s %7s\n", "PACKAGE", "FULL", "EXCLUSIVE", "SHARED", "PATHS", "%PROF"
    for (k = 1; k <= no; k++) row(oid[k])
    printf "%-22s %11s %12s %11s %7d %6.1f%%\n", "other total", human(o_full), human(o_excl),
      human(o_full - o_excl), o_paths, (total ? 100 * o_excl / total : 0)
  }

  printf "\n%-42s %s\n", "reclaimable by dropping the declared deps", human(d_excl)
  if (no) printf "%-42s %s\n", "reclaimable by dropping the module-added ones", human(o_excl)
  printf "%-42s %s\n", "shared by two or more packages", human(shared)
  printf "%-42s %s in %d paths\n", "profile bookkeeping (manifest, profile dir)", human(bookkeeping), nbookkeeping
  if (skipped) printf "%-42s %d deps\n", "skipped (not in the store)", skipped
  if (covered + bookkeeping != total) printf "%-42s %s\n", "MISMATCH, unattributed paths", human(total - covered - bookkeeping)
  printf "\n"
}
' "$tmp/membership.tsv"