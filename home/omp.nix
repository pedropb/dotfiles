{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.0.8";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-pWrcR/L08GJEHTMWYiw3ken//HyZaFTk9Z5DNTvZHH8=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-AlgfI85LcQBgbYOYcbP1AZ3VRCPz/s0LFnDQClg5UZI=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-kWl0/dZC080pTLLBddIGbeShlVTGFkNiAj2KloEcAfc=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-sVxxYqPMdImMKs52UkFYpKJOQ/nQhlAefacanzoYX4A=";
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
