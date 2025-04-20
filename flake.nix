{
  description = "An(other) uptime monitor.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }: let
  in
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        packages.default = pkgs.buildGoModule {
          pname = "neko";
          version = "0.0.4";
          src = pkgs.lib.cleanSource self;

          vendorHash = "sha256-Qlc3nAAvkq/XaWaBHO9cXVOQPYoB4230ViINVZYQwZU=";

          subPackages = ["cmd/server"];
        };
      }
    );
}
