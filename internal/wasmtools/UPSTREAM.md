# Embedded wasm-tools

Go builds embed the checked-in WASI Preview 1 guest. Installing the Go module
does not download wasm-tools, compile Rust, or access GitHub Releases.

The Monday Renovate workflow (also manually dispatchable) updates the exact
wasm-tools dependency in Cargo.toml and Cargo.lock. The temporary Cargo patch
for 1.263.0 supplies upstream's post-release parser-limit fixes. A newer exact
version can resolve from crates.io without that patch. Set the repository or
organization Actions secret `RENOVATE_TOKEN` to a bot token with contents and
pull-request write access; workflow write access is needed for action updates.

On a candidate PR, CI runs `make wasm-tools` (Cargo --locked, with the Rust
toolchain and WASI target from rust-toolchain.toml), then runs the Go tests
against the rebuilt guest. Keep the PR up to date and require the candidate
job before merging: promotion checks for the successful artifact for the
exact merged tree. Artifacts expire after 30 days; rerun before merging if
necessary. The candidate has no release permissions.

After merge, CI publishes those bytes and their SHA-256 to a GitHub Release
tagged `wasm-tools/v<upstream>-kjanat.<digest-prefix>`. This is separate from
Go module tags and never changes the repository's Latest release. CI opens
a second PR that embeds the published asset and records its tag. That PR
checks the binary against the release and runs the normal integration suite.
