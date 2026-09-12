{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.1.18";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-A1o13LJJ7bk5+gK3T8fA35sutlkHk0n7/7GI4bVYlXo=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-Kf6tG2Z9yWm4JcWqR5iq21t+ybLxXDMEsQGW16+kFx0=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-GugnPCMc64jOvJlxkBz39dl+1BSc3HQKgU9inNK03LI=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-RUIfml8RK8R8ufd8S015J2Mfj/hZYk+CEofshU6yOfw=";
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
