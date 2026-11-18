set -euo pipefail
: "${GH_TOKEN:?Set RENOVATE_TOKEN to a bot token with contents and pull request write access}"
tag=$(cat candidate/tag)
branch="update/${tag}"
gh auth setup-git
git fetch origin master
git switch -C "$branch" origin/master
mkdir -p candidate/published
gh release download "$tag" -p wasm-tools.wasm.gz -p SHA256SUMS --dir candidate/published --clobber
(cd candidate/published && sha256sum --check SHA256SUMS)
cmp candidate/wasm-tools.wasm.gz candidate/published/wasm-tools.wasm.gz
cp candidate/published/wasm-tools.wasm.gz internal/wasmtools/
echo "$tag" > internal/wasmtools/embedded-release
git add internal/wasmtools/{embedded-release,wasm-tools.wasm.gz}
git diff --cached --quiet && exit 0
git config user.name 'github-actions[bot]'
git config user.email '41898282+github-actions[bot]@users.noreply.github.com'
if ! git ls-remote --exit-code origin "refs/heads/$branch" >/dev/null; then
  git commit -m "Update embedded guest to $tag"
  git push origin "HEAD:refs/heads/$branch"
fi
if [[ -z $(gh pr list --state all --head "$branch" --json number --jq '.[0].number // empty') ]]; then
  gh pr create --base master --head "$branch" --title "Update embedded guest to $tag" \
    --body "Embed the tested wasm-tools release: https://github.com/$GITHUB_REPOSITORY/releases/tag/$tag"
fi
