{
  description = "A Nix-flake-based Go development environment";

  inputs = {
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1";
    flake-parts.url = "github:hercules-ci/flake-parts";
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem =
        { system, ... }:
        let
          goVersion = 24;
          pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [
              (final: _prev: {
                go = final."go_1_${toString goVersion}";
              })
            ];
          };
          git-hooks-check = inputs.git-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              commitizen.enable = true;
              gitleaks = {
                enable = true;
                name = "gitleaks";
                entry = "${pkgs.gitleaks}/bin/gitleaks detect --source . --verbose --redact";
                language = "system";
                pass_filenames = false;
              };
            };
          };
        in
        {
          packages.default = pkgs.buildGoModule {
            pname = "repiq";
            version = "dev";
            src = ./.;
            vendorHash = "sha256-U69tqE0QQC1kDVUa436bB2ElCIkooB7YRVtV2+EPILg=";
            meta = {
              description = "Fetch objective metrics for OSS repositories";
              homepage = "https://github.com/yutakobayashidev/repiq";
              license = pkgs.lib.licenses.mit;
              mainProgram = "repiq";
            };
          };

          checks.git-hooks = git-hooks-check;

          devShells.default = pkgs.mkShellNoCC {
            inherit (git-hooks-check) shellHook;
            packages = with pkgs; [
              go
              gotools
              golangci-lint
            ];
          };
        };
    };
}
