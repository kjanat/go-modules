"""Print the release tag for the locked guest and the tested archive."""

import hashlib
from pathlib import Path
import tomllib

root = Path("internal/wasmtools")
[tool] = [
    p
    for p in tomllib.loads((root / "Cargo.lock").read_text())["package"]
    if p["name"] == "wasm-tools"
]
digest = hashlib.sha256((root / "wasm-tools.wasm.gz").read_bytes()).hexdigest()
print(f"wasm-tools/v{tool['version']}-kjanat.{digest[:12]}")
