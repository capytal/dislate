{
  description = "learning.rs";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    templ.url = "github:a-h/templ?ref=v0.2.707";
  };
  outputs = { nixpkgs, ... } @ inputs:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        f system pkgs);
      templ = system: inputs.templ.packages.${system}.templ;
    in
    {
      devShells = forAllSystems (system: pkgs: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            docker-compose
            (templ system)
          ];
        };
      });
    };
}
