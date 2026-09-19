{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.2.6";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-1JjaQNV34f+mgcqGMsLqQKn3IqCLiAQSAR033/7pUTo=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-1ViAD6Mmq8aK48u7PlNjIuxkO9z93HHmqYiFfb4alKM=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-ByRcvgUMOZmrXOqbq/6E5+iBnS9NXkm+9HwKrLa5V+Q=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-DzhZjJHoI9jM4H8VHsOZnVHyE/ssw+B9ifGvju+SR6I=";
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
