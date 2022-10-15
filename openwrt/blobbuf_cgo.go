package openwrt

/*
#include <libubox/blob.h>
#include <libubox/blobmsg_json.h>
*/
import "C"
import (
	"unsafe"

	"github.com/hzwesoft-github/underscore/json"
)

type BlobBuf struct {
	ptr *C.struct_blob_buf
}

func NewBlobBuf() *BlobBuf {
	p := C.calloc(1, C.sizeof_struct_blob_buf)
	C.memset(p, 0, C.sizeof_struct_blob_buf)
	return &BlobBuf{
		ptr: (*C.struct_blob_buf)(p),
	}
}

func (buf *BlobBuf) Init(id int) int {
	return int(C.blob_buf_init(buf.ptr, C.int(id)))
}

func (buf *BlobBuf) Free() {
	C.blob_buf_free(buf.ptr)
}

func (buf *BlobBuf) AddJsonFrom(obj any) error {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case string:
		buf.AddJsonFromString(v)
	default:
		ret, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		buf.AddJsonFromString(string(ret))
	}

	return nil
}

func (buf *BlobBuf) AddJsonFromString(str string) error {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	_, err := C.blobmsg_add_json_from_string(buf.ptr, cstr)
	return err
}
