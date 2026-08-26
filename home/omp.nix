{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.0.5";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-mPysPtwUU9r1OIVE6ZLk3Z1DoKKZ52lZSgwN52m64qA=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-9jCRA9Y1lA6YoxfS/wM/CiAj/PDxzDESuCaVNE/rF+w=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-n6Yy2gnMbytiW7eaopHVKYX5NqEECNJ41VCYP5FGB4U=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-1aMiryQc6+JmKzt5L/KdPqbmE2QyjpFslCkGXzRjke0=";
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
