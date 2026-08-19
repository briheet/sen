{
  description = "sen: Source level runtime and analysis visualization";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  };

  outputs = inputs: {
    devShells = builtins.mapAttrs (system: pkgs: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
        ];
      };
    }) inputs.nixpkgs.legacyPackages;
  };
}
