package assembly

import (
	"cmp"
	"os"
	"reflect"
	"testing"
)

// ============================================================================
// Test helper functions
// ============================================================================

//go:noinline
func testSub(a, b int) int { return a - b }

//go:noinline
func testMul(a, b int) int { return a * b }

//go:noinline
func testDivMod(a, b int) (int, int) { return a / b, a % b }

//go:noinline
func testConcat(a, b string) string { return a + b }

//go:noinline
func testSwap(a, b int) (int, int) { return b, a }

var testGlobalFloat = 3.14

var sharedAssembly DwarfAssembly

func TestMain(m *testing.M) {
	var err error
	sharedAssembly, err = NewDwarfAssembly()
	if err != nil {
		panic("NewDwarfAssembly: " + err.Error())
	}
	code := m.Run()
	sharedAssembly.Close()
	os.Exit(code)
}

func newAssembly(t *testing.T) DwarfAssembly {
	t.Helper()
	return sharedAssembly
}

const (
	testAddFQN    = "github.com/go-hotfix/assembly.testAdd"
	testSubFQN    = "github.com/go-hotfix/assembly.testSub"
	testMulFQN    = "github.com/go-hotfix/assembly.testMul"
	testDivModFQN = "github.com/go-hotfix/assembly.testDivMod"
	testConcatFQN = "github.com/go-hotfix/assembly.testConcat"
	testSwapFQN   = "github.com/go-hotfix/assembly.testSwap"
	testBarFQN    = "github.com/go-hotfix/assembly.foo.bar"

	testGlobalIntFQN   = "github.com/go-hotfix/assembly.testGlobalInt"
	testGlobalFloatFQN = "github.com/go-hotfix/assembly.testGlobalFloat"

	typeDwarfAssembly = "github.com/go-hotfix/assembly.dwarfAssembly"
	typeFoo           = "github.com/go-hotfix/assembly.foo"
)

// ============================================================================
// Function Discovery
// ============================================================================

func TestFuncs(t *testing.T) {
	asm := newAssembly(t)

	var found []string
	for name := range asm.Funcs() {
		if name == testAddFQN {
			found = append(found, name)
		}
		if name == testBarFQN {
			found = append(found, name)
		}
	}

	if len(found) < 2 {
		t.Fatalf("Funcs: expected at least testAdd and foo.bar, got %v", found)
	}
}

func TestFuncs_StopEarly(t *testing.T) {
	asm := newAssembly(t)

	count := 0
	for range asm.Funcs() {
		count++
		break
	}

	if count != 1 {
		t.Fatalf("Funcs: break should stop after 1 iteration, got %d", count)
	}
}

func TestFindFunc(t *testing.T) {
	asm := newAssembly(t)

	entry, err := asm.FindFunc(testAddFQN)
	if err != nil {
		t.Fatalf("FindFunc(%s) error: %v", testAddFQN, err)
	}
	if entry.Entry == 0 {
		t.Fatal("FindFunc: Entry is 0")
	}
	if entry.End <= entry.Entry {
		t.Fatalf("FindFunc: End(%#x) <= Entry(%#x)", entry.End, entry.Entry)
	}
}

func TestFindFunc_Pc(t *testing.T) {
	asm := newAssembly(t)

	entry, err := asm.FindFunc(testAddFQN)
	if err != nil {
		t.Fatalf("FindFunc(%s) error: %v", testAddFQN, err)
	}
	if entry.Entry == 0 {
		t.Fatal("FindFunc: Entry is 0")
	}
}

func TestFindFunc_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.FindFunc("nonexistent.Func")
	if err == nil {
		t.Fatal("FindFunc: expected error for non-existent function")
	}
}

// ============================================================================
// FindFuncValue (callable reflect.Value)
// ============================================================================

func TestFindFuncValue(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testSubFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue(%s) error: %v", testSubFQN, err)
	}

	results := fn.Call([]reflect.Value{reflect.ValueOf(10), reflect.ValueOf(3)})
	if len(results) != 1 {
		t.Fatalf("FindFuncValue: expected 1 result, got %d", len(results))
	}
	if results[0].Int() != 7 {
		t.Fatalf("FindFuncValue call: got %d, want 7", results[0].Int())
	}
}

func TestFindFuncValue_MultipleReturns(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testDivModFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue(%s) error: %v", testDivModFQN, err)
	}

	results := fn.Call([]reflect.Value{reflect.ValueOf(17), reflect.ValueOf(5)})
	if len(results) != 2 {
		t.Fatalf("FindFuncValue: expected 2 results, got %d", len(results))
	}
	if results[0].Int() != 3 {
		t.Fatalf("FindFuncValue div: got %d, want 3", results[0].Int())
	}
	if results[1].Int() != 2 {
		t.Fatalf("FindFuncValue mod: got %d, want 2", results[1].Int())
	}
}

func TestFindFuncValue_StringParams(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testConcatFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue(%s) error: %v", testConcatFQN, err)
	}

	results := fn.Call([]reflect.Value{reflect.ValueOf("hello "), reflect.ValueOf("world")})
	if results[0].String() != "hello world" {
		t.Fatalf("FindFuncValue concat: got %q, want %q", results[0].String(), "hello world")
	}
}

func TestFindFuncValue_Swap(t *testing.T) {
	asm := newAssembly(t)

	fn, err := asm.FindFuncValue(testSwapFQN, false)
	if err != nil {
		t.Fatalf("FindFuncValue(%s) error: %v", testSwapFQN, err)
	}

	results := fn.Call([]reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2)})
	if results[0].Int() != 2 || results[1].Int() != 1 {
		t.Fatalf("FindFuncValue swap: got (%d, %d), want (2, 1)", results[0].Int(), results[1].Int())
	}
}

func TestFindFuncValue_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.FindFuncValue("nonexistent.Func", false)
	if err == nil {
		t.Fatal("FindFuncValue: expected error for non-existent function")
	}
}

// ============================================================================
// CallFunc variations
// ============================================================================

func TestCallFunc_WrongArgType(t *testing.T) {
	asm := newAssembly(t)

	// CallFunc validates type assignability before calling.
	// Passing a string where int is expected should fail.
	_, err := asm.CallFunc(testAddFQN, false, []reflect.Value{
		reflect.ValueOf("not_an_int"),
		reflect.ValueOf(1),
	})
	if err == nil {
		t.Fatal("CallFunc: expected error for wrong argument type")
	}
}

func TestCallFunc_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.CallFunc("nonexistent.Func", false, nil)
	if err == nil {
		t.Fatal("CallFunc: expected error for non-existent function")
	}
}

func TestCallFunc_Mul(t *testing.T) {
	asm := newAssembly(t)

	results, err := asm.CallFunc(testMulFQN, false, []reflect.Value{
		reflect.ValueOf(6),
		reflect.ValueOf(7),
	})
	if err != nil {
		t.Fatalf("CallFunc(%s) error: %v", testMulFQN, err)
	}
	if results[0].Int() != 42 {
		t.Fatalf("CallFunc mul: got %d, want 42", results[0].Int())
	}
}

// ============================================================================
// Type Lookup
// ============================================================================

func TestFindType_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.FindType("nonexistent.Type")
	if err == nil {
		t.Fatal("FindType: expected error for non-existent type")
	}
}

func TestFindType_KnownStruct(t *testing.T) {
	asm := newAssembly(t)

	rt, err := asm.FindType(typeFoo)
	if err != nil {
		t.Fatalf("FindType(%s) error: %v", typeFoo, err)
	}
	want := reflect.TypeOf(foo{})
	if rt != want {
		t.Fatalf("FindType: got %v, want %v", rt, want)
	}
}

func TestFindFuncType_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.FindFuncType("nonexistent.Func", false)
	if err == nil {
		t.Fatal("FindFuncType: expected error for non-existent function")
	}
}

func TestFindFuncType_MatchesReflect(t *testing.T) {
	asm := newAssembly(t)

	tests := []struct {
		name     string
		fqn      string
		variadic bool
		wantType reflect.Type
	}{
		{testAddFQN, testAddFQN, false, reflect.TypeOf(testAdd)},
		{testSubFQN, testSubFQN, false, reflect.TypeOf(testSub)},
		{testMulFQN, testMulFQN, false, reflect.TypeOf(testMul)},
		{testDivModFQN, testDivModFQN, false, reflect.TypeOf(testDivMod)},
		{testConcatFQN, testConcatFQN, false, reflect.TypeOf(testConcat)},
		{testSwapFQN, testSwapFQN, false, reflect.TypeOf(testSwap)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := asm.FindFuncType(tt.fqn, tt.variadic)
			if err != nil {
				t.Fatalf("FindFuncType(%s) error: %v", tt.fqn, err)
			}
			if got != tt.wantType {
				t.Errorf("FindFuncType(%s) = %v, want %v", tt.fqn, got, tt.wantType)
			}
		})
	}
}

// ============================================================================
// ForeachType
// ============================================================================

func TestTypes_StopEarly(t *testing.T) {
	asm := newAssembly(t)

	count := 0
	for range asm.Types() {
		count++
		break
	}

	if count != 1 {
		t.Fatalf("Types: break should stop after 1 iteration, got %d", count)
	}
}

func TestTypes_FindsMultiple(t *testing.T) {
	asm := newAssembly(t)

	found := map[string]bool{}
	for name := range asm.Types() {
		if name == typeDwarfAssembly {
			found[typeDwarfAssembly] = true
		}
		if name == typeFoo {
			found[typeFoo] = true
		}
	}

	if !found[typeDwarfAssembly] || !found[typeFoo] {
		t.Fatalf("Types: expected to find both types, got %v", found)
	}
}

// ============================================================================
// Global Variables
// ============================================================================

func TestFindGlobal_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, err := asm.FindGlobal("nonexistent.Var")
	if err != ErrNotFound {
		t.Fatalf("FindGlobal: got err=%v, want ErrNotFound", err)
	}
}

func TestFindGlobal_Float(t *testing.T) {
	asm := newAssembly(t)

	val, err := asm.FindGlobal(testGlobalFloatFQN)
	if err != nil {
		t.Fatalf("FindGlobal(%s) error: %v", testGlobalFloatFQN, err)
	}

	// The value should match the current value of testGlobalFloat
	if val.Float() != testGlobalFloat {
		t.Fatalf("FindGlobal float: got %v, want %v", val.Float(), testGlobalFloat)
	}
}

func TestGlobals_Mutation(t *testing.T) {
	asm, err := NewDwarfAssembly()
	if err != nil {
		t.Fatalf("NewDwarfAssembly: %v", err)
	}
	defer asm.Close()

	newVal := int64(99999)
	for name, value := range asm.Globals() {
		if name == testGlobalIntFQN {
			value.SetInt(newVal)
			break
		}
	}

	got, err := asm.FindGlobal(testGlobalIntFQN)
	if err != nil {
		t.Fatalf("FindGlobal error: %v", err)
	}
	if got.Int() != newVal {
		t.Fatalf("Globals mutation: got %d, want %d", got.Int(), newVal)
	}

	// Restore original value
	testGlobalInt = int(newVal - 1)
}

// ============================================================================
// Plugin Search
// ============================================================================

func TestPlugins_ContainsKnownLib(t *testing.T) {
	asm := newAssembly(t)

	found := false
	for lib := range asm.Plugins() {
		if len(lib) > 0 {
			found = true
		}
	}
	if !found {
		t.Skip("Plugins: no shared libraries found on this platform")
	}
}

func TestFindPlugin_NotFound(t *testing.T) {
	asm := newAssembly(t)

	_, _, err := asm.FindPlugin("this_library_does_not_exist_12345")
	if err != ErrNotFound {
		t.Fatalf("FindPlugin: got err=%v, want ErrNotFound", err)
	}
}

// ============================================================================
// CreateFuncForCodePtr
// ============================================================================

func TestCreateFuncForCodePtr(t *testing.T) {
	asm := newAssembly(t)

	entry, err := asm.FindFunc(testAddFQN)
	if err != nil {
		t.Fatalf("FindFunc error: %v", err)
	}

	fnType := reflect.TypeOf(testAdd)
	fn := CreateFuncForCodePtr(fnType, entry.Entry)

	results := fn.Call([]reflect.Value{reflect.ValueOf(100), reflect.ValueOf(200)})
	if results[0].Int() != 300 {
		t.Fatalf("CreateFuncForCodePtr call: got %d, want 300", results[0].Int())
	}
}

// ============================================================================
// Generic function with cmp.Ordered
// ============================================================================

func TestCallFunc_GenericMinInt(t *testing.T) {
	asm := newAssembly(t)

	fqn := "github.com/go-hotfix/assembly.genericMin[int]"
	results, err := asm.CallFunc(fqn, false, []reflect.Value{
		reflect.ValueOf(100),
		reflect.ValueOf([]int{1}),
	})
	if err != nil {
		t.Fatalf("CallFunc(%s) error: %v", fqn, err)
	}
	if results[0].Int() != 1 {
		t.Fatalf("CallFunc genericMin[int]: got %d, want 1", results[0].Int())
	}
}

// Ensure genericMin is kept alive
var _ = genericMin[int]
var _ = cmp.Or[string]
