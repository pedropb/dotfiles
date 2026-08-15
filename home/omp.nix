{ lib, stdenvNoCC, fetchurl }:
let
  version = "17.3.4";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-dqbCL4ukujGePVKK3NkhlJ4DOPKxMEJyHmS5kPb//hY=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-zClVM8MtHl3B/r6MWXLBI15BuWFDGL7FdEHQPEdfcME=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-jifnv+SfwPM/bLC1ASirhf5UAzMNHftbs0zx90Is3Og=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-P85LJWKAZLDNe/vGJF7NraMxdQ7Us0Gspr0pukR4qrU=";
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
