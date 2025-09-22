# assembly
go runtime assembly library.

* Please keep the debugging symbols when compiling, and disable function inline `-gcflags=all=-l`

## API Overview
```go
// NewDwarfAssembly creates and initializes a new DwarfAssembly instance.
// It loads DWARF debug information from the current binary and prepares it
// for symbol lookup, type inspection, and function invocation.
func NewDwarfAssembly() (DwarfAssembly, error)

// DwarfAssembly provides an interface for analyzing binary programs using DWARF debug information.
// It enables access to global variables, type definitions, and function information within binary files,
// supporting operations such as dynamic image loading, symbol lookup, and function invocation.
type DwarfAssembly interface {
	// BinaryInfo returns the underlying binary information object containing loaded modules,
	// functions, types, and other debug information.
	BinaryInfo() *proc.BinaryInfo
	// LoadImage dynamically loads a shared library image into the process address space.
	// path specifies the file path of the image to load.
	// entryPoint specifies the entry point address of the image.
	// Returns an error if loading fails.
	LoadImage(path string, entryPoint uint64) error
	// Close releases all associated resources, including loaded images and binary information.
	Close() error

	// FindGlobal looks up a global variable by name.
	// name specifies the name of the global variable to find.
	// Returns the reflect.Value of the global variable, or an error if not found.
	FindGlobal(name string) (reflect.Value, error)
	// ForeachGlobal iterates over all global variables, executing the callback function for each variable.
	// fn is a callback function that receives the variable name and value.
	// Returning false from the callback terminates iteration.
	ForeachGlobal(fn func(name string, value reflect.Value) bool)

	// ForeachType iterates over all type definitions, executing the callback function for each type.
	// f is a callback function that receives the type name.
	// Returning false from the callback terminates iteration.
	// Returns an error if iteration fails.
	ForeachType(f func(name string) bool) error
	// FindType looks up a type definition by name.
	// name specifies the name of the type to find.
	// Returns the reflect.Type object, or an error if not found.
	FindType(name string) (reflect.Type, error)

	// FindFuncEntry looks up function entry information by name.
	// name specifies the name of the function to find.
	// Returns the function object containing entry address details, or an error if not found.
	FindFuncEntry(name string) (*proc.Function, error)
	// FindFuncPc looks up a function's entry address by name.
	// name specifies the name of the function to find.
	// Returns the program counter (PC) value of the function, or 0 with an error if not found.
	FindFuncPc(name string) (uint64, error)
	// FindFuncType looks up a function's type signature by name.
	// name specifies the name of the function to find.
	// variadic indicates whether to treat the function as a variadic function.
	// Returns the reflect.Type of the function, or an error if not found.
	FindFuncType(name string, variadic bool) (reflect.Type, error)
	// FindFunc looks up a function by name and creates a callable reflect.Value.
	// name specifies the name of the function to find.
	// variadic indicates whether to treat the function as a variadic function.
	// Returns a callable reflect.Value of the function, or an error if not found.
	FindFunc(name string, variadic bool) (reflect.Value, error)
	// ForeachFunc iterates over all functions, executing the callback function for each function.
	// f is a callback function that receives the function name and entry address.
	// Returning false from the callback terminates iteration.
	ForeachFunc(f func(name string, pc uint64) bool)
	// CallFunc invokes a function by name.
	// name specifies the name of the function to call.
	// variadic indicates whether to treat the function as a variadic function.
	// args specifies the list of function arguments.
	// Returns the function call results, or an error if invocation fails.
	CallFunc(name string, variadic bool, args []reflect.Value) ([]reflect.Value, error)

	// SearchPluginByName searches for a plugin by name.
	// name specifies the name of the plugin to find.
	// Returns the library file path and memory address where the plugin is located,
	// or an error if not found.
	SearchPluginByName(name string) (lib string, addr uint64, err error)
	// SearchPlugins searches for all available plugins.
	// Returns lists of library file paths and memory addresses for all plugins found,
	// or an error if the search fails.
	SearchPlugins() (libs []string, addrs []uint64, err error)
}
```

### Go Test
```
$ go test -gcflags "all=-N -l" -ldflags "-w=false -s=false" ./...
```

### License

The repository released under version 2.0 of the Apache License.