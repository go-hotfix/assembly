# Assembly

[![Go Reference][1]][2] [![license-Apache 2][3]][4] [![Go Report Card][5]][6] [![Test][7]][8]

Provides DWARF-based binary analysis for Go programs, including symbol lookup, type inspection, function invocation, and plugin discovery. Supports Linux, macOS, and Windows platforms.

[1]: https://pkg.go.dev/badge/github.com/go-hotfix/assembly.svg
[2]: https://pkg.go.dev/github.com/go-hotfix/assembly
[3]: https://img.shields.io/badge/license-Apache%202-blue.svg
[4]: https://opensource.org/licenses/Apache-2.0
[5]: https://goreportcard.com/badge/github.com/go-hotfix/assembly
[6]: https://goreportcard.com/report/github.com/go-hotfix/assembly
[7]: https://img.shields.io/github/actions/workflow/status/go-hotfix/assembly/test.yml?branch=main&label=test
[8]: https://github.com/go-hotfix/assembly/actions?query=workflow%3ATest

## API Overview
```go
// NewDwarfAssembly creates and initializes a new DwarfAssembly instance.
// It loads DWARF debug information from the current binary and prepares it
// for symbol lookup, type inspection, and function invocation.
func NewDwarfAssembly() (DwarfAssembly, error)

// DwarfAssembly provides an interface for analyzing binary programs using DWARF debug information.
type DwarfAssembly interface {
	// LoadImage dynamically loads a shared library image into the process address space.
	LoadImage(path string, entryPoint uint64) error
	// Close releases all associated resources.
	Close() error

	// Global variables
	FindGlobal(name string) (reflect.Value, error)
	Globals() iter.Seq2[string, reflect.Value]

	// Type definitions
	FindType(name string) (reflect.Type, error)
	Types() iter.Seq[string]

	// Functions
	FindFunc(name string) (*proc.Function, error)
	FindFuncType(name string, variadic bool) (reflect.Type, error)
	FindFuncValue(name string, variadic bool) (reflect.Value, error)
	Funcs() iter.Seq[string]
	CallFunc(name string, variadic bool, args []reflect.Value) ([]reflect.Value, error)

	// Plugins
	FindPlugin(name string) (lib string, addr uint64, err error)
	Plugins() iter.Seq[string]
}
```

### Go Test
```
$ go test -gcflags "all=-N -l" -ldflags "-w=false -s=false" ./...
```

### License

The repository released under version 2.0 of the Apache License.
