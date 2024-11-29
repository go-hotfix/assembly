package assembly

import (
	"debug/dwarf"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/go-delve/delve/pkg/dwarf/godwarf"
	"github.com/go-delve/delve/pkg/proc"
)

// ModuleData counterpart to proc.moduleData
type ModuleData struct {
	text, etext   uint64
	types, etypes uint64
	typemapVar    *proc.Variable
}

//go:linkname findType github.com/go-delve/delve/pkg/proc.(*BinaryInfo).findType
func findType(bi *proc.BinaryInfo, name string) (godwarf.Type, error)

//go:linkname loadModuleData github.com/go-delve/delve/pkg/proc.LoadModuleData
func loadModuleData(bi *proc.BinaryInfo, mem proc.MemoryReadWriter) ([]ModuleData, error)

//go:linkname imageToModuleData github.com/go-delve/delve/pkg/proc.(*BinaryInfo).imageToModuleData
func imageToModuleData(bi *proc.BinaryInfo, image *proc.Image, mds []ModuleData) *ModuleData

//go:linkname dwarfToRuntimeType github.com/go-delve/delve/pkg/proc.dwarfToRuntimeType
func dwarfToRuntimeType(bi *proc.BinaryInfo, mem proc.MemoryReadWriter, typ godwarf.Type) (typeAddr uint64, typeKind uint64, found bool, err error)

type localMemory int

func (mem *localMemory) ReadMemory(data []byte, addr uint64) (int, error) {
	buf := entryAddress(uintptr(addr), len(data))
	return copy(data, buf), nil
}

func (mem *localMemory) WriteMemory(addr uint64, data []byte) (int, error) {
	return 0, ErrNotSupport
}

func dwarfTypeName(dtyp dwarf.Type) string {
	switch dtyp := dtyp.(type) {
	case *dwarf.StructType:
		return dtyp.StructName
	case *dwarf.PtrType:
		return "*" + dwarfTypeName(dtyp.Type)
	case *dwarf.EnumType:
		return dtyp.EnumName
	default:
		name := dtyp.Common().Name
		if name != "" {
			return name
		}
		return dtyp.String()
	}
}

func entryType(data *dwarf.Data, entry *dwarf.Entry) (dwarf.Type, error) {
	off, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
	if !ok {
		return nil, fmt.Errorf("unable to find type offset for entry")
	}
	return data.Type(off)
}

func entryAddress(p uintptr, l int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: p, Len: l, Cap: l}))
}

type Func struct {
	codePtr uintptr
}

// CreateFuncForCodePtr https://github.com/alangpierce/go-forceexport/blob/8f1d6941cd755b975763ddb1f836561edddac2b8/forceexport.go#L31-L51
func CreateFuncForCodePtr(ftyp reflect.Type, codePtr uint64) reflect.Value {
	// Use reflect.MakeFunc to create a well-formed function value that's
	// guaranteed to be of the right type and guaranteed to be on the heap
	// (so that we can modify it). We give a nil delegate function because
	// it will never actually be called.
	newFuncVal := reflect.MakeFunc(ftyp, nil)
	// Use reflection on the reflect.Value (yep!) to grab the underling
	// function value pointer. Trying to call newFuncVal.Pointer() wouldn't
	// work because it gives the code pointer rather than the function value
	// pointer. The function value is a struct that starts with its code
	// pointer, so we can swap out the code pointer with our desired value.
	funcValuePtr := reflect.ValueOf(newFuncVal).FieldByName("ptr").Pointer()
	funcPtr := (*Func)(unsafe.Pointer(funcValuePtr))
	funcPtr.codePtr = uintptr(codePtr)
	return newFuncVal
}
