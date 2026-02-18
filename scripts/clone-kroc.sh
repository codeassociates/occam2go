#!/bin/bash
# Clone the KRoC (Kent Retargetable occam Compiler) repository.
# This provides the occam "course" standard library source code
# needed for transpiling programs that use it.

set -e

REPO_URL="https://github.com/concurrency/kroc.git"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TARGET_DIR="$PROJECT_DIR/kroc"

if [ -d "$TARGET_DIR" ]; then
    echo "kroc/ already exists. To re-clone, remove it first:"
    echo "  rm -rf $TARGET_DIR"
    exit 1
fi

echo "Cloning KRoC repository into kroc/..."
git clone "$REPO_URL" "$TARGET_DIR"

# The KRoC repo contains .go files (in tests/ccsp-comparisons/go/) that fail
# to compile. Adding a go.mod creates a module boundary so that "go build ./..."
# and "go test ./..." from the project root skip the kroc/ directory entirely.
cat > "$TARGET_DIR/go.mod" <<'EOF'
module kroc-vendored

go 1.25.6
EOF

echo "Done."
