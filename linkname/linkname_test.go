//go:build go1.23

package linkname

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestFuncPCForName_RuntimeFunctions(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		shouldFind bool
	}{
		{
			name:       "runtime.stopTheWorld",
			funcName:   "runtime.stopTheWorld",
			shouldFind: true,
		},
		{
			name:       "runtime.startTheWorld",
			funcName:   "runtime.startTheWorld",
			shouldFind: true,
		},
		{
			name:       "runtime.gcBgMarkWorker",
			funcName:   "runtime.gcBgMarkWorker",
			shouldFind: true,
		},
		{
			name:       "runtime.mallocgc",
			funcName:   "runtime.mallocgc",
			shouldFind: true,
		},
		{
			name:       "nonexistent function",
			funcName:   "nonexistent.function.that.does.not.exist",
			shouldFind: false,
		},
		{
			name:       "random fake package",
			funcName:   "xyz.abc123.FakeFunc",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := FuncPCForName(tt.funcName)
			if tt.shouldFind {
				if pc == 0 {
					t.Errorf("FuncPCForName(%q) returned 0, expected non-zero", tt.funcName)
				} else {
					// Verify the PC maps back to the correct function
					fn := runtime.FuncForPC(pc)
					if fn == nil {
						t.Errorf("runtime.FuncForPC(%#x) returned nil for %q", pc, tt.funcName)
					} else if fn.Name() != tt.funcName {
						t.Errorf("FuncPCForName(%q) returned PC for %q", tt.funcName, fn.Name())
					}
				}
			} else {
				if pc != 0 {
					t.Errorf("FuncPCForName(%q) returned %#x, expected 0", tt.funcName, pc)
				}
			}
		})
	}
}

func TestFuncPCForName_MapInitialized(t *testing.T) {
	if nameMap == nil {
		t.Fatal("nameMap is nil, init() did not run properly")
	}
	if len(nameMap) == 0 {
		t.Fatal("nameMap is empty, no functions discovered")
	}
	t.Logf("Discovered %d functions in moduledata", len(nameMap))
}

func TestFuncPCForName_Findfunc(t *testing.T) {
	// Test that the single go:linkname to runtime.findfunc works
	f := getModuleData
	entry := **(**uintptr)(unsafe.Pointer(&f))
	_, md := findfunc(entry)
	if md == nil {
		t.Fatal("findfunc returned nil moduledata")
	}
}
