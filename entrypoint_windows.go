package assembly

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func getEntrypoint(targetModulePath string) (uintptr, error) {

	processHandle := windows.CurrentProcess()

	var modules [1024]windows.Handle
	var needed uint32
	if err := windows.EnumProcessModules(processHandle, &modules[0], uint32(unsafe.Sizeof(modules[0]))*1024, &needed); err != nil {
		return 0, err
	}

	var moduleList []string
	var modulePathUTF16Bytes [windows.MAX_PATH]uint16
	var count = needed / uint32(unsafe.Sizeof(modules[0]))

	for i := uint32(0); i < count; i++ {
		var mi windows.ModuleInfo
		if err := windows.GetModuleInformation(processHandle, modules[i], &mi, uint32(unsafe.Sizeof(mi))); err != nil {
			return 0, err
		}

		if err := windows.GetModuleFileNameEx(processHandle, modules[i], &modulePathUTF16Bytes[0], windows.MAX_PATH); err != nil {
			return 0, err
		}

		var modulePath = syscall.UTF16ToString(modulePathUTF16Bytes[:])

		if targetModulePath == modulePath {
			return mi.EntryPoint, nil
		}

		moduleList = append(moduleList, modulePath)
	}

	return 0, fmt.Errorf("module not found: %s not found in [%s]", targetModulePath, moduleList)
}
