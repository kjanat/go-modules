set -euo pipefail
cd candidate
tag=$(cat tag)
[[ $tag =~ ^wasm-tools/v[0-9]+\.[0-9]+\.[0-9]+-kjanat\.[0-9a-f]{12}$ ]]
version=${tag#wasm-tools/v}
version=${version%-kjanat.*}
title="wasm-tools $version"
sum=$(sha256sum wasm-tools.wasm.gz)
[[ $sum == "$(cat SHA256SUMS)" && ${sum:0:12} == "${tag##*.}" ]]
python3 - "$version" "$tag" > release-notes.md <<'PY'
import os
import sys
import tomllib
from pathlib import Path

version, tag = sys.argv[1:]
print(f"[Upstream release notes for wasm-tools {version}]"
      f"(https://github.com/bytecodealliance/wasm-tools/releases/tag/v{version})")
root = Path("../internal/wasmtools")
packages = tomllib.loads((root / "Cargo.lock").read_text())["package"]
for pkg in packages:
    if pkg["name"] == "wit-parser":
        patch = Path(f"internal/wasmtools/patches/wit-parser-{pkg['version']}.patch")
        if (Path("..") / patch).is_file():
            print(f"\nLocal patch: [{patch.name}]"
                  f"(https://github.com/{os.environ['GITHUB_REPOSITORY']}/blob/{tag}/{patch})")
PY
if gh release view "$tag" >/dev/null 2>&1; then
  mkdir -p published
  gh release download "$tag" --dir published -p wasm-tools.wasm.gz -p SHA256SUMS
  cmp published/wasm-tools.wasm.gz wasm-tools.wasm.gz
  cmp published/SHA256SUMS SHA256SUMS
  gh release edit "$tag" --title "$title" --notes-file release-notes.md
else
  gh release create "$tag" --target "$(git rev-parse HEAD)" --latest=false \
    --title "$title" --notes-file release-notes.md wasm-tools.wasm.gz SHA256SUMS
fi
