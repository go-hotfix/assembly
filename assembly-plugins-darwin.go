//go:build darwin

package assembly

import (
	"iter"
	"strings"
	"unsafe"
)

func (da *dwarfAssembly) FindPlugin(name string) (string, uint64, error) {
	libs, addrs := da.loadPlugins()
	for i := 0; i < len(libs); i++ {
		if strings.LastIndex(libs[i], name) >= 0 {
			return libs[i], addrs[i], nil
		}
	}
	return "", 0, ErrNotFound
}

func (da *dwarfAssembly) Plugins() iter.Seq[string] {
	return func(yield func(string) bool) {
		libs, _ := da.loadPlugins()
		for _, name := range libs {
			if !yield(name) {
				return
			}
		}
	}
}

func (da *dwarfAssembly) loadPlugins() (libs []string, addrs []uint64) {
	count := dyldImageCount()

	for i := uint32(0); i < count; i++ {
		namePtr := dyldGetImageName(i)
		header := dyldGetImageHeader(i)
		if header == 0 {
			continue
		}
		libs = append(libs, goString(namePtr))
		addrs = append(addrs, uint64(header))
	}

	return libs, addrs
}

func goString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var buf []byte
	for {
		b := *(*byte)(unsafe.Add(pointerOf(ptr), len(buf)))
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf)
}
