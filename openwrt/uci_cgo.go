package openwrt

/*
#include <uci.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"path"
	"unsafe"
)

const (
	UCI_CONFIG_FOLDER = "/etc/config"
)

type UciContext struct {
	ptr *C.struct_uci_context
}

type UciPackage struct {
	ptr    *C.struct_uci_package
	parent *UciContext
}

type UciSection struct {
	ptr    *C.struct_uci_section
	parent *UciPackage
}

type UciOption struct {
	ptr    *C.struct_uci_option
	parent *UciSection
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

// * UciContext

func (ctx *UciContext) LoadPackage(name string) *UciPackage {
	cpackage, err := ctx.uci_load(name)
	if err != nil {
		return nil
	}

	return &UciPackage{cpackage, ctx}
}

// ~ FIXME
func (ctx *UciContext) LookupPtr(str string) *UciPtr {
	ptr, err := ctx.uci_lookup_ptr(str)
	if err != nil {
		return nil
	}
	return &UciPtr{ptr}
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

func (ctx *UciContext) DelPackage(name string) error {
	config := path.Join(UCI_CONFIG_FOLDER, name)
	if _, err := os.Stat(config); err != nil {
		return nil
	}

	return os.Remove(config)
}

// * UciPackage

func (pkg *UciPackage) Unload() error {
	return pkg.parent.uci_unload(pkg.ptr)
}

func (pkg *UciPackage) LoadSection(name string) *UciSection {
	csection := pkg.parent.uci_lookup_section(pkg.ptr, name)
	if csection == nil {
		return nil
	}

	return &UciSection{csection, pkg}
}

func (pkg *UciPackage) AddSection(name string, typ string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ctype := C.CString(typ)
	defer C.free(unsafe.Pointer(ctype))

	var uciptr C.struct_uci_ptr
	uciptr._package = pkg.ptr.e.name
	uciptr.section = cname
	uciptr.value = ctype

	return pkg.parent.uci_set(&uciptr)
}

func (pkg *UciPackage) AddUnnamedSection(typ string) *UciSection {
	ctype := C.CString(typ)
	defer C.free(unsafe.Pointer(ctype))

	csection, err := pkg.parent.uci_add_section(pkg.ptr, ctype)
	if err != nil {
		return nil
	}

	return &UciSection{csection, pkg}
}

func (pkg *UciPackage) DelSection(name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var uciptr C.struct_uci_ptr
	uciptr._package = pkg.ptr.e.name
	uciptr.section = cname

	// FIXME

}

// * UciSection

func (ctx *UciContext) Set(ptr *UciPtr) error {
	ret, err := C.uci_set(ctx.ptr, ptr.ptr)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_set(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_set(ctx.ptr, ptr)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_add_list(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_add_list(ctx.ptr, ptr)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_add_section(pkg *C.struct_uci_package, typ *C.char) (section *C.struct_uci_section, err error) {
	var ret C.int
	ret, err = C.uci_add_section(ctx.ptr, pkg, typ, &section)
	err = ctx.uci_ret_to_error(ret, err)
	if err != nil {
		return nil, err
	}

	return section, nil
}

func (ctx *UciContext) uci_commit(pkg *C.struct_uci_package, overwrite bool) error {
	ret, err := C.uci_commit(ctx.ptr, &pkg, C.bool(overwrite))
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_load(name string) (pkg *C.struct_uci_package, err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var ret C.int
	ret, err = C.uci_load(ctx.ptr, cname, &pkg)
	err = ctx.uci_ret_to_error(ret, err)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (ctx *UciContext) uci_unload(pkg *C.struct_uci_package) error {
	ret, err := C.uci_unload(ctx.ptr, pkg)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_lookup_package(name string) *C.struct_uci_package {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.uci_lookup_package(ctx.ptr, cname)
}

func (ctx *UciContext) uci_lookup_section(pkg *C.struct_uci_package, name string) *C.struct_uci_section {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.uci_lookup_section(ctx.ptr, pkg, cname)
}

func (ctx *UciContext) uci_lookup_option(section *C.struct_uci_section, name string) *C.struct_uci_option {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.uci_lookup_option(ctx.ptr, section, cname)
}

func (ctx *UciContext) uci_lookup_ptr(str string) (*C.struct_uci_ptr, error) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var ptr C.struct_uci_ptr
	ret, err := C.uci_lookup_ptr(ctx.ptr, &ptr, cstr, true)
	err = ctx.uci_ret_to_error(ret, err)
	if err != nil {
		return nil, err
	}

	return &ptr, nil
}

func (ctx *UciContext) uci_ret_to_error(ret C.int, err error) error {
	if err != nil {
		return err
	}
	if ret != C.UCI_OK {
		return fmt.Errorf("%d: %s", int(ret), ctx.ErrorString(""))
	}

	return nil
}

func (ctx *UciContext) ErrorString(prefix string) string {
	// TODO
	return ""
}
