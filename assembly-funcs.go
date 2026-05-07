package assembly

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/go-delve/delve/pkg/proc"
)

// Funcs returns an iterator over all function names.
func (da *dwarfAssembly) Funcs() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, function := range da.binaryInfo.Functions {
			if function.Entry != 0 {
				if !yield(function.Name) {
					return
				}
			}
		}
	}
}

// FindFunc looks up a function by name.
// Returns the function object containing entry address details, or an error if not found.
func (da *dwarfAssembly) FindFunc(name string) (*proc.Function, error) {
	f, err := da.findFunc(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (da *dwarfAssembly) findFuncPc(name string) (uint64, error) {
	f, err := da.findFunc(name)
	if err != nil {
		return 0, err
	}
	return f.Entry, nil
}

// FindFuncType looks up a function's type signature by name.
// variadic indicates whether to treat the function as a variadic function.
// Returns the reflect.Type of the function, or an error if not found.
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

// FindFuncValue looks up a function by name and creates a callable reflect.Value.
// variadic indicates whether to treat the function as a variadic function.
// Returns a callable reflect.Value of the function, or an error if not found.
func (da *dwarfAssembly) FindFuncValue(name string, variadic bool) (reflect.Value, error) {
	pc, err := da.findFuncPc(name)
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

// CallFunc invokes a function by name.
// variadic indicates whether to treat the function as a variadic function.
// args specifies the list of function arguments.
// Returns the function call results, or an error if invocation fails.
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
			return nil, fmt.Errorf("type mismatch arg: %d:%s, except: %s, got: %s", i, inName, inTyp.String(), arg.Type().String())
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

	_, args, err := funcCallArgs(f, da.binaryInfo, true)
	if nil != err {
		return nil, nil, nil, nil, fmt.Errorf("resolve function args failed: %s:%w", f.Name, err)
	}

	var inTyps []reflect.Type
	var outTyps []reflect.Type
	var inNames []string
	var outNames []string

	for idx, arg := range args {
		argType := resolveTypedef(arg.typ)
		rtyp, err := da.FindType(godwarfTypeName(argType))
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("resolve function arg failed: %s arg: %d: (%s %s): %w", f.Name, idx, arg.name, argType.String(), err)
		}

		if arg.isret {
			outTyps = append(outTyps, rtyp)
			outNames = append(outNames, arg.name)
		} else {
			inTyps = append(inTyps, rtyp)
			inNames = append(inNames, arg.name)
		}
	}

	return inTyps, outTyps, inNames, outNames, nil
}
