# macOS app bundle

Build the Darwin binary and place it at:

```text
DeepSeek Harness Helper.app/Contents/MacOS/deepseek-harness-helper
```

Copy `Info.plist` to `DeepSeek Harness Helper.app/Contents/Info.plist`, sign the complete bundle with `codesign`, notarize it for distribution, then launch it once with `open`. LaunchServices reads `CFBundleURLTypes` and routes `sub2api-harness://` URLs to the executable as its single argument.

The repository does not contain a signing identity, entitlements, or notarization credentials. An unsigned local development bundle can be ad-hoc signed with `codesign --force --deep --sign -`, but that is not a distributable artifact.
