package assembly

import (
	"cmp"
	"fmt"
	"reflect"
	"testing"
)

func testAdd(a, b int) int {
	return a + b
}

func testMax(a int, nums ...int) int {
	if len(nums) == 0 {
		return a
	}

	var _max = a
	for _, n := range nums {
		if n > _max {
			_max = n
		}
	}
	return _max
}

func genericMin[T cmp.Ordered](a T, nums ...T) T {
	if len(nums) == 0 {
		return a
	}

	var _min = a
	for _, n := range nums {
		if n < _min {
			_min = n
		}
	}
	return _min
}

var testGlobalInt = 11001
var testGlobalString = "hello world"

func TestDwarfAssembly(t *testing.T) {

	asm, err := NewDwarfAssembly()
	if nil != err {
		t.Fatalf("NewDwarfAssembly() error: %v", err)
	}

	if nil == asm.BinaryInfo() {
		t.Fatalf("asm.BinaryInfo() is nil")
	}

	type TestCaseFunc func(t *testing.T, asm DwarfAssembly)

	var testCases = []TestCaseFunc{
		AssemblyTestFindType,
		AssemblyTestFindFunc,
		AssemblyTestFindVariadicFunc,
		AssemblyTestFindGenericVariadicFunc,
		AssemblyTestGlobalVar,
		AssemblyTestPlugin,
	}

	for _, testCase := range testCases {
		testCase(t, asm)
	}

	if err = asm.Close(); nil != err {
		t.Fatalf("DwarfAssembly.Close error: %v", err)
	}

}

func AssemblyTestFindType(t *testing.T, asm DwarfAssembly) {
	var found = false
	var err = asm.ForeachType(func(name string) bool {
		found = "github.com/go-hotfix/assembly.dwarfAssembly" == name
		return !found
	})

	if nil != err {
		t.Fatalf("ForeachType() error: %v", err)
	}

	if !found {
		t.Fatalf("ForeachType() not found")
	}

	asmType, err := asm.FindType("github.com/go-hotfix/assembly.dwarfAssembly")
	if nil != err {
		t.Fatalf("FindType() error: %v", err)
	}

	wantType := reflect.TypeOf(dwarfAssembly{})

	if wantType != asmType {
		t.Fatalf("FindType() got = %v, want %v", asmType, wantType)
	}
}

func AssemblyTestFindFunc(t *testing.T, asm DwarfAssembly) {

	asmType, err := asm.FindFuncType("github.com/go-hotfix/assembly.testAdd", false)
	if nil != err {
		t.Fatalf("FindFuncType() error: %v", err)
	}

	wantType := reflect.TypeOf(testAdd)

	if wantType != asmType {
		t.Fatalf("FindFuncType() got = %v, want %v", asmType, wantType)
	}

	callResults, err := asm.CallFunc("github.com/go-hotfix/assembly.testAdd", false, []reflect.Value{reflect.ValueOf(100), reflect.ValueOf(1)})
	if nil != err {
		t.Fatalf("CallFunc(%s) error: %v", wantType.String(), err)
	}

	wantValue := testAdd(100, 1)
	gotValue := callResults[0].Int()

	if int64(wantValue) != gotValue {
		t.Fatalf("CallFunc(%s) got = %v, want %v", wantType.String(), gotValue, wantValue)
	}

}

func AssemblyTestFindVariadicFunc(t *testing.T, asm DwarfAssembly) {
	asmType, err := asm.FindFuncType("github.com/go-hotfix/assembly.testMax", true)
	if nil != err {
		t.Fatalf("FindFuncType() error: %v", err)
	}

	wantType := reflect.TypeOf(testMax)

	if wantType != asmType {
		t.Fatalf("FindFuncType() got = %v, want %v", asmType, wantType)
	}

	callResults, err := asm.CallFunc("github.com/go-hotfix/assembly.testMax", false, []reflect.Value{reflect.ValueOf(100), reflect.ValueOf([]int{1})})
	if nil != err {
		t.Fatalf("CallFunc(%s) error: %v", wantType.String(), err)
	}

	wantValue := testMax(100, 1)
	gotValue := callResults[0].Int()

	if int64(wantValue) != gotValue {
		t.Fatalf("CallFunc(%s) got = %v, want %v", wantType.String(), gotValue, wantValue)
	}
}

func AssemblyTestFindGenericVariadicFunc(t *testing.T, asm DwarfAssembly) {
	asmType, err := asm.FindFuncType("github.com/go-hotfix/assembly.genericMin[int]", true)
	if nil != err {
		t.Fatalf("FindFuncType() error: %v", err)
	}

	wantType := reflect.TypeOf(genericMin[int])

	if wantType != asmType {
		t.Fatalf("FindFuncType() got = %v, want %v", asmType, wantType)
	}

	callResults, err := asm.CallFunc("github.com/go-hotfix/assembly.genericMin[int]", false, []reflect.Value{reflect.ValueOf(100), reflect.ValueOf([]int{1})})
	if nil != err {
		t.Fatalf("CallFunc(%s) error: %v", wantType.String(), err)
	}

	wantValue := genericMin(100, 1)
	gotValue := callResults[0].Int()

	if int64(wantValue) != gotValue {
		t.Fatalf("CallFunc(%s) got = %v, want %v", wantType.String(), gotValue, wantValue)
	}
}

func AssemblyTestGlobalVar(t *testing.T, asm DwarfAssembly) {

	var wantIntValue = int64(testGlobalInt + 1)
	var wantGlobalString = testGlobalString + "!"

	asm.ForeachGlobal(func(name string, value reflect.Value) bool {
		switch name {
		case "github.com/go-hotfix/assembly.testGlobalInt":
			value.SetInt(wantIntValue)
		case "github.com/go-hotfix/assembly.testGlobalString":
			value.SetString(wantGlobalString)
		}
		return true
	})

	globalIntValue, err := asm.FindGlobal("github.com/go-hotfix/assembly.testGlobalInt")
	if nil != err {
		t.Fatalf("FindGlobal() error: %v", err)
	}

	if globalIntValue.Int() != wantIntValue {
		t.Fatalf("testGlobalInt got = %v, want %v", globalIntValue.Int(), wantIntValue)
	}

	globalStringValue, err := asm.FindGlobal("github.com/go-hotfix/assembly.testGlobalString")
	if nil != err {
		t.Fatalf("testGlobalString error: %v", err)
	}

	if globalStringValue.String() != wantGlobalString {
		t.Fatalf("testGlobalString got = %v, want %v", globalStringValue.String(), wantGlobalString)
	}

}

func AssemblyTestPlugin(t *testing.T, asm DwarfAssembly) {

	libs, addrs, err := asm.SearchPlugins()
	if nil != err {
		t.Fatalf("SearchPlugins() error: %v", err)
	}

	if len(libs) != len(addrs) {
		t.Fatalf("len(libs) != len(addrs)")
	}

	for i, lib := range libs {
		if len(lib) > 0 {
			fmt.Println(lib, addrs[i])
		}
	}

	_, _, err = asm.SearchPluginByName("not-found-image")
	if err != ErrNotFound {
		t.Fatalf("SearchPluginByName failed")
	}
}
