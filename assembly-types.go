package assembly

import (
	"debug/dwarf"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/go-delve/delve/pkg/dwarf/godwarf"
	"github.com/go-delve/delve/pkg/proc"
)

func (da *dwarfAssembly) ForeachType(f func(name string) bool) error {
	types, err := da.binaryInfo.Types()
	if err != nil {
		return err
	}
	for _, name := range types {
		if !f(name) {
			break
		}
	}
	return nil
}

func (da *dwarfAssembly) FindType(name string) (reflect.Type, error) {
	dwarfType, err := findType(da.binaryInfo, name)
	if err != nil {
		return nil, err
	}

	typeAddr, err := da.dwarfToRuntimeType(dwarfType, name)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(*(*interface{})(unsafe.Pointer(&typeAddr)))
	return typ, nil
}

func (da *dwarfAssembly) findImageType(img *proc.Image, name string) uint64 {
	if da.imageTypes == nil {
		da.imageTypes = make(map[*proc.Image]map[string]uint64)
	}
	cache, ok := da.imageTypes[img]
	if !ok {
		cache = make(map[string]uint64)
		da.imageTypes[img] = cache

		reader := img.DwarfReader()
		md := imageToModuleData(da.binaryInfo, img, da.modules)
		if md == nil {
			return 0
		}

		rRuntimeTypes := reflect.ValueOf(img).Elem().FieldByName("runtimeTypeToDIE")
		iter := rRuntimeTypes.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()

			offset := v.FieldByName("offset").Uint()
			reader.Seek(dwarf.Offset(offset))
			entry, err := reader.Next()
			if err != nil || entry == nil {
				continue
			}
			entryName, ok := entry.Val(dwarf.AttrName).(string)
			if !ok {
				continue
			}
			if k.Uint() == 0 {
				continue
			}

			typeAddr := md.types + k.Uint()
			if typeAddr < md.types || typeAddr >= md.etypes {
				cache[entryName] = k.Uint()
			} else {
				cache[entryName] = typeAddr
			}
		}
	}

	return cache[name]
}

func (da *dwarfAssembly) dwarfToRuntimeType(typ godwarf.Type, name string) (typeAddr uint64, err error) {
	bi := da.binaryInfo
	mds := da.modules

	if typ.Common().Index >= len(bi.Images) {
		return 0, fmt.Errorf("could not find image for type %s", name)
	}
	img := bi.Images[typ.Common().Index]
	rdr := img.DwarfReader()
	rdr.Seek(typ.Common().Offset)
	e, err := rdr.Next()
	if err != nil {
		return 0, fmt.Errorf("could not find dwarf entry for type:%s err:%s", name, err)
	}
	entryName, ok := e.Val(dwarf.AttrName).(string)
	if !ok || entryName != name {
		return 0, fmt.Errorf("could not find name for type:%s entry:%s", name, entryName)
	}
	off, ok := e.Val(godwarf.AttrGoRuntimeType).(uint64)
	if !ok || off == 0 {
		for i, img := range bi.Images {
			if i == 0 {
				continue
			}
			addr := da.findImageType(img, name)
			if addr != 0 {
				return addr, nil
			}
		}
		return 0, fmt.Errorf("could not find runtime type for type:%s", name)
	}

	md := imageToModuleData(bi, img, mds)
	if md == nil {
		return 0, fmt.Errorf("could not find module data for type %s", name)
	}

	typeAddr = md.types + off
	if typeAddr < md.types || typeAddr >= md.etypes {
		return off, nil
	}
	return typeAddr, nil
}
