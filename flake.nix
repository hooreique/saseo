{
  description = "saseo adds and removes small rc snippets that should stick around for a while, but not forever.";

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
        in
        {
          packages = {
            saseo = saseo;
            default = saseo;
          };

          devShells.default = pkgs.mkShell {
            packages = [
              pkgs.nushell
              pkgs.nufmt
              pkgs.scdoc
            ];
          };

          checks.saseo = saseo;

          formatter = pkgs.nufmt;
        };
    };
}
