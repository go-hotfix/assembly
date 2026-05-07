//go:build darwin

package assembly

//go:cgo_import_dynamic _dyld_image_count _dyld_image_count "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _dyld_get_image_name _dyld_get_image_name "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _dyld_get_image_header _dyld_get_image_header "/usr/lib/libSystem.B.dylib"

func dyldImageCount() uint32
func dyldGetImageName(index uint32) uintptr
func dyldGetImageHeader(index uint32) uintptr
