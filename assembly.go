package assembly

import (
	"errors"
	"os"
	"reflect"
	"runtime"

	"github.com/go-delve/delve/pkg/proc"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrNotSupport       = errors.New("not support")
	ErrTooManyLibraries = errors.New("number of loaded libraries exceeds maximum")
)

type DwarfAssembly interface {
	BinaryInfo() *proc.BinaryInfo
	LoadImage(path string, entryPoint uint64) error
	Close() error

	FindGlobal(name string) (reflect.Value, error)
	ForeachGlobal(fn func(name string, value reflect.Value) bool)

	ForeachType(f func(name string) bool) error
	FindType(name string) (reflect.Type, error)

	FindFuncEntry(name string) (*proc.Function, error)
	FindFuncPc(name string) (uint64, error)
	FindFuncType(name string, variadic bool) (reflect.Type, error)
	FindFunc(name string, variadic bool) (reflect.Value, error)
	ForeachFunc(f func(name string, pc uint64) bool)
	CallFunc(name string, variadic bool, args []reflect.Value) ([]reflect.Value, error)

	SearchPluginByName(name string) (lib string, addr uint64, err error)
	SearchPlugins() (libs []string, addrs []uint64, err error)
}

type dwarfAssembly struct {
	binaryInfo *proc.BinaryInfo
	modules    []ModuleData
	globals    map[string]reflect.Value
	imageTypes map[*proc.Image]map[string]uint64
}

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

func (da *dwarfAssembly) BinaryInfo() *proc.BinaryInfo {
	return da.binaryInfo
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
