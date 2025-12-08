{
  description = "Age Vault - A simple age encryption tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "age-vault";
          version = "0.1.0";
          src = ./.;
          
          vendorHash = "sha256-MOwgWBUCBTW1h2I0n9VEKwcKpC1H/bD+hBjhLFyfnvg=";

          meta = with pkgs.lib; {
            description = "A simple age encryption tool";
            homepage = "https://github.com/leolimasa/age-vault";
            license = licenses.mit;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gotools
            gopls
            go-tools
            golangci-lint
            age
            sops
            openssh
          ];

          shellHook = ''
            export ENV_NAME="$ENV_NAME age-vault"
            echo "Age Vault development environment"
            echo "Go version: $(go version)"
          '';
        };
      }
    );
}
