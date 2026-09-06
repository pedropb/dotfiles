{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.1.11";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-qBsqmNmdWzRJHSUjICDVPLDlUmCV7yHEy51Cg+ZRGAc=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-qAdkXRFEcKHirXKzZG8RhTIUGjaw3gYkeP75majD+sM=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-IN1slG3LYeL3nfg/7upROufuk8cRkNHLjn+BqC1eXhM=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-Kyx4W7uv07Q/tgluSuY6A0LE2Aad4vh1qbduqZR8I4M=";
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
