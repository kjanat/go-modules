set -euo pipefail
cd candidate
tag=$(cat tag)
[[ $tag =~ ^wasm-tools/v[0-9]+\.[0-9]+\.[0-9]+-kjanat\.[0-9a-f]{12}$ ]]
sum=$(sha256sum wasm-tools.wasm.gz)
[[ $sum == "$(cat SHA256SUMS)" && ${sum:0:12} == "${tag##*.}" ]]
if gh release view "$tag" >/dev/null 2>&1; then
  mkdir -p published
  gh release download "$tag" --dir published -p wasm-tools.wasm.gz -p SHA256SUMS
  cmp published/wasm-tools.wasm.gz wasm-tools.wasm.gz
  cmp published/SHA256SUMS SHA256SUMS
else
  gh release create "$tag" --target "$(git rev-parse HEAD)" --latest=false \
    --title "wasm-tools ${tag#wasm-tools/v}" --notes "Tested WASI Preview 1 guest. SHA-256: ${sum%% *}" wasm-tools.wasm.gz SHA256SUMS
fi
