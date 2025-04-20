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
    {
      overlay = _: prev: let
        pkgs = nixpkgs.legacyPackages.${prev.system};
      in {
        neko = pkgs.buildGoModule {
          pname = "neko";
          version = "0.0.4";
          src = pkgs.lib.cleanSource self;

          vendorHash = "sha256-Qlc3nAAvkq/XaWaBHO9cXVOQPYoB4230ViINVZYQwZU=";

          subPackages = ["cmd/server"];
        };
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          overlays = [self.overlay];
          inherit system;
        };
      in rec {
        packages = with pkgs; {
          inherit neko;
          default = neko;
        };
      }
    );
}
