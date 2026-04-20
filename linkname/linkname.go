//go:build go1.23

package linkname

import (
	"reflect"
	"runtime"
	"unsafe"
)

//go:linkname findfunc runtime.findfunc
func findfunc(_ uintptr) (unsafe.Pointer, unsafe.Pointer)

var nameMap map[string]uintptr

func init() {
	nameMap = make(map[string]uintptr)
	md := getModuleData()
	if md == nil {
		return
	}

	buildFuncMap(md)

	// Note: moduledata.next traversal is not implemented because the `next` field
	// offset varies across Go versions and all statically-compiled functions
	// (runtime, delve, gohook) are in the main moduledata.
}

func buildFuncMap(md unsafe.Pointer) {
	textStart := *(*uintptr)(unsafe.Pointer(uintptr(md) + uintptr(textOffset)))

	// Read ftab slice header from moduledata
	funcTabPtr := *(**functab)(unsafe.Pointer(uintptr(md) + uintptr(funcTabOffset)))
	funcTabLen := *(*int)(unsafe.Pointer(uintptr(md) + uintptr(funcTabOffset) + unsafe.Sizeof(uintptr(0))))

	// Build a Go slice from raw data pointer and length
	header := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(funcTabPtr)),
		Len:  funcTabLen,
		Cap:  funcTabLen,
	}
	funcTabs := *(*[]functab)(unsafe.Pointer(&header))

	for i := range funcTabs {
		pc := textStart + uintptr(funcTabs[i].entryoff)
		fun := runtime.FuncForPC(pc)
		if fun != nil {
			nameMap[fun.Name()] = pc
		}
	}
}

type functab struct {
	entryoff uint32
	funcoff  uint32
}

func getModuleData() unsafe.Pointer {
	// Get the PC of this function itself to pass to findfunc
	f := getModuleData
	// Function value in Go is a pointer to a closure struct {code_ptr, ...}
	// Dereference twice: &f → function value pointer → code pointer
	entry := **(**uintptr)(unsafe.Pointer(&f))
	_, md := findfunc(entry)
	return md
}

// FuncPCForName returns the entry PC (program counter) for the function with the given full name.
// Returns 0 if the function is not found.
func FuncPCForName(name string) uintptr {
	return nameMap[name]
}
