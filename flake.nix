{
  description = "Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go runtime - choose your version
            go

            # Language server & tools
            gopls # LSP for eglot
            delve # Debugger for dape

            # Formatters & linters
            gofumpt # Formatter
            gotools # goimports, etc.
            golangci-lint # Meta-linter

            # Additional Go tools
            gomodifytags # Struct tag manipulation
            gore # Go REPL
            gotests # Generate tests
            impl # Generate interface implementations
            netcat-openbsd
            # Project-specific tools (uncomment as needed)
            # goose # DB migrations
            # sqlc  # SQL code generator
            # air   # Live reload
          ];

          shellHook = ''
            echo "🐹 Go development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "✓ gopls (LSP)"
            echo "✓ delve (debugger)"
            echo "✓ gofumpt (formatter)"
            echo "✓ golangci-lint (linter)"
            echo ""
            echo "Available commands:"
            echo "  go build       - Build the project"
            echo "  go test        - Run tests"
            echo "  go run .       - Run the project"
            echo "  golangci-lint run - Lint code"

            export GOPATH="$HOME/go"
            export GOBIN="$HOME/go/bin"
            export GOMODCACHE="$HOME/go/pkg/mod"
            export PATH="$HOME/go/bin:$PATH"
          '';

          # Environment variables
          env = {
            CGO_ENABLED = "1";
          };
        };
      }
    );
}
