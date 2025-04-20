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

      nixosModules.default = {
        pkgs,
        lib,
        config,
        ...
      }: let
        inherit (lib) mkIf;

        cfg = config.services.neko;
        settingsFormat = pkgs.formats.yaml {};
        configFile = settingsFormat.generate "config.yaml" cfg.settings;
      in {
        options = {
          services.neko = {
            enable = lib.mkEnableOption "";
            package = lib.mkOption {
              type = lib.types.package;
              default = pkgs.neko;
              description = "";
            };
            settings = lib.mkOption {
              type = settingsFormat.type;
              default = {
                listenAddress = "127.0.0.1:8300";
                handlers = [];
              };
              description = ''
              '';
            };
          };
        };
        config = mkIf cfg.enable {
          nixpkgs.overlays = [self.overlay];
          systemd.services.neko = {
            wantedBy = ["multi-user.target"];
            after = ["network.target"];
            description = "";

            environment = {
              NEKO_CONFIG = configFile;
            };

            serviceConfig = {
              Restart = "always";
              DynamicUser = "yes";
              ExecStart = "${cfg.package}/bin/server";
              AmbientCapabilities = ["CAP_NET_RAW"];
              CapabilityBoundingSet = ["CAP_NET_RAW"];
            };
          };
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
