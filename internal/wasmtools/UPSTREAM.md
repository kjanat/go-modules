# Embedded wasm-tools

Go builds embed the checked-in WASI Preview 1 guest. Installing the Go module
does not download wasm-tools, compile Rust, or access GitHub Releases.

The installed Renovate GitHub App uses the Monday update window in
renovate.json (Europe/Amsterdam) to propose Cargo.toml and Cargo.lock updates.
Cargo.lock resolves wasm-tools and its dependencies from crates.io. The host
packager's build.rs compiles the locked CLI for wasm32-wasip1. For wit-parser
0.263.0 only, it applies the [upstream parser-limit fix](https://github.com/bytecodealliance/wasm-tools/pull/2813)
to a private copy of the published source; the Cargo registry is never edited.
Beepboop publishes the release
using a repository-scoped installation token obtained through GitHub OIDC.

On a candidate PR, CI runs `make wasm-tools` (Cargo --locked, with the Rust
toolchain and WASI target from rust-toolchain.toml), then runs the Go tests
against the rebuilt guest. Keep the PR up to date and require the candidate
job before merging: promotion checks for the successful artifact for the
exact merged tree. Artifacts expire after 30 days; rerun before merging if
necessary. The candidate has no release permissions.

After merge, CI publishes those bytes and their SHA-256 to a GitHub Release
tagged `wasm-tools/v<upstream>-kjanat.<digest-prefix>`. This is separate from
Go module tags and never changes the repository's Latest release. Publishing does not update the guest embedded by the Go module.
