#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUT="$ROOT/dist"
VERSION=${VERSION:-dev}
if ! printf '%s\n' "$VERSION" | grep -Eq '^(dev|[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?)$'; then
  echo "invalid Helper version: $VERSION" >&2
  exit 2
fi
STAGING="$OUT/.staging"
rm -rf "$OUT"
mkdir -p "$STAGING"
trap 'rm -rf "$STAGING"' EXIT
cd "$ROOT"

go test ./...

for target in windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
  GOOS=${target%/*}
  GOARCH=${target#*/}
  suffix=""
  [ "$GOOS" = windows ] && suffix=".exe"
  output="$OUT/deepseek-harness-helper-$GOOS-$GOARCH$suffix"
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" \
    -o "$output" ./cmd/deepseek-harness-helper
done

for arch in amd64 arm64; do
  linux_stage="$STAGING/linux-$arch"
  mkdir -p "$linux_stage"
  cp "$OUT/deepseek-harness-helper-linux-$arch" "$linux_stage/deepseek-harness-helper"
  chmod 755 "$linux_stage/deepseek-harness-helper"
  cp "$ROOT/README.md" "$linux_stage/README.md"
  tar -czf "$OUT/deepseek-harness-helper-linux-$arch.tar.gz" -C "$linux_stage" .

  app_parent="$STAGING/darwin-$arch"
  app="$app_parent/DeepSeek Harness Helper.app/Contents"
  mkdir -p "$app/MacOS"
  cp "$ROOT/packaging/macos/Info.plist" "$app/Info.plist"
  cp "$OUT/deepseek-harness-helper-darwin-$arch" "$app/MacOS/deepseek-harness-helper"
  chmod 755 "$app/MacOS/deepseek-harness-helper"
  tar -czf "$OUT/deepseek-harness-helper-darwin-$arch.tar.gz" -C "$app_parent" "DeepSeek Harness Helper.app"
done

rm -f \
  "$OUT/deepseek-harness-helper-linux-amd64" \
  "$OUT/deepseek-harness-helper-linux-arm64" \
  "$OUT/deepseek-harness-helper-darwin-amd64" \
  "$OUT/deepseek-harness-helper-darwin-arm64"

assets="
deepseek-harness-helper-windows-amd64.exe
deepseek-harness-helper-windows-arm64.exe
deepseek-harness-helper-linux-amd64.tar.gz
deepseek-harness-helper-linux-arm64.tar.gz
deepseek-harness-helper-darwin-amd64.tar.gz
deepseek-harness-helper-darwin-arm64.tar.gz
"
: > "$OUT/SHA256SUMS"
for asset in $assets; do
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$OUT/$asset" | sed "s#  $OUT/#  #" >> "$OUT/SHA256SUMS"
  else
    shasum -a 256 "$OUT/$asset" | sed "s#  $OUT/#  #" >> "$OUT/SHA256SUMS"
  fi
done
