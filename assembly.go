package assembly

import (
	"errors"
	"iter"
	"os"
	"reflect"
	"runtime"

	"github.com/go-delve/delve/pkg/proc"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrNotSupport = errors.New("not support")
)

// DwarfAssembly provides an interface for analyzing binary programs using DWARF debug information.
type DwarfAssembly interface {
	// LoadImage dynamically loads a shared library image into the process address space.
	LoadImage(path string, entryPoint uint64) error
	// Close releases all associated resources.
	Close() error

	// FindGlobal looks up a global variable by name.
	FindGlobal(name string) (reflect.Value, error)
	// Globals returns an iterator over all global variables.
	Globals() iter.Seq2[string, reflect.Value]

	// FindType looks up a type definition by name.
	FindType(name string) (reflect.Type, error)
	// Types returns an iterator over all type definition names.
	// Panics if the underlying type information cannot be loaded.
	Types() iter.Seq[string]

	// FindFunc looks up a function by name.
	// Returns the function object containing entry address details, or an error if not found.
	FindFunc(name string) (*proc.Function, error)
	// FindFuncType looks up a function's type signature by name.
	// variadic indicates whether to treat the function as a variadic function.
	FindFuncType(name string, variadic bool) (reflect.Type, error)
	// FindFuncValue looks up a function by name and creates a callable reflect.Value.
	// variadic indicates whether to treat the function as a variadic function.
	FindFuncValue(name string, variadic bool) (reflect.Value, error)
	// Funcs returns an iterator over all function names.
	Funcs() iter.Seq[string]
	// CallFunc invokes a function by name.
	// variadic indicates whether to treat the function as a variadic function.
	// args specifies the list of function arguments.
	CallFunc(name string, variadic bool, args []reflect.Value) ([]reflect.Value, error)

	// FindPlugin searches for a plugin by name.
	FindPlugin(name string) (lib string, addr uint64, err error)
	// Plugins returns an iterator over all plugin library names.
	Plugins() iter.Seq[string]
}

type dwarfAssembly struct {
	binaryInfo *proc.BinaryInfo
	modules    []ModuleData
	globals    map[string]reflect.Value
	imageTypes map[*proc.Image]map[string]uint64
}

// NewDwarfAssembly creates and initializes a new DwarfAssembly instance.
func NewDwarfAssembly() (DwarfAssembly, error) {
	path, err := os.Executable()
	if nil != err {
		return nil, err
	}

	assembly := &dwarfAssembly{binaryInfo: proc.NewBinaryInfo(runtime.GOOS, runtime.GOARCH)}

	var entryPoint uintptr
	if entryPoint, err = getEntrypoint(path); nil != err {
		return nil, err
	}

	if err = assembly.LoadImage(path, uint64(entryPoint)); nil != err {
		return nil, err
	}

	runtime.SetFinalizer(assembly, (*dwarfAssembly).Close)
	return assembly, nil
}

func (da *dwarfAssembly) LoadImage(path string, entryPoint uint64) (err error) {

	if 0 == len(da.binaryInfo.Images) {
		if err = da.binaryInfo.LoadBinaryInfo(path, entryPoint, nil); nil != err {
			return
		}
	} else {
		if err = da.binaryInfo.AddImage(path, entryPoint); nil != err {
			return
		}
	}

	return da.refreshModules()
}

func (da *dwarfAssembly) refreshModules() error {
	modules, err := loadModuleData(da.binaryInfo, new(localMemory))
	if nil != err {
		return err
	}
	da.modules = modules
	da.globals = nil
	return nil
}

func (da *dwarfAssembly) Close() error {
	da.modules = nil
	da.globals = nil
	da.imageTypes = nil
	runtime.SetFinalizer(da, nil)
	return da.binaryInfo.Close()
}
