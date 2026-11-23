# Release

This document describes the steps to release a new version of module `github.com/kjanat/go-modules/cm`.

## 1. Update [CHANGELOG.md](./CHANGELOG.md)

1. Add the latest changes to [CHANGELOG.md](./CHANGELOG.md).
1. Rename the Unreleased section to the new version number and add today's date (YYYY-MM-DD).
1. Submit a pull request with these updates.

## 2. Tag the release

Once the pull request is merged, tag the merge commit with a `cm/`-prefixed version and push the tag:

```console
git tag --sign --annotate cm/v0.4.0 --message cm/v0.4.0
git push origin cm/v0.4.0
```
