{ lib, stdenvNoCC, fetchurl }:
let
  version = "18.17.0";
  sources = {
    aarch64-darwin = {
      asset = "omp-darwin-arm64";
      hash = "sha256-UfKDbyZ6i6bBLPtQ+A518DAcz7G9U1NfxvnfAxmdo2Q=";
    };
    x86_64-darwin = {
      asset = "omp-darwin-x64";
      hash = "sha256-HZU75mQz33Jar41w4Z2VT9mC9n/QIsFkps/KbYU/HSA=";
    };
    aarch64-linux = {
      asset = "omp-linux-arm64";
      hash = "sha256-qUI/16adhpv1FJMSJrIdT/VUdCJfV04pEf7EV/xHDmg=";
    };
    x86_64-linux = {
      asset = "omp-linux-x64";
      hash = "sha256-pM5gAasQEsqDX0GE1nVDdTG9GsR3p7qEJO0LVHbp5Ms=";
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
