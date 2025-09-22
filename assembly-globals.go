package assembly

import (
	"debug/dwarf"
	"reflect"
	"unsafe"

	"github.com/go-delve/delve/pkg/proc"
)

// FindGlobal looks up a global variable by name.
// name specifies the name of the global variable to find.
// Returns the reflect.Value of the global variable, or an error if not found.
func (da *dwarfAssembly) FindGlobal(name string) (reflect.Value, error) {
	if nil == da.globals {
		da.loadGlobals()
	}

	if value, ok := da.globals[name]; ok {
		return value, nil
	}
	return reflect.Value{}, ErrNotFound
}

// ForeachGlobal iterates over all global variables, executing the callback function for each variable.
// fn is a callback function that receives the variable name and value.
// Returning false from the callback terminates iteration.
func (da *dwarfAssembly) ForeachGlobal(fn func(name string, value reflect.Value) bool) {
	if nil == da.globals {
		da.loadGlobals()
	}

	for name, value := range da.globals {
		if !fn(name, value) {
			break
		}
	}
}

func (da *dwarfAssembly) loadGlobals() {
	da.globals = make(map[string]reflect.Value)

	packageVars := reflect.ValueOf(da.binaryInfo).Elem().FieldByName("packageVars")
	if packageVars.IsValid() {
		for i, size := 0, packageVars.Len(); i < size; i++ {
			rv := packageVars.Index(i)
			rName := rv.FieldByName("name")
			rAddr := rv.FieldByName("addr")
			rOffset := rv.FieldByName("offset")
			rCU := rv.FieldByName("cu")
			if !rName.IsValid() || !rAddr.IsValid() || !rCU.IsValid() || !rOffset.IsValid() {
				continue
			}
			rImage := rCU.Elem().FieldByName("image")
			if !rImage.IsValid() {
				continue
			}
			rDwarf := rImage.Elem().FieldByName("dwarf")
			if !rDwarf.IsValid() {
				continue
			}
			image := (*proc.Image)(unsafe.Pointer(rImage.Pointer()))

			reader := image.DwarfReader()
			reader.Seek(dwarf.Offset(rOffset.Uint()))
			entry, err := reader.Next()
			if err != nil || entry == nil || entry.Tag != dwarf.TagVariable {
				continue
			}
			name, ok := entry.Val(dwarf.AttrName).(string)
			if !ok || rName.String() != name {
				continue
			}

			off, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
			if !ok {
				continue
			}

			dtyp, err := image.Type(off)
			if err != nil {
				continue
			}

			dname := godwarfTypeName(dtyp)
			if dname == "<unspecified>" || dname == "" {
				continue
			}

			rtyp, err := da.FindType(dname)
			if err != nil || rtyp == nil {
				continue
			}
			da.globals[name] = reflect.NewAt(rtyp, unsafe.Pointer(uintptr(rAddr.Uint()))).Elem()
		}
	}
}
