#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR"
SPRINT=3
GROUP="05"
TP="21"
OUTPUT="$SCRIPT_DIR/PSI_S${SPRINT}_TP${TP}_G${GROUP}.ZIP"

rm -f "$OUTPUT"

cd "$PROJECT_DIR"

7z a -tzip "$OUTPUT" . \
  -xr!.opencode \
  -xr!.git \
  -xr!node_modules \
  -xr!AGENTS.md \
  -xr!UNIVERSE.md \
  -xr!STORIES.md \
  -xr!hey \
  -xr!user-stories.txt \
  -xr!DESIGN.md \
  -xr!zip.sh \
  -xr!UNIVERSE.md \
  -xr!TODO.md \
  -xr!.go \
  -xr!bin \
  -xr!frontend/dist \
  -xr!frontend/.angular \
  -xr!frontend/.nx \
  -xr!.env \
  -xr!.editorconfig \
  -xr!.gitignore \
  -xr!.env.local \
  -xr!"*.log" \
  -xr!"*.exe" \
  -xr!"*.dll" \
  -xr!"*.so" \
  -xr!"*.dylib" \
  -xr!"*.out" \
  -xr!"*.test" \
  -xr!PSI_S${SPRINT}_TP${TP}_G${GROUP}.ZIP > /dev/null

echo "Created $OUTPUT ($(du -h "$OUTPUT" | cut -f1))"
