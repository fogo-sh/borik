{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils/master";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      version = builtins.substring 0 8 self.lastModifiedDate;
    in
      flake-utils.lib.eachDefaultSystem (system:
        let pkgs = import nixpkgs { inherit system; };
        in
        {
          devShell = pkgs.mkShell {
            buildInputs = [
              pkgs.go
              pkgs.gotools
              pkgs.imagemagick6
              pkgs.pkg-config
            ];
            shellHook = ''
              export CGO_CFLAGS_ALLOW=-Xpreprocessor
            '';
          };

          packages.default = pkgs.buildGoModule {
            pname = "borik";
            inherit version;
            src = ./.;

            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ pkgs.imagemagick6 ];

            vendorSha256 = "sha256-TL+1hALB3iQRkitrBVXz1QuLdYadvwkcKHChrYSPD0I=";
          };

          apps.default = {
            type = "app";
            program = "${self.packages.${system}.default}/bin/borik";
          };
        });
}
