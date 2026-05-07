package assembly

import (
	"os"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/go-delve/delve/pkg/proc"
)

const (
	testVoidFQN        = "github.com/go-hotfix/assembly.testVoid"
	testAliasArgFQN    = "github.com/go-hotfix/assembly.testAliasArg"
	testGlobalPtrFQN   = "github.com/go-hotfix/assembly.testGlobalPtr"
	testGlobalAliasFQN = "github.com/go-hotfix/assembly.testGlobalAlias"
	typeTestAlias      = "github.com/go-hotfix/assembly.testAlias"
)

// ============================================================================
// LoadImage — cover AddImage branch
// ============================================================================

func TestLoadImage_AddImage(t *testing.T) {
	asm, err := NewDwarfAssembly()
	if err != nil {
		t.Fatalf("NewDwarfAssembly: %v", err)
	}
	defer asm.Close()

	path, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable: %v", err)
	}

	// Second call to LoadImage goes through the AddImage branch
	// because binaryInfo.Images already has entries from NewDwarfAssembly.
	err = asm.LoadImage(path, 0)
	if err != nil {
		t.Fatalf("LoadImage (AddImage) error: %v", err)
	}
}

// ============================================================================
// LoadImage — cover LoadBinaryInfo error path (empty Images + bad path)
// ============================================================================

func TestLoadImage_LoadBinaryInfoError(t *testing.T) {
	da := &dwarfAssembly{binaryInfo: proc.NewBinaryInfo(runtime.GOOS, runtime.GOARCH)}
	err := da.LoadImage("/nonexistent/path/to/binary", 0)
	if err == nil {
		t.Fatal("LoadImage: expected error for nonexistent path with empty Images")
	}
}

// ============================================================================
// CallFunc — cover variadic arg expansion (variadic=true, multiple args)
// ============================================================================

func TestCallFunc_VariadicArgExpansion(t *testing.T) {
	asm := newAssembly(t)

	results, err := asm.CallFunc("github.com/go-hotfix/assembly.testMax", true, []reflect.Value{
		reflect.ValueOf(10),
		reflect.ValueOf(3),
		reflect.ValueOf(7),
		reflect.ValueOf(15),
	})
	if err != nil {
		t.Fatalf("CallFunc variadic: %v", err)
	}
	if results[0].Int() != 15 {
		t.Fatalf("CallFunc variadic: got %d, want 15", results[0].Int())
	}
}

// ============================================================================
// CallFunc — cover non-variadic slice arg path
// ============================================================================

func TestCallFunc_VariadicSingleSliceArg(t *testing.T) {
	asm := newAssembly(t)

	results, err := asm.CallFunc("github.com/go-hotfix/assembly.testMax", false, []reflect.Value{
		reflect.ValueOf(10),
		reflect.ValueOf([]int{3, 7}),
	})
	if err != nil {
		t.Fatalf("CallFunc non-variadic: %v", err)
	}
	if results[0].Int() != 10 {
		t.Fatalf("CallFunc non-variadic: got %d, want 10", results[0].Int())
	}
}

// ============================================================================
// CallFunc — cover len(inTyps)<=0 via no-arg function with args
// ============================================================================

func TestCallFunc_NoArgWithArgs(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.CallFunc(testVoidFQN, false, []reflect.Value{
		reflect.ValueOf(1),
	})
	if err == nil {
		t.Fatal("CallFunc: expected error for passing args to no-arg function")
	}
}

// ============================================================================
// CallFunc — cover too many args
// ============================================================================

func TestCallFunc_TooManyArgs(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.CallFunc(testAddFQN, false, []reflect.Value{
		reflect.ValueOf(1),
		reflect.ValueOf(2),
		reflect.ValueOf(3),
	})
	if err == nil {
		t.Fatal("CallFunc: expected error for too many arguments")
	}
}

// ============================================================================
// CallFunc — cover generic variadic with expanded args
// ============================================================================

func TestCallFunc_GenericVariadic(t *testing.T) {
	asm := newAssembly(t)

	fqn := "github.com/go-hotfix/assembly.genericMin[int]"
	results, err := asm.CallFunc(fqn, true, []reflect.Value{
		reflect.ValueOf(10),
		reflect.ValueOf(3),
		reflect.ValueOf(7),
	})
	if err != nil {
		t.Fatalf("CallFunc generic variadic: %v", err)
	}
	if results[0].Int() != 3 {
		t.Fatalf("CallFunc generic variadic: got %d, want 3", results[0].Int())
	}
}

// ============================================================================
// CallFunc — cover method call (foo.bar)
// ============================================================================

func TestCallFunc_Method(t *testing.T) {
	asm := newAssembly(t)

	results, err := asm.CallFunc(testBarFQN, false, []reflect.Value{
		reflect.ValueOf(fooInstance),
	})
	if err != nil {
		t.Fatalf("CallFunc method: %v", err)
	}
	if results[0].Int() != 999 {
		t.Fatalf("CallFunc method: got %d, want 999", results[0].Int())
	}
}

// ============================================================================
// CallFunc — cover typedef arg (resolveTypedef TypedefType branch)
// ============================================================================

func TestCallFunc_AliasArg(t *testing.T) {
	asm := newAssembly(t)

	results, err := asm.CallFunc(testAliasArgFQN, false, []reflect.Value{
		reflect.ValueOf(testAlias(10)),
	})
	if err != nil {
		t.Fatalf("CallFunc alias arg: %v", err)
	}
	if results[0].Int() != 11 {
		t.Fatalf("CallFunc alias arg: got %d, want 11", results[0].Int())
	}
}

// ============================================================================
// FindFuncValue — cover variadic path
// ============================================================================

func TestFindFuncValue_Variadic(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue("github.com/go-hotfix/assembly.testMax", true)
	if err != nil {
		t.Fatalf("FindFuncValue variadic: %v", err)
	}

	results := fn.Call([]reflect.Value{
		reflect.ValueOf(5),
		reflect.ValueOf(2),
		reflect.ValueOf(8),
		reflect.ValueOf(1),
	})
	if results[0].Int() != 8 {
		t.Fatalf("FindFuncValue variadic call: got %d, want 8", results[0].Int())
	}
}

// ============================================================================
// FindFuncValue — cover method
// ============================================================================

func TestFindFuncValue_Method(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testBarFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue method: %v", err)
	}

	results := fn.Call([]reflect.Value{reflect.ValueOf(fooInstance)})
	if results[0].Int() != 999 {
		t.Fatalf("FindFuncValue method call: got %d, want 999", results[0].Int())
	}
}

// ============================================================================
// FindPlugin — cover found path
// ============================================================================

func TestFindPlugin_Found(t *testing.T) {
	asm := newAssembly(t)

	searchName := "libc"
	lib, addr, err := asm.FindPlugin(searchName)
	if err != nil {
		t.Skipf("FindPlugin(%s): %v (may not exist on this platform)", searchName, err)
	}
	if lib == "" {
		t.Fatal("FindPlugin: empty lib name")
	}
	if addr == 0 {
		t.Fatal("FindPlugin: addr is 0")
	}
}

// ============================================================================
// FindFunc — verify Name field
// ============================================================================

func TestFindFunc_EntryRange(t *testing.T) {
	asm := newAssembly(t)

	entry, err := asm.FindFunc(testAddFQN)
	if err != nil {
		t.Fatalf("FindFunc(%s) error: %v", testAddFQN, err)
	}
	if entry.Name != testAddFQN {
		t.Fatalf("FindFunc Name: got %q, want %q", entry.Name, testAddFQN)
	}
}

// ============================================================================
// FindFuncType — error path
// ============================================================================

func TestFindFuncType_ErrorPath(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.FindFuncType("nonexistent.Func", false)
	if err == nil {
		t.Fatal("FindFuncType: expected error for non-existent function")
	}
}

// ============================================================================
// LoadImage — error on bad path
// ============================================================================

func TestLoadImage_BadPath(t *testing.T) {
	asm := newAssembly(t)

	err := asm.LoadImage("/nonexistent/path/to/binary", 0)
	if err == nil {
		t.Fatal("LoadImage: expected error for bad path")
	}
}

// ============================================================================
// Globals — cover stop-early
// ============================================================================

func TestGlobals_StopEarly(t *testing.T) {
	asm := newAssembly(t)

	count := 0
	for range asm.Globals() {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("Globals: break should stop after 1, got %d", count)
	}
}

// ============================================================================
// FindGlobal — int, string, pointer, alias types
// ============================================================================

func TestFindGlobal_Int(t *testing.T) {
	asm := newAssembly(t)

	val, err := asm.FindGlobal(testGlobalIntFQN)
	if err != nil {
		t.Fatalf("FindGlobal int: %v", err)
	}
	if val.Int() != int64(testGlobalInt) {
		t.Fatalf("FindGlobal int: got %d, want %d", val.Int(), testGlobalInt)
	}
}

func TestFindGlobal_String(t *testing.T) {
	asm := newAssembly(t)

	val, err := asm.FindGlobal("github.com/go-hotfix/assembly.testGlobalString")
	if err != nil {
		t.Fatalf("FindGlobal string: %v", err)
	}
	if val.String() != testGlobalString {
		t.Fatalf("FindGlobal string: got %q, want %q", val.String(), testGlobalString)
	}
}

// Pointer global — triggers godwarfTypeName PtrType branch in loadGlobals
func TestFindGlobal_Ptr(t *testing.T) {
	asm := newAssembly(t)

	val, err := asm.FindGlobal(testGlobalPtrFQN)
	if err != nil {
		t.Fatalf("FindGlobal ptr: %v", err)
	}
	if val.IsNil() {
		t.Log("FindGlobal ptr: nil pointer")
	}
}

// Alias type global — triggers resolveTypedef TypedefType via loadGlobals
func TestFindGlobal_Alias(t *testing.T) {
	asm := newAssembly(t)

	val, err := asm.FindGlobal(testGlobalAliasFQN)
	if err != nil {
		t.Fatalf("FindGlobal alias: %v", err)
	}
	if val.Int() != int64(testGlobalAlias) {
		t.Fatalf("FindGlobal alias: got %d, want %d", val.Int(), testGlobalAlias)
	}
}

// ============================================================================
// Plugins — cover stop-early
// ============================================================================

func TestPlugins_StopEarly(t *testing.T) {
	asm := newAssembly(t)

	count := 0
	for range asm.Plugins() {
		count++
		break
	}
	if count == 0 {
		t.Skip("Plugins: no shared libraries found on this platform")
	}
	if count != 1 {
		t.Fatalf("Plugins: break should stop after 1, got %d", count)
	}
}

// ============================================================================
// Funcs — cover finding specific function
// ============================================================================

func TestFuncs_FindsTestAdd(t *testing.T) {
	asm := newAssembly(t)

	found := false
	for name := range asm.Funcs() {
		if name == testAddFQN {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Funcs: testAdd not found")
	}
}

// ============================================================================
// WriteMemory — cover the unsupported path
// ============================================================================

func TestWriteMemory_Unsupported(t *testing.T) {
	var mem localMemory
	_, err := mem.WriteMemory(0, []byte{0x90})
	if err != ErrNotSupport {
		t.Fatalf("WriteMemory: got err=%v, want ErrNotSupport", err)
	}
}

// ============================================================================
// FindType — cover alias type
// ============================================================================

func TestFindType_Alias(t *testing.T) {
	asm := newAssembly(t)

	rtyp, err := asm.FindType(typeTestAlias)
	if err != nil {
		t.Fatalf("FindType(%s): %v", typeTestAlias, err)
	}
	if rtyp == nil {
		t.Fatal("FindType: got nil type")
	}
}

// ============================================================================
// FindFuncValue for no-arg function
// ============================================================================

func TestFindFuncValue_Void(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testVoidFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue void: %v", err)
	}

	results := fn.Call(nil)
	if results[0].Int() != 42 {
		t.Fatalf("FindFuncValue void call: got %d, want 42", results[0].Int())
	}
}

// ============================================================================
// CreateFuncForCodePtr — different type signatures
// ============================================================================

func TestCreateFuncForCodePtr_MultipleReturns(t *testing.T) {
	asm := newAssembly(t)

	entry, err := asm.FindFunc(testDivModFQN)
	if err != nil {
		t.Fatalf("FindFunc error: %v", err)
	}

	fnType := reflect.TypeOf(testDivMod)
	fn := CreateFuncForCodePtr(fnType, entry.Entry)

	results := fn.Call([]reflect.Value{reflect.ValueOf(17), reflect.ValueOf(5)})
	if results[0].Int() != 3 || results[1].Int() != 2 {
		t.Fatalf("CreateFuncForCodePtr divmod: got (%d, %d), want (3, 2)", results[0].Int(), results[1].Int())
	}
}

// ============================================================================
// Close — double close safety
// ============================================================================

func TestDoubleClose(t *testing.T) {
	asm, err := NewDwarfAssembly()
	if err != nil {
		t.Fatalf("NewDwarfAssembly: %v", err)
	}
	if err = asm.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	_ = asm.Close() // second close must not panic
}

// ============================================================================
// FindFuncType — cover error from getFunctionArgTypes
// ============================================================================

func TestFindFuncType_GenericInstantiation(t *testing.T) {
	asm := newAssembly(t)

	ftyp, err := asm.FindFuncType("github.com/go-hotfix/assembly.genericMin[int]", true)
	if err != nil {
		t.Fatalf("FindFuncType generic: %v", err)
	}
	if ftyp == nil {
		t.Fatal("FindFuncType generic: nil type")
	}
}

// ============================================================================
// FindFuncValue — cover void function
// ============================================================================

func TestFindFuncValue_NoArg(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testVoidFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue no-arg: %v", err)
	}
	if fn.IsNil() {
		t.Fatal("FindFuncValue no-arg: nil function")
	}
}

// ============================================================================
// ReadMemory — verify localMemory ReadMemory works
// ============================================================================

func TestReadMemory_LocalMemory(t *testing.T) {
	var mem localMemory
	var val int = 0x42424242
	buf := make([]byte, 4)
	n, err := mem.ReadMemory(buf, uint64(uintptr(unsafe.Pointer(&val))))
	if err != nil {
		t.Fatalf("ReadMemory: %v", err)
	}
	if n != 4 {
		t.Fatalf("ReadMemory: got %d bytes, want 4", n)
	}
}
