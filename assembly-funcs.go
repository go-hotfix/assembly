package assembly

import (
	"debug/dwarf"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/go-delve/delve/pkg/proc"
)

func (da *dwarfAssembly) ForeachFunc(f func(name string, pc uint64) bool) {
	for _, function := range da.binaryInfo.Functions {
		if function.Entry != 0 {
			if !f(function.Name, function.Entry) {
				break
			}
		}
	}
}

func (da *dwarfAssembly) FindFuncEntry(name string) (*proc.Function, error) {
	f, err := da.findFunc(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (da *dwarfAssembly) FindFuncPc(name string) (uint64, error) {
	f, err := da.findFunc(name)
	if err != nil {
		return 0, err
	}
	return f.Entry, nil
}

func (da *dwarfAssembly) FindFuncType(name string, variadic bool) (reflect.Type, error) {
	f, err := da.findFunc(name)
	if err != nil {
		return nil, err
	}
	inTyps, outTyps, _, _, err := da.getFunctionArgTypes(f)
	if err != nil {
		return nil, err
	}

	ftyp := reflect.FuncOf(inTyps, outTyps, variadic)
	return ftyp, nil
}

func (da *dwarfAssembly) FindFunc(name string, variadic bool) (reflect.Value, error) {
	pc, err := da.FindFuncPc(name)
	if err != nil {
		return reflect.Value{}, err
	}
	ftyp, err := da.FindFuncType(name, variadic)
	if err != nil {
		return reflect.Value{}, err
	}

	newFunc := CreateFuncForCodePtr(ftyp, pc)
	return newFunc, nil
}

func (da *dwarfAssembly) CallFunc(name string, variadic bool, args []reflect.Value) ([]reflect.Value, error) {
	f, err := da.findFunc(name)
	if err != nil {
		return nil, err
	}

	inTyps, outTyps, inNames, _, err := da.getFunctionArgTypes(f)
	if err != nil {
		return nil, err
	}

	ftyp := reflect.FuncOf(inTyps, outTyps, variadic)
	newFunc := CreateFuncForCodePtr(ftyp, f.Entry)

	getInTyp := func(i int) (reflect.Type, string) {
		if len(inTyps) <= 0 {
			return nil, ""
		}
		if i < len(inTyps)-1 {
			return inTyps[i], inNames[i]
		}
		if variadic {
			return inTyps[len(inTyps)-1].Elem(), inNames[len(inNames)-1]
		}
		if i < len(inTyps) {
			return inTyps[i], inNames[i]
		}
		return nil, ""
	}

	for i, arg := range args {
		inTyp, inName := getInTyp(i)
		if inTyp == nil {
			return nil, fmt.Errorf("len mismatch %d", i)
		}

		if !arg.Type().AssignableTo(inTyp) {
			return nil, fmt.Errorf("type mismatch %d:%s", i, inName)
		}
	}

	out := newFunc.Call(args)
	return out, nil
}

func (da *dwarfAssembly) findFunc(name string) (*proc.Function, error) {
	if fns, _ := da.binaryInfo.FindFunction(name); nil != fns {
		return fns[len(fns)-1], nil
	}
	return nil, ErrNotFound
}

func (da *dwarfAssembly) getFunctionArgTypes(f *proc.Function) ([]reflect.Type, []reflect.Type, []string, []string, error) {
	rOffset := reflect.ValueOf(f).Elem().FieldByName("offset")
	rCU := reflect.ValueOf(f).Elem().FieldByName("cu")
	if !rOffset.IsValid() || !rCU.IsValid() {
		return nil, nil, nil, nil, ErrNotSupport
	}
	rImage := rCU.Elem().FieldByName("image")
	if !rImage.IsValid() {
		return nil, nil, nil, nil, ErrNotSupport
	}
	rDwarf := rImage.Elem().FieldByName("dwarf")
	if !rDwarf.IsValid() {
		return nil, nil, nil, nil, ErrNotSupport
	}
	image := (*proc.Image)(unsafe.Pointer(rImage.Pointer()))
	dwarfData := (*dwarf.Data)(unsafe.Pointer(rDwarf.Pointer()))

	reader := image.DwarfReader()
	reader.Seek(dwarf.Offset(rOffset.Uint()))
	entry, err := reader.Next()
	if err != nil || entry == nil || entry.Tag != dwarf.TagSubprogram {
		return nil, nil, nil, nil, fmt.Errorf("get function arg types not found %s", f.Name)
	}
	name, ok := entry.Val(dwarf.AttrName).(string)
	if !ok || f.Name != name {
		return nil, nil, nil, nil, fmt.Errorf("get function arg types name err %s:%s", f.Name, name)
	}

	var inTyps []reflect.Type
	var outTyps []reflect.Type
	var inNames []string
	var outNames []string

	for {
		child, err := reader.Next()
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("get function arg types reader err %s:%s", f.Name, err.Error())
		}
		if child == nil || child.Tag == 0 {
			break
		}
		if child.Tag != dwarf.TagFormalParameter {
			continue
		}

		dtyp, err := entryType(dwarfData, child)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("get function arg types type err %s:%w", f.Name, err)
		}
		dname := dwarfTypeName(dtyp)
		rtyp, err := da.FindType(dname)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("get function arg types type err %s(%s):%w", f.Name, dname, err)
		}

		isret, _ := child.Val(dwarf.AttrVarParam).(bool)
		if isret {
			outTyps = append(outTyps, rtyp)
			outNames = append(outNames, dname)
		} else {
			inTyps = append(inTyps, rtyp)
			inNames = append(inNames, dname)
		}
	}
	return inTyps, outTyps, inNames, outNames, nil
}
