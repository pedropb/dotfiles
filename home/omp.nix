{ lib, stdenvNoCC, fetchurl }:
let
  version = "17.4.2";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-NX1eegDsPTUsrF38/roVeB4eLQqQdEeSInF6e13dBAY=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-OlUgRNxBJr3mHHxHCLkjoILJ42cyzrymABeU0nx+xaE=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-pP3o+CpqIpuBW1KR3BEdtMYFMsst+EhLSsJlQRbL2/w=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-IYqGhMKxEla0fii6ExrfsqA+mI7d2FZ72Da3xR3QIAU=";
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
