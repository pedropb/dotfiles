_scd_component_matches() {
  emulate -L zsh
  setopt extended_glob

  local needle=$1
  local haystack=$2
  local pattern=$needle[1]
  local index

  for (( index = 2; index <= $#needle; index++ )); do
    pattern+="*$needle[index]"
  done

  [[ $haystack = (#i)$~pattern* ]]
}

_scd_repositories() {
  emulate -L zsh
  setopt null_glob

  local root=$HOME/src
  local query=$1
  local relative
  local path
  local count
  local index
  local start
  local -a components parts

  IFS=/ read -rA parts <<< "$query"
  count=$#parts
  (( count > 0 && count <= 3 )) || return
  start=$(( 4 - count ))

  for path in "$root"/*/*/*(N/); do
    relative=${path#$root/}
    IFS=/ read -rA components <<< "$relative"

    for (( index = 1; index <= count; index++ )); do
      _scd_component_matches "$parts[index]" "$components[start + index - 1]" || break
    done

    (( index > count )) && print -r -- "$relative"
  done
}

scd() {
  emulate -L zsh

  local root=$HOME/src
  local -a matches

  if (( $# == 0 )); then
    builtin cd -- "$root"
    return
  fi

  if (( $# != 1 )); then
    print -u2 -- "usage: scd [repository]"
    return 2
  fi

  if [[ -d $1 ]]; then
    builtin cd -- "$1"
    return
  fi

  if [[ -d "$root/$1" ]]; then
    builtin cd -- "$root/$1"
    return
  fi

  matches=("${(@f)$(_scd_repositories "$1")}")
  (( $#matches == 1 )) && [[ -z $matches[1] ]] && matches=()
  case $#matches in
    0)
      print -u2 -- "scd: no repository matches '$1'"
      return 1
      ;;
    1)
      builtin cd -- "$root/$matches[1]"
      ;;
    *)
      print -u2 -- "scd: ambiguous repository '$1':"
      print -u2 -l -- "$matches[@]"
      return 1
      ;;
  esac
}

_scd_complete_children() {
  emulate -L zsh
  setopt null_glob

  local root=$HOME/src
  local query=$1
  local directory
  local parent
  local fragment
  local child
  local -a candidates

  if [[ $query == */* ]]; then
    parent=${query%/*}
    fragment=${query##*/}
    directory="$root/$parent"
  else
    parent=
    fragment=$query
    directory=$root
  fi

  [[ -d $directory ]] || return 1

  for child in "$directory"/*(N/); do
    [[ -z $fragment ]] || _scd_component_matches "$fragment" "${child:t}" || continue
    candidates+=("${parent:+$parent/}${child:t}")
  done

  (( $#candidates )) || return 1
  compadd -U -S / -- "$candidates[@]"
}

_scd() {
  _scd_complete_children "$words[CURRENT]" && return

  local -a matches

  matches=("${(@f)$(_scd_repositories "$words[CURRENT]")}")
  (( $#matches == 1 )) && [[ -z $matches[1] ]] && matches=()
  (( $#matches )) && compadd -U -- "$matches[@]"
}

compdef _scd scd
