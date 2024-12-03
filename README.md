# assembly
go runtime assembly library.

* Please keep the debugging symbols when compiling, and disable function inline `-gcflags=all=-l`

## API Overview
```
func NewDwarfAssembly() (DwarfAssembly, error)

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
```

### Go Test
```
$ go test -c -gcflags="all=-l -N" ./...
$ ./assembly.test
```

### License

The repository released under version 2.0 of the Apache License.