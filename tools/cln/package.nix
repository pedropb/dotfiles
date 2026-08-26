{ lib, buildGoModule }:
buildGoModule {
  pname = "cln";
  version = "0.1.0";
  src = ./.;

  # No third-party Go modules to vendor: cln is intentionally dependency-free.
  vendorHash = null;

  meta = {
    description = "Clone git repositories from short provider/namespace/repo shorthand";
    mainProgram = "cln";
    license = lib.licenses.mit;
    platforms = lib.platforms.unix;
  };
}
