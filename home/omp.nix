{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.8.3";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-RCFTi2qYhSfu3wjsMqqD4n7DEiej+RAo2PWcED5Te2A=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-KqXjX7LFsawaI278skWmDlM70guVrcFIecFmwM/xLN8=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-FwhSt4Fe904SEvh+afp/JzDzFxZHkXffQ20+qoMtjKU=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-jLbWoANaPF2cQKWIUso2FWnq/Czsu+70Uh4P78jMvDc=";
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
