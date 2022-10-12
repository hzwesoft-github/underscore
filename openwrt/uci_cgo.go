package openwrt

/*
#include <uci.h>
*/
import "C"

type UciContext struct {
	ptr *C.struct_uci_context
}

func NewUciContext() *UciContext {
	return &UciContext{
		ptr: C.uci_alloc_context(),
	}
}

func (ctx *UciContext) Free() {
	C.uci_free_context(ctx.ptr)
}

func (ctx *UciContext) ErrorString(prefix string) {
	// TODO
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

type UciPtr struct {
	ptr C.struct_uci_ptr
}
