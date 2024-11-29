//go:build !windows

package assembly

func getEntrypoint(targetModulePath string) (uintptr, error) {
	return 0, nil
}
