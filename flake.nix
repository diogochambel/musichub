{
  description = "A dev shell that was generated";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          # Put packages here
          go
          nodejs_20
        ];

        shellHook = ''
          echo "Flake shell ready"
        '';
      };
    };
}
