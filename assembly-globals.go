package assembly

import (
	"debug/dwarf"
	"iter"
	"reflect"

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

// Globals returns an iterator over all global variables.
// Each iteration yields the variable name and its reflect.Value.
func (da *dwarfAssembly) Globals() iter.Seq2[string, reflect.Value] {
	return func(yield func(string, reflect.Value) bool) {
		if da.globals == nil {
			da.loadGlobals()
		}
		for name, value := range da.globals {
			if !yield(name, value) {
				return
			}
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
			image := (*proc.Image)(pointerOf(rImage.Pointer()))

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
			addr := uintptr(rAddr.Uint())
			da.globals[name] = reflect.NewAt(rtyp, pointerOf(addr)).Elem()
		}
	}
}
