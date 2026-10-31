{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.14.2";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-m3vaC9jsjR/Y+N7aB2GA3u8tSSHuYOfS85BE7n2UN+4=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-YoNeUazA8H0ZST2nNrchXTIk/VCZdUt+QsOw/cLrf0U=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-r73wS+PdvjSgfpHJE9220zWjoCEPUgrFoZeiRYiK1j4=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-zeC8SsjsMzxIgAOaUkL/M9v8ZUR0fWeuuZuwd4dK7sE=";
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
