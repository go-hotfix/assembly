//go:build !windows && !darwin

package assembly

func getEntrypoint(targetModulePath string) (uintptr, error) {
	return 0, nil
}
