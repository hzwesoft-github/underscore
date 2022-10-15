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

static void list_options(struct uci_section *section, struct uci_option **option, int *option_len)
{
	int i;
	struct uci_element *element;

	i = 0;
	uci_foreach_element(&section->options, element)
  {
		i++;
  }

	struct uci_option **ptr = calloc(i, sizeof(struct uci_option*));

	i = 0;
	uci_foreach_element(&section->options, element)
  {
		ptr[i++] = uci_to_option(element);
  }

	option = ptr;
	*option_len = i;
}

static void option_str_value(struct uci_option *option, char *value)
{
	value = option->v.string;
}

static void option_list_value(struct uci_option *option, char **list, int *list_len)
{
	int i;
	struct uci_element *element = NULL;

	i = 0;
	uci_foreach_element(&option->v.list, element)
  {
    i++;
  }

	char **ptr = calloc(i, sizeof(char*));
	i = 0;
	uci_foreach_element(&option->v.list, element)
  {
		ptr[i++] = element->name;
  }

	list = ptr;
	*list_len = i;
}

*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"unsafe"

	"github.com/hzwesoft-github/underscore/lang"
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

// params are package name, section name, option name and value
func (ctx *UciContext) build_uciptr(params ...string) (uciptr C.struct_uci_ptr, err error) {
	if len(params) == 0 {
		return uciptr, errors.New("ng: package must specified")
	}

	if len(params) == 1 {
		return uciptr, errors.New("ng: section must specified")
	}

	packageName := params[0]
	if lang.IsBlank(packageName) {
		return uciptr, errors.New("ng: package name must specified")
	}

	sectionName := params[1]
	if lang.IsBlank(sectionName) {
		return uciptr, errors.New("ng: section name must specified")
	}

	cPackageName := C.CString(packageName)
	defer C.free(unsafe.Pointer(cPackageName))

	cSectionName := C.CString(sectionName)
	defer C.free(unsafe.Pointer(cSectionName))

	uciptr._package = cPackageName
	uciptr.section = cSectionName

	if len(params) > 2 {
		optionName := params[2]

		if !lang.IsBlank(optionName) {
			cOptionName := C.CString(optionName)
			defer C.free(unsafe.Pointer(cOptionName))

			uciptr.option = cOptionName
		}
	}

	if len(params) > 3 {
		value := params[3]

		if !lang.IsBlank(value) {
			cValue := C.CString(value)
			defer C.free(unsafe.Pointer(cValue))

			uciptr.value = cValue
		}
	}

	return uciptr, nil
}

func (ctx *UciContext) take_effect(uciptr *C.struct_uci_ptr) (err error) {
	if err = ctx.uci_commit(uciptr.p, false); err != nil {
		return err
	}

	return ctx.uci_unload(uciptr.p)
}

func (ctx *UciContext) Set(packageName, sectionName, optionName, value string) error {
	uciptr, err := ctx.build_uciptr(packageName, sectionName, optionName, value)
	if err != nil {
		return err
	}

	err = ctx.uci_set(&uciptr)
	if err != nil {
		return err
	}

	return ctx.take_effect(&uciptr)
}

func (ctx *UciContext) Marshal(packageName, sectionName, sectionType string, obj any) error {
	if lang.IsBlank(packageName) || lang.IsBlank(sectionName) {
		return errors.New("ng: package or section must not be empty")
	}

	pkg := ctx.LoadPackage((packageName))
	if pkg == nil {
		ctx.AddPackage(packageName)

		pkg = ctx.LoadPackage(packageName)
		if pkg == nil {
			return fmt.Errorf("ng: package %s create error", packageName)
		}
	}
	defer pkg.Unload()

	// FIXME 增加有名和无名的Section
	section := pkg.LoadSection(sectionName)
	if section != nil {
		if err := pkg.DelSection(sectionName); err != nil {
			return nil
		}

		pkg.AddSection(sectionName, sectionType)
		return fmt.Errorf("ng: section %s.%s create error", packageName, sectionName)
	}

	typ := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct && typ.Kind() != reflect.Map {
		return errors.New("ng: obj must be struct or map")
	}

	switch typ.Kind() {
	case reflect.Struct:
		var optionName string

		for i := 0; i < typ.NumField(); i++ {
			omitEmpty := false
			field := typ.Field(i)

			tagValue := field.Tag.Get("uci")
			if tagValue == "-" {
				continue
			}

			value := val.Field(i)
			if value.Kind() != reflect.String || value.Kind() != reflect.Slice {
				return fmt.Errorf("ng: struct field value must string or []string: %s %s", typ.Name(), field.Name)
			}
			if value.Kind() == reflect.Slice && value.Elem().Kind() != reflect.String {
				return fmt.Errorf("ng: struct field value must string or []string: %s %s", typ.Name(), field.Name)
			}

			if tagValue == "" {
				optionName = field.Name
			} else {
				tags := strings.Split(tagValue, ",")
				if len(tags) == 0 || len(tags) > 2 {
					return fmt.Errorf("ng: tag format error: %s %s", typ.Name(), field.Name)
				}

				optionName = tags[0]
				if len(tags) > 1 {
					if tags[1] != "omitempty" {
						return fmt.Errorf("ng: tag format error: %s %s", typ.Name(), field.Name)
					}

					omitEmpty = true
				}
			}

			switch value.Kind() {
			case reflect.String:

			case reflect.Slice:

			}

		}
	case reflect.Map:
	}

	return nil
}

func (ctx *UciContext) Unmarshal(packageName, sectionName string, obj any) error {
	// TODO
	return nil
}

// * UciPackage

func (pkg *UciPackage) Unload() error {
	return pkg.parent.uci_unload(pkg.ptr)
}

func (pkg *UciPackage) Commit(overwrite bool) error {
	return pkg.parent.uci_commit(pkg.ptr, overwrite)
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

	uciptr.p = pkg.ptr
	// uciptr._package = pkg.ptr.e.name
	uciptr.section = cname
	uciptr.value = ctype

	return pkg.parent.uci_set(&uciptr)
}

func (pkg *UciPackage) AddUnnamedSection(typ string) (*UciSection, error) {
	ctype := C.CString(typ)
	defer C.free(unsafe.Pointer(ctype))

	csection, err := pkg.parent.uci_add_section(pkg.ptr, ctype)
	if err != nil {
		return nil, err
	}

	name := C.GoString(csection.e.name)

	return &UciSection{name, csection, pkg}, nil
}

func (pkg *UciPackage) DelSection(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var uciptr C.struct_uci_ptr

	uciptr.p = pkg.ptr
	// uciptr._package = pkg.ptr.e.name
	uciptr.section = cname

	return pkg.parent.uci_delete(&uciptr)
}

func (pkg *UciPackage) DelUnnamedSection(section *UciSection) error {
	var uciptr C.struct_uci_ptr

	uciptr.p = pkg.ptr
	// uciptr._package = pkg.ptr.e.name
	uciptr.section = section.ptr.e.name

	return pkg.parent.uci_delete(&uciptr)
}

// FIXME
func (pkg *UciPackage) ListSections() []UciSection {
	var csections *C.struct_uci_section
	var clength C.int

	C.list_sections(pkg.ptr, &csections, &clength)

	sectionPtr := unsafe.Pointer(csections)
	defer C.free(sectionPtr)
	length := int(clength)

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

// FIXME
func (section *UciSection) LoadOption(name string) *UciOption {
	coption := section.parent.parent.uci_lookup_option(section.ptr, name)
	if coption == nil {
		return nil
	}

	option := &UciOption{
		Name:   name,
		ptr:    coption,
		parent: section,
	}

	if coption._type == C.UCI_TYPE_STRING {
		option.Type = UCI_TYPE_STRING
		var cvalue *C.char
		C.option_str_value(option.ptr, cvalue)
		option.Value = C.GoString(cvalue)

	} else if coption._type == C.UCI_TYPE_LIST {
		option.Type = UCI_TYPE_LIST
		var cvalues *C.char
		var clength C.int

		C.option_list_value(option.ptr, &cvalues, &clength)

		valuePtr := unsafe.Pointer(cvalues)
		defer C.free(valuePtr)
		length := int(clength)

		valueArray := (*[1 << 10]*C.char)(valuePtr)
		slice := valueArray[0:length:length]

		option.Values = make([]string, 0)
		for _, v := range slice {
			name := C.GoString(v)
			option.Values = append(option.Values, name)
		}

	} else {
		return nil
	}

	return option
}

func (section *UciSection) SetStringOption(name string, value string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	var uciptr C.struct_uci_ptr

	uciptr.p = section.parent.ptr
	uciptr.s = section.ptr
	// uciptr._package = section.parent.ptr.e.name
	// uciptr.section = section.ptr.e.name
	uciptr.option = cname
	uciptr.value = cvalue

	return section.parent.parent.uci_set(&uciptr)
}

func (section *UciSection) AddListOption(name string, values ...string) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	for _, value := range values {
		cvalue := C.CString(value)
		defer C.free(unsafe.Pointer(cvalue))

		var uciptr C.struct_uci_ptr

		uciptr.p = section.parent.ptr
		uciptr.s = section.ptr
		// uciptr._package = section.parent.ptr.e.name
		// uciptr.section = section.ptr.e.name
		uciptr.option = cname
		uciptr.value = cvalue

		if err = section.parent.parent.uci_add_list(&uciptr); err != nil {
			return err
		}
	}

	return nil
}

func (section *UciSection) DelOption(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var uciptr C.struct_uci_ptr

	uciptr.p = section.parent.ptr
	uciptr.s = section.ptr
	// uciptr._package = section.parent.ptr.e.name
	// uciptr.section = section.ptr.e.name
	uciptr.option = cname

	return section.parent.parent.uci_delete(&uciptr)
}

func (section *UciSection) DelFromList(name string, value string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	var uciptr C.struct_uci_ptr

	uciptr.p = section.parent.ptr
	uciptr.s = section.ptr
	// uciptr._package = section.parent.ptr.e.name
	// uciptr.section = section.ptr.e.name
	uciptr.option = cname
	uciptr.value = cvalue

	return section.parent.parent.uci_del_list(&uciptr)
}

// FIXME
func (section *UciSection) ListOptions() []UciOption {
	var coptions *C.struct_uci_option
	var clength C.int

	C.list_options(section.ptr, &coptions, &clength)

	optionPtr := unsafe.Pointer(coptions)
	defer C.free(optionPtr)
	length := int(clength)

	optionArray := (*[1 << 10]C.struct_uci_option)(optionPtr)
	slice := optionArray[0:length:length]

	options := make([]UciOption, 0)
	for _, v := range slice {
		name := C.GoString(v.e.name)
		options = append(options, *section.LoadOption(name))
	}

	return options
}

// * internal

func (ctx *UciContext) uci_set(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_set(ctx.ptr, ptr)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_delete(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_delete(ctx.ptr, ptr)
	return ctx.uci_ret_to_error(ret, err)
}

func (ctx *UciContext) uci_del_list(ptr *C.struct_uci_ptr) error {
	ret, err := C.uci_del_list(ctx.ptr, ptr)
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
