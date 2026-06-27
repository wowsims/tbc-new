package worker

import _ "embed"

// HighsWASM is the HiGHS WebAssembly module consumed by backend reforge optimizer runtime.
//
//go:embed highs.wasm
var HighsWASM []byte
