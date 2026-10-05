{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.6.1";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-tcpc0XuMwJ7ONoRZiP1AH30QAuGrW5VgDg8yiSQkbVI=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-TIyl+N/pgHb7aIj3ZGWkO+VQOrOhg8viDMF0j51bm6g=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-y3gVMwuxF4d+ThM1YuqC4wAe5HDkc5NVJQe1p/BAf0o=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-ySpoRtAphOhPB8Y2LRit03g/Uo9JT/zx4PJuWU8ydGM=";
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
