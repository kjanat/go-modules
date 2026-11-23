// Package the registry CLI as a WASI guest for go:embed.
use std::{
    env, fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

fn run(cmd: &mut Command) {
    assert!(
        cmd.status().expect("start build tool").success(),
        "{cmd:?} failed"
    );
}

fn main() {
    for input in [
        "Cargo.lock",
        "Cargo.toml",
        "rust-toolchain.toml",
        "build.rs",
        "patches",
    ] {
        println!("cargo:rerun-if-changed={input}");
    }
    let target = "wasm32-wasip1";
    let out = PathBuf::from(env::var_os("OUT_DIR").unwrap());
    // A separate target directory avoids locking the host Cargo invocation.
    let cargo = env::var_os("CARGO").unwrap();
    let metadata = Command::new(&cargo)
        .args([
            "metadata",
            "--locked",
            "--format-version=1",
            "--filter-platform",
            target,
        ])
        .output()
        .expect("read Cargo metadata");
    assert!(
        metadata.status.success(),
        "{}",
        String::from_utf8_lossy(&metadata.stderr)
    );
    let graph: serde_json::Value = serde_json::from_slice(&metadata.stdout).unwrap();
    let packages = graph["packages"].as_array().unwrap();
    let mut build = Command::new(cargo);
    build
        .args([
            "build",
            "--locked",
            "--release",
            "--target",
            target,
            "-p",
            "wasm-tools",
        ])
        .env("CARGO_ENCODED_RUSTFLAGS", "")
        .env("CARGO_TARGET_DIR", out.join("guest"));
    for pkg in packages.iter().filter(|p| p["name"] == "wit-parser") {
        let version = pkg["version"].as_str().unwrap();
        let patch = PathBuf::from(format!("patches/wit-parser-{version}.patch"));
        let src = Path::new(pkg["manifest_path"].as_str().unwrap())
            .parent()
            .unwrap();
        if !patch.exists() {
            continue;
        }
        let local = out.join("wit-parser");
        if local.exists() {
            fs::remove_dir_all(&local).unwrap();
        }
        run(Command::new("cp").arg("-R").arg(src).arg(&local));
        run(Command::new("chmod").args(["-R", "u+w"]).arg(&local));
        run(Command::new("patch")
            .args(["--fuzz=0", "-p1", "-i"])
            .arg(fs::canonicalize(patch).unwrap())
            .current_dir(&local));
        let paths = serde_json::json!([local]);
        build.args(["--config", &format!("paths={paths}")]);
        // trim-paths keeps a workspace-relative path, and this copy's path holds the build-script hash.
        let relative = local
            .strip_prefix(env::var("CARGO_MANIFEST_DIR").unwrap())
            .unwrap();
        build.env(
            "CARGO_ENCODED_RUSTFLAGS",
            format!(
                "--remap-path-prefix={}=wit-parser-{version}",
                relative.display()
            ),
        );
    }
    run(&mut build);
    let product = out
        .join("guest")
        .join(target)
        .join("release/wasm-tools.wasm");
    run(Command::new("gzip")
        .args(["-n", "-c"])
        .arg(product)
        .stdout(Stdio::from(
            fs::File::create(out.join("wasm-tools.wasm.gz")).unwrap(),
        )));
}
