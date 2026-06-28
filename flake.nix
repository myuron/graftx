{
  description = "flake for Go with Agent-skills";

  inputs = {
    nixpkgs = {
      url = "https://flakehub.com/f/NixOS/nixpkgs/0.1";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
    agent-skills = {
      url = "github:Kyure-A/agent-skills-nix";
    };
    anthropic-skills = {
      url = "github:anthropics/skills";
      flake = false;
    };
  };

  outputs = { nixpkgs, flake-utils, agent-skills, anthropic-skills, ... }:
  flake-utils.lib.eachDefaultSystem (
    system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
      agentLib = agent-skills.lib.agent-skills;
      sources = {
        anthropic = {
          path = anthropic-skills;
          subdir = "skills";
        };
      };
      catalog = agentLib.discoverCatalog sources;
      allowlist = agentLib.allowlistFor {
        inherit catalog sources;
        # Add Agent Skills
        enable = [
          "doc-coauthoring"
          "skill-creator"
        ];
      };
      selection = agentLib.selectSkills {
        inherit catalog allowlist sources;
        skills = { };
      };
      bundle = agentLib.mkBundle { inherit pkgs selection; };
      localTargets = {
        claude = agentLib.defaultLocalTargets.claude // { enable = true; };
      };
    in
    {
      packages.default = pkgs.buildGoModule {
        pname = "graftx";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-TUbaUoqDZoQTkcOMtoE/FlAiqkWN+x49JeGkDguh2UU=";
        # サンドボックス環境ではホームディレクトリが存在しないためテストをスキップ
        doCheck = false;
      };
      apps = {
        skills-install-local = {
          type = "app";
          program = "${agentLib.mkLocalInstallScript {inherit pkgs bundle; targets = localTargets; }}/bin/skills-install-local";
        };
      };
      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          uv
          go
          gotools
          golangci-lint

        ];
      };
    }
  );
}
