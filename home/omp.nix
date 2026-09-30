{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.4.4";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-524CghJC+zaERnap37efvl+wadk+879iYl7DOf6dCSs=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-GFQf0Vp3BxlakEG3SPKx1kfzlRStINnEHfUwRCRN6rU=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-YC7v3cD9hwQ/jwjYYo1yWS5YAsADoF8gPx7LxjqP3TA=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-JMgw/OsL1ohL9b8seit0B7wj+v5lXpJMaV75vjCORvM=";
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
