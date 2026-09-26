{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.3.2";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-n8zyzdekcsk8tU2Z2V2RTIa19BZSBU483QsXxdKDk6Y=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-aV08/T3DEZg2LwvkNE3+ZNzEZV+57zMW2UNorXrylNY=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-9W9Hdau/JpxIHHh5lpHqpZp9aWRPp9WSx3Xixl97E9o=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-jLvNS+p6e4YRahM1LzHjd4/U2T35MQNrsXcXOLBwJTQ=";
    };
  };
  source = sources.${stdenvNoCC.hostPlatform.system}
    or (throw "Unsupported system for omp: ${stdenvNoCC.hostPlatform.system}");
in
stdenvNoCC.mkDerivation {
  pname = "omp";
  inherit version;

  src = fetchurl {
    url = "https://github.com/can1357/oh-my-pi/releases/download/v${version}/${source.asset}";
    inherit (source) hash;
  };

  dontUnpack = true;

  installPhase = ''
    mkdir -p "$out/bin"
    cp "$src" "$out/bin/omp"
    chmod +x "$out/bin/omp"
  '';

  meta = {
    description = "AI coding agent for the terminal";
    homepage = "https://github.com/can1357/oh-my-pi";
    license = lib.licenses.mit;
    mainProgram = "omp";
    platforms = builtins.attrNames sources;
  };
}
