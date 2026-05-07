//go:build darwin

package assembly

import "fmt"

var errNoEntryPoint = fmt.Errorf("failed to get image header for main executable")

func getEntrypoint(targetModulePath string) (uintptr, error) {
	header := dyldGetImageHeader(0)
	if header == 0 {
		return 0, errNoEntryPoint
	}
	return header, nil
}
