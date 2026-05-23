{
  description = "A comprehensive CLI tool for Linear";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages = rec {
          linctl = pkgs.buildGoModule {
            pname = "linctl";
            inherit version;

            src = self;

            vendorHash = "sha256-Nt/V5IS0UY4ROh7epKmtAN3VDFJlCnqmKRk1AVRASgQ=";

            ldflags = [
              "-s"
              "-w"
              "-X github.com/dorkitude/linctl/cmd.version=${version}"
            ];

            meta = with pkgs.lib; {
              description = "A comprehensive CLI tool for Linear";
              homepage = "https://github.com/dorkitude/linctl";
              license = licenses.mit;
              mainProgram = "linctl";
            };
          };

          default = linctl;
        };
      }
    );
}
