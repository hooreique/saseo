{
  description = "bash-aware block management for shell rc files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];

      perSystem =
        { pkgs, ... }:
        let
          saseo = pkgs.callPackage ./package.nix { };
          devPackages = [
            pkgs.go
            pkgs.nixfmt
            pkgs.scdoc
          ];
          lintPackages = [
            pkgs.go-tools
            pkgs.gosec
          ];
        in
        {
          packages = {
            saseo = saseo;
            default = saseo;
          };

          devShells = {
            default = pkgs.mkShell {
              packages = devPackages;
            };

            lint = pkgs.mkShell {
              packages = devPackages ++ lintPackages;
            };
          };

          checks.saseo = saseo;

          formatter = pkgs.writeShellApplication {
            name = "saseo-fmt";
            runtimeInputs = [
              pkgs.go
              pkgs.nixfmt
            ];
            text = ''
              gofmt -w .
              nixfmt flake.nix package.nix
            '';
          };
        };
    };
}
