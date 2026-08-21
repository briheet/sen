{
  description = "sen: Source level runtime and analysis visualization";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

    systems = {
      url = "github:nix-systems/default";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      systems,
    }:
    let
      inherit (nixpkgs) lib legacyPackages;

      forAllSystems = lib.genAttrs (import systems);
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = legacyPackages.${system};
        in
        rec {
          sen = pkgs.buildGoModule {
            pname = "sen";
            version = "0.1.0";

            src = self;
            vendorHash = "sha256-vk4V6gOxhK8KhrgELN79azbnHSd7naQUtfDUSLsNEhQ=";

            subPackages = [ "cmd/sen" ];
            ldflags = [
              "-s"
              "-w"
            ];

            meta = {
              description = "Source-level runtime and analysis visualization";
              homepage = "https://github.com/briheet/sen";
              license = lib.licenses.mit;
              mainProgram = "sen";
              platforms = lib.platforms.all;
            };
          };

          default = sen;
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = lib.getExe self.packages.${system}.default;
          meta.description = "Run sen";
        };
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.redis
              pkgs.postgresql
            ];
          };
        }
      );

      formatter = forAllSystems (system: legacyPackages.${system}.nixfmt);
    };
}
