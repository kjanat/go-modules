// Install the completed artifact even when Cargo reuses a cached build script.
fn main() -> std::io::Result<()> {
    let dst = std::path::Path::new(env!("CARGO_MANIFEST_DIR"));
    let tmp = dst.join("wasm-tools.wasm.gz.tmp");
    std::fs::copy(concat!(env!("OUT_DIR"), "/wasm-tools.wasm.gz"), &tmp)?;
    std::fs::rename(tmp, dst.join("wasm-tools.wasm.gz"))
}
