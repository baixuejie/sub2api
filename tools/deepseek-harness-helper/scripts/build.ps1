param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"
if ($Version -notmatch '^(dev|[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?)$') {
    throw "Invalid Helper version: $Version"
}
$root = Split-Path -Parent $PSScriptRoot
$out = Join-Path $root "dist"
$staging = Join-Path $out ".staging"
Remove-Item -Recurse -Force $out -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force $staging | Out-Null

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Ext = "" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Ext = "" }
)

Push-Location $root
try {
    go test ./...
    foreach ($target in $targets) {
        $env:CGO_ENABLED = "0"
        $env:GOOS = $target.GOOS
        $env:GOARCH = $target.GOARCH
        $name = "deepseek-harness-helper-$($target.GOOS)-$($target.GOARCH)$($target.Ext)"
        go build -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $out $name) ./cmd/deepseek-harness-helper
        if ($LASTEXITCODE -ne 0) { throw "go build failed for $($target.GOOS)/$($target.GOARCH)" }
    }

    foreach ($arch in @("amd64", "arm64")) {
        $linuxStage = Join-Path $staging "linux-$arch"
        New-Item -ItemType Directory -Force $linuxStage | Out-Null
        Copy-Item (Join-Path $out "deepseek-harness-helper-linux-$arch") (Join-Path $linuxStage "deepseek-harness-helper") -Force
        Copy-Item (Join-Path $root "README.md") (Join-Path $linuxStage "README.md") -Force
        $linuxArchive = Join-Path $out "deepseek-harness-helper-linux-$arch.tar.gz"
        tar -czf $linuxArchive -C $linuxStage .
        if ($LASTEXITCODE -ne 0) { throw "tar failed for linux/$arch" }

        $app = Join-Path (Join-Path $staging "darwin-$arch") "DeepSeek Harness Helper.app"
        $contents = Join-Path $app "Contents"
        $macos = Join-Path $contents "MacOS"
        New-Item -ItemType Directory -Force $macos | Out-Null
        Copy-Item (Join-Path $root "packaging/macos/Info.plist") (Join-Path $contents "Info.plist") -Force
        Copy-Item (Join-Path $out "deepseek-harness-helper-darwin-$arch") (Join-Path $macos "deepseek-harness-helper") -Force
        $darwinArchive = Join-Path $out "deepseek-harness-helper-darwin-$arch.tar.gz"
        tar -czf $darwinArchive -C (Split-Path -Parent $app) "DeepSeek Harness Helper.app"
        if ($LASTEXITCODE -ne 0) { throw "tar failed for darwin/$arch" }
    }

    foreach ($arch in @("amd64", "arm64")) {
        Remove-Item (Join-Path $out "deepseek-harness-helper-linux-$arch") -Force
        Remove-Item (Join-Path $out "deepseek-harness-helper-darwin-$arch") -Force
    }

    $assetNames = @(
        "deepseek-harness-helper-windows-amd64.exe",
        "deepseek-harness-helper-windows-arm64.exe",
        "deepseek-harness-helper-linux-amd64.tar.gz",
        "deepseek-harness-helper-linux-arm64.tar.gz",
        "deepseek-harness-helper-darwin-amd64.tar.gz",
        "deepseek-harness-helper-darwin-arm64.tar.gz"
    )
    $checksumLines = foreach ($asset in $assetNames) {
        $hash = (Get-FileHash -Algorithm SHA256 (Join-Path $out $asset)).Hash.ToLowerInvariant()
        "$hash  $asset"
    }
    Set-Content -Path (Join-Path $out "SHA256SUMS") -Value $checksumLines -Encoding ascii
} finally {
    Pop-Location
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force $staging -ErrorAction SilentlyContinue
}
