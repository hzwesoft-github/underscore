package openwrt

/*
#include <uci.h>
#include <stdlib.h>

static void list_sections(struct uci_package *package, struct uci_section **section, int *section_len)
{
	int i;
	struct uci_element *element;

	i = 0;
	uci_foreach_element(&package->sections, element)
  {
		i++;
  }

	struct uci_section **ptr = calloc(i, sizeof(struct uci_section*));

	i = 0;
	uci_foreach_element(&package->sections, element)
  {
		ptr[i++] = uci_to_section(element);
  }

	section = ptr;
	*section_len = i;
}

static void option_str_value(struct uci_option *option, char *value)
{

}

static void option_list_value(struct uci_option *option, char **list, int *list_len)
{

}

*/
import "C"
import (
	"fmt"
	"os"
	"path"
	"unsafe"
)

type UciOptonType int

const (
	UCI_TYPE_STRING UciOptonType = iota
	UCI_TYPE_LIST
)

const (
	UCI_CONFIG_FOLDER = "/etc/config"
)

type UciContext struct {
	ptr *C.struct_uci_context
}

type UciPackage struct {
	Name string

	ptr    *C.struct_uci_package
	parent *UciContext
}

type UciSection struct {
	Name string

	ptr    *C.struct_uci_section
	parent *UciPackage
}

type UciOption struct {
	Type   UciOptonType
	Name   string
	Value  string
	Values []string

	ptr    *C.struct_uci_option
	parent *UciSection
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

	return &UciPackage{name, cpackage, ctx}
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

	return &UciSection{name, csection, pkg}
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

	name := C.GoString(csection.e.name)

	return &UciSection{name, csection, pkg}
}

func (pkg *UciPackage) DelSection(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var uciptr C.struct_uci_ptr
	uciptr._package = pkg.ptr.e.name
	uciptr.section = cname

	return pkg.parent.uci_delete(&uciptr)
}

func (pkg *UciPackage) DelUnnamedSection(section *UciSection) error {
	var uciptr C.struct_uci_ptr
	uciptr._package = pkg.ptr.e.name
	uciptr.section = section.ptr.e.name

	return pkg.parent.uci_delete(&uciptr)
}

// FIXME
func (pkg *UciPackage) ListSections() []UciSection {
	var csections *C.struct_uci_section
	var clength *C.int
	C.list_sections(pkg.ptr, &csections, clength)

	sectionPtr := unsafe.Pointer(csections)
	defer C.free(sectionPtr)
	length := int(*clength)

	sectionArray := (*[1 << 10]C.struct_uci_section)(sectionPtr)
	slice := sectionArray[0:length:length]

	sections := make([]UciSection, 0)
	for _, v := range slice {
		name := C.GoString(v.e.name)
		sections = append(sections, UciSection{name, &v, pkg})
	}

	return sections
}

// * UciSection

// TODO
func (section *UciSection) LoadOption(name string) *UciOption {
	coption := section.parent.parent.uci_lookup_option(section.ptr, name)
	if coption == nil {
		return nil
	}

	option := &UciOption{}
	option.Name = name
	if coption._type == C.UCI_TYPE_STRING {
		option.Type = UCI_TYPE_STRING

	} else if coption._type == C.UCI_TYPE_LIST {
		option.Type = UCI_TYPE_LIST

	} else {
		return nil
	}

	return option
}

func (ctx *UciContext) uci_set(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_set(ctx.ptr, ptr)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_delete(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_delete(ctx.ptr, ptr)
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
