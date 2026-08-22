{
  description = "bash-aware block management for shell rc files";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { nixpkgs, ... }:
    let
      forAllSys =
        perSys:
        nixpkgs.lib.genAttrs [ "aarch64-darwin" "aarch64-linux" "x86_64-linux" ] (
          system:
          let
            pkgs = nixpkgs.legacyPackages.${system};
          in
          perSys {
            pkgs = pkgs;
            saseo = pkgs.callPackage ./package.nix { };
            dev-pkgs = [
              pkgs.go
              pkgs.nixfmt
              pkgs.scdoc
            ];
            lint-pkgs = [
              pkgs.go-tools
              pkgs.gosec
            ];
          }
        );
    in
    {
      packages = forAllSys (
        { saseo, ... }:
        {
          saseo = saseo;
          default = saseo;
        }
      );

      devShells = forAllSys (
        {
          pkgs,
          dev-pkgs,
          lint-pkgs,
          ...
        }:
        {
          default = pkgs.mkShell {
            packages = dev-pkgs;
          };

          lint = pkgs.mkShell {
            packages = dev-pkgs ++ lint-pkgs;
          };
        }
      );

      checks = forAllSys (
        { saseo, ... }:
        {
          saseo = saseo;
        }
      );

      formatter = forAllSys (
        { pkgs, ... }:
        pkgs.writeShellApplication {
          name = "saseo-fmt";
          runtimeInputs = [
            pkgs.go
            pkgs.nixfmt
          ];
          text = ''
            gofmt -w .
            nixfmt flake.nix package.nix
          '';
        }
      );
    };
}
