package openwrt

/*
#include <uci.h>
*/
import "C"
import (
	"fmt"
	"os"
	"path"
)

const (
	UCI_CONFIG_FOLDER = "/etc/config"
)

type UciContext struct {
	ptr *C.struct_uci_context
}

type UciPackage struct {
	ptr *C.struct_uci_package
}

type UciSection struct {
	ptr *C.struct_uci_section
}

type UciOption struct {
	ptr *C.struct_uci_option
}

type UciElement struct {
	ptr *C.struct_uci_element
}

type UciPtr struct {
	ptr *C.struct_uci_ptr
}

func NewUciContext() *UciContext {
	return &UciContext{
		ptr: C.uci_alloc_context(),
	}
}

func (ctx *UciContext) Free() {
	C.uci_free_context(ctx.ptr)
}

func (ctx *UciContext) AddPackage(name string) error {
	config := path.Join(UCI_CONFIG_FOLDER, name)
	if _, err := os.Stat(config); err == nil {
		return nil
	}

	file, err := os.Create(config)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	return err
}

func (ctx *UciContext) Set(ptr *UciPtr) error {
	ret, err := C.uci_set(ctx.ptr, ptr.ptr)
	if err != nil {
		return err
	}
	if ret != C.UCI_OK {
		return fmt.Errorf("%d: %s", int(ret), ctx.ErrorString(""))
	}

	return nil
}

func (ctx *UciContext) AddSection(packageName string, sectionName string, sectionType string) error {
	cPackageName := C.CString(packageName)
	var ptr C.struct_uci_ptr
	ptr._package = cPackageName
	return nil
}

func (ctx *UciContext) ErrorString(prefix string) string {
	// TODO
	return ""
}
