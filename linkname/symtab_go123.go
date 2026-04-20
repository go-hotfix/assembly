//go:build go1.23 && !go1.27

package linkname

// Offsets for runtime.moduledata fields on amd64/arm64 (Go 1.23).
// Verified against Go 1.23.0 source: src/runtime/symtab.go
//
// moduledata struct layout (64-bit):
//   offset 0:   pcHeader     *pcHeader
//   offset 8:   funcnametab  []byte
//   offset 32:  cutab        []uint32
//   offset 56:  filetab      []byte
//   offset 80:  pctab        []byte
//   offset 104: pclntable    []byte
//   offset 128: ftab         []functab      ← funcTabOffset
//   offset 152: findfunctab  uintptr
//   offset 160: minpc        uintptr
//   offset 168: maxpc        uintptr
//   offset 176: text         uintptr        ← textOffset
//   ...
//   offset 560: next         *moduledata    ← nextOffset

const (
	funcTabOffset = 128 // moduledata.ftab []functab
	textOffset    = 176 // moduledata.text uintptr
)
