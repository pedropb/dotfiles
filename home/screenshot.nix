{ pkgs, lib, ... }:

# Flameshot-based screenshot capture + annotation, bound to a hotkey that
# mirrors macOS's native ⌘⇧4 (shifted by ⌃ so it doesn't collide with the
# system shortcut). Darwin-only: skhd is a launchd-managed daemon and
# home-manager's services.skhd module asserts the darwin platform.
#
# Once the editor is open, press ⌘A to select the whole screen if you want
# a full-screen capture instead of dragging a region — no separate hotkey
# needed for that.
let
  # Confirmed upstream macOS bug, reproduced directly (not assumed):
  # flameshot's own clipboard copy (`-c`, the in-app Copy button, ⌘C) is
  # broken on macOS when launched as a background/CLI process — even with
  # the window explicitly activated and focused, ⌘C left the system
  # clipboard completely empty. Matches long-standing, still-open upstream
  # reports (flameshot-org/flameshot issues #2567, #2879, #3963, #4630).
  # `-p` (save to file) is reliable.
  #
  # Workaround: launch un-focused (skhd/launchd grants no app activation, so
  # we do it ourselves), let the user annotate and accept normally (saving
  # to ~/Pictures/Screenshots via -p, which works), then once flameshot
  # exits, push whatever new file it wrote onto the clipboard ourselves via
  # osascript — a completely independent code path from flameshot's,
  # verified end-to-end with a real `pngpaste` roundtrip.
  flameshotCapture = pkgs.writeShellScript "flameshot-capture" ''
    set -eu
    DIR="$HOME/Pictures/Screenshots"
    mkdir -p "$DIR"
    before=$(ls -t "$DIR"/*.png 2>/dev/null | head -1 || true)

    "${pkgs.flameshot}/bin/flameshot" gui -p "$DIR" &
    fpid=$!

    # Poll for real activation instead of a blind sleep: launchd/skhd grant
    # no app activation, and a fixed delay was racy (flameshot sometimes
    # hadn't finished creating/registering its overlay window yet).
    /usr/bin/osascript -e "
      tell application \"System Events\"
        repeat 40 times
          if exists (first process whose unix id is $fpid) then exit repeat
          delay 0.05
        end repeat
        set frontmost of (first process whose unix id is $fpid) to true
        repeat 40 times
          if frontmost of (first process whose unix id is $fpid) then exit repeat
          delay 0.05
        end repeat
      end tell
    " >/dev/null 2>&1 || true

    wait "$fpid" 2>/dev/null || true

    after=$(ls -t "$DIR"/*.png 2>/dev/null | head -1 || true)
    if [ -n "$after" ] && [ "$after" != "$before" ]; then
      /usr/bin/osascript -e "set the clipboard to (read (POSIX file \"$after\") as «class PNGf»)"
    fi
  '';
in
lib.mkIf pkgs.stdenv.isDarwin {
  services.skhd = {
    enable = true;
    config = ''
      # ⌘⌃⇧4 — region select (⌘A inside the editor selects the whole
      # screen), opens the annotator, lands on disk + clipboard on accept
      cmd + ctrl + shift - 4 : ${flameshotCapture}
    '';
  };

  home.activation.createScreenshotsDirectory =
    lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      $DRY_RUN_CMD mkdir -p "$HOME/Pictures/Screenshots"
    '';
}
