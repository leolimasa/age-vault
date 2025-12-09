{
  description = "Age Vault - Share secrets across machines. Built on top of the `age` encryption tool.";

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
          
          vendorHash = "sha256-RLRvYay5nuTWSAN0qvSpgzkZhDHNGMi49m3JwtUNw0s=";

          meta = with pkgs.lib; {
            description = "Share secrets across machines. Built on top of the `age` encryption tool.";
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
