package openwrt

/*
#include <libubox/blobmsg_json.h>
#include <libubus.h>

extern void connection_lost_callback(struct ubus_context *ctx);
extern int ubus_handler_stub(struct ubus_context *ctx, struct ubus_object *obj, struct ubus_request_data *req, char *method, struct blob_attr *msg);

static void ubus_connection_lost(struct ubus_context *ctx)
{
	ctx->connection_lost = connection_lost_callback;
}

static int ubus_handler_stub_wrapper(struct ubus_context *ctx, struct ubus_object *obj, struct ubus_request_data *req, const char *method, struct blob_attr *msg){
	return ubus_handler_stub(ctx, obj, req, (char *)method, msg);
}

static struct blobmsg_policy init_blobmsg_policy(char *field_name, enum blobmsg_type field_type)
{
	struct blobmsg_policy policy = {
		.name = field_name,
		.type = field_type,
	};

	return policy;
}

static struct ubus_method init_ubus_method(char *method_name, struct blobmsg_policy *policies, int n_policy)
{
	struct ubus_method method = {
		.name = method_name,
		.handler = ubus_handler_stub_wrapper,
		.policy = policies,
		.n_policy = n_policy,
		.tags = 0,
	};

	return method;
}

extern void ubus_data_handler_stub(struct ubus_request *req, int type, struct blob_attr *msg);

extern void ubus_event_handler_stub(struct ubus_context *ctx, struct ubus_event_handler *ev, char *type, struct blob_attr *msg);

static void ubus_event_handler_wrapper(struct ubus_context *ctx, struct ubus_event_handler *ev, const char *type, struct blob_attr *msg)
{
	ubus_event_handler_stub(ctx, ev, (char *)type, msg);
}

static void bind_ubus_event_handler(struct ubus_event_handler *ue)
{
	ue->cb = ubus_event_handler_wrapper;
}
*/
import "C"
import (
	"fmt"
	"path"
	"sync"
	"time"
	"unsafe"

	"github.com/hzwesoft-github/underscore/lang"
)

const DEFAULT_SOCK = "/var/run/ubus/ubus.sock"

type BlobmsgType int

const (
	BLOBMSG_TYPE_UNSPEC BlobmsgType = iota
	BLOBMSG_TYPE_ARRAY
	BLOBMSG_TYPE_TABLE
	BLOBMSG_TYPE_STRING
	BLOBMSG_TYPE_INT64
	BLOBMSG_TYPE_INT32
	BLOBMSG_TYPE_INT16
	BLOBMSG_TYPE_INT8
	BLOBMSG_TYPE_BOOL
	BLOBMSG_TYPE_DOUBLE
)

func (typ BlobmsgType) toEnum() C.enum_blobmsg_type {
	switch typ {
	case BLOBMSG_TYPE_UNSPEC:
		return C.BLOBMSG_TYPE_UNSPEC
	case BLOBMSG_TYPE_ARRAY:
		return C.BLOBMSG_TYPE_ARRAY
	case BLOBMSG_TYPE_TABLE:
		return C.BLOBMSG_TYPE_TABLE
	case BLOBMSG_TYPE_STRING:
		return C.BLOBMSG_TYPE_STRING
	case BLOBMSG_TYPE_INT64:
		return C.BLOBMSG_TYPE_INT64
	case BLOBMSG_TYPE_INT32:
		return C.BLOBMSG_TYPE_INT32
	case BLOBMSG_TYPE_INT16:
		return C.BLOBMSG_TYPE_INT16
	case BLOBMSG_TYPE_INT8:
		return C.BLOBMSG_TYPE_INT8
	case BLOBMSG_TYPE_BOOL:
		return C.BLOBMSG_TYPE_BOOL
	case BLOBMSG_TYPE_DOUBLE:
		return C.BLOBMSG_TYPE_DOUBLE
	default:
		return C.BLOBMSG_TYPE_UNSPEC
	}
}

var (
	ubusHandlerMap        lang.DMap[string, string, UbusHandler] = lang.NewDMap[string, string, UbusHandler]()
	ubusDataHandlerMap    map[int32]UbusDataHandler              = make(map[int32]UbusDataHandler)
	ubusDataHandlerErrors map[int32]error                        = make(map[int32]error)
	ubusEventHandlerMap   map[string]UbusEventHandler            = make(map[string]UbusEventHandler)
	mutex                 sync.Mutex
)

// encapsulate ubus_context
type UbusContext struct {
	ptr *C.struct_ubus_context

	ubusObjPtrMap map[string]_UbusObjectPtr
	listeners     map[string]_UbusEventListener
}

// callback
type UbusHandler func(obj string, method string, req *UbusRequestData, msg string)
type UbusDataHandler func(msg string) error
type UbusEventHandler func(event string, msg string)

// ubus object related
type UbusObject struct {
	Name    string
	Methods []UbusMethod
}

func (obj *UbusObject) AddMethod(name string, handler UbusHandler, fields ...UbusMethodField) {
	if obj.Methods == nil {
		obj.Methods = make([]UbusMethod, 0)
	}

	obj.Methods = append(obj.Methods, UbusMethod{name, handler, fields})
}

type UbusMethod struct {
	Name    string
	Handler UbusHandler
	Fields  []UbusMethodField
}

type UbusMethodField struct {
	Name string
	Type BlobmsgType
}

// pointers to be free
type _UbusObjectPtr struct {
	ready        bool
	objName      *C.char
	objPtr       *C.struct_ubus_object
	objTypePtr   *C.struct_ubus_object_type
	methodNames  []*C.char
	methodPtrs   []*C.struct_ubus_method
	fieldNames   []*C.char
	policiesPtrs []*C.struct_blobmsg_policy
}

func (ptr *_UbusObjectPtr) init() {
	ptr.methodNames = make([]*C.char, 0)
	ptr.methodPtrs = make([]*C.struct_ubus_method, 0)
	ptr.fieldNames = make([]*C.char, 0)
	ptr.policiesPtrs = make([]*C.struct_blobmsg_policy, 0)
}

func (ptr *_UbusObjectPtr) free() {
	if ptr.ready {
		return
	}

	if ptr.objName != nil {
		C.free(unsafe.Pointer(ptr.objName))
	}
	if ptr.objPtr != nil {
		C.free(unsafe.Pointer(ptr.objPtr))
	}
	if ptr.objTypePtr != nil {
		C.free(unsafe.Pointer(ptr.objTypePtr))
	}
	if len(ptr.methodNames) > 0 {
		for i := 0; i < len(ptr.methodNames); i++ {
			C.free(unsafe.Pointer(ptr.methodNames[i]))
		}
	}
	if len(ptr.methodPtrs) > 0 {
		for i := 0; i < len(ptr.methodPtrs); i++ {
			C.free(unsafe.Pointer(ptr.methodPtrs[i]))
		}
	}
	if len(ptr.fieldNames) > 0 {
		for i := 0; i < len(ptr.fieldNames); i++ {
			C.free(unsafe.Pointer(ptr.fieldNames[i]))
		}
	}
	if len(ptr.policiesPtrs) > 0 {
		for i := 0; i < len(ptr.policiesPtrs); i++ {
			C.free(unsafe.Pointer(ptr.policiesPtrs[i]))
		}
	}
}

// ubus event related
type _UbusEventListener struct {
	ptr *C.struct_ubus_event_handler
}

func (l *_UbusEventListener) free() {
	C.free(unsafe.Pointer(l.ptr))
}

//export connection_lost_callback
func connection_lost_callback(ctx *C.struct_ubus_context) {
	cpath := C.CString(DEFAULT_SOCK)
	defer C.free(unsafe.Pointer(cpath))

	for {
		if int32(C.ubus_reconnect(ctx, cpath)) == 0 {
			break
		}

		<-time.After(time.Second)
	}

	C.ubus_add_uloop(ctx)
}

// new connection to ubusd
func NewUbusContext(reconnect bool) (context *UbusContext, err error) {
	cpath := C.CString(DEFAULT_SOCK)
	defer C.free(unsafe.Pointer(cpath))

	var ctx *C.struct_ubus_context
	if ctx, err = C.ubus_connect(cpath); err != nil {
		return nil, err
	}

	if reconnect {
		C.ubus_connection_lost(ctx)
	}

	return &UbusContext{
		ptr:           ctx,
		ubusObjPtrMap: make(map[string]_UbusObjectPtr),
		listeners:     make(map[string]_UbusEventListener),
	}, nil
}

func (ctx *UbusContext) AddULoop() error {
	_, err := C.ubus_add_uloop(ctx.ptr)
	return err
}

func (ctx *UbusContext) Free() error {
	if _, err := C.ubus_free(ctx.ptr); err != nil {
		return err
	}

	for _, ptr := range ctx.ubusObjPtrMap {
		ptr.free()
	}

	for _, l := range ctx.listeners {
		l.free()
	}

	return nil
}

//export ubus_handler_stub
func ubus_handler_stub(ctx *C.struct_ubus_context, obj *C.struct_ubus_object,
	req *C.struct_ubus_request_data, method *C.char, msg *C.struct_blob_attr) C.int {
	objName := C.GoString(obj.name)
	methodName := C.GoString(method)

	if !lang.HasDMapKey(ubusHandlerMap, objName, methodName) {
		return C.int(3)
	}

	str := C.blobmsg_format_json_indent(msg, C.bool(true), C.int(0))
	defer C.free(unsafe.Pointer(str))

	r := &UbusRequestData{
		ptr: req,
	}
	ubusHandlerMap[objName][methodName](objName, methodName, r, C.GoString(str))

	return C.int(0)
}

// encapsulate ubus_request_data
type UbusRequestData struct {
	ptr *C.struct_ubus_request_data
}

// encapsulate ubus_add_object
func (ctx *UbusContext) AddObject(obj *UbusObject) error {
	freePtr := _UbusObjectPtr{}
	freePtr.init()
	defer func() {
		if r := recover(); r != nil {
			freePtr.free()
			return
		}

		freePtr.free()
	}()

	cObjName := C.CString(obj.Name)
	// ^defer free
	freePtr.objName = cObjName

	cObjTypePtr := (*C.struct_ubus_object_type)(C.calloc(1, C.sizeof_struct_ubus_object_type))
	// ^defer free
	freePtr.objTypePtr = cObjTypePtr
	cObjTypePtr.name = cObjName
	cObjTypePtr.id = 0

	cObjPtr := (*C.struct_ubus_object)(C.calloc(1, C.sizeof_struct_ubus_object))
	// ^defer free
	freePtr.objPtr = cObjPtr
	cObjPtr.name = cObjName
	cObjPtr._type = cObjTypePtr

	if len(obj.Methods) == 0 {
		cObjTypePtr.n_methods = 0
		cObjPtr.n_methods = 0
	} else {
		cMethods := make([]C.struct_ubus_method, 0)
		for _, method := range obj.Methods {
			cMethodName := C.CString(method.Name)
			// ^defer free
			freePtr.methodNames = append(freePtr.methodNames, cMethodName)

			cPolicies := make([]C.struct_blobmsg_policy, 0)
			for _, field := range method.Fields {
				cFieldName := C.CString(field.Name)
				// ^defer free
				freePtr.fieldNames = append(freePtr.fieldNames, cFieldName)

				cPolicy := C.init_blobmsg_policy(cFieldName, field.Type.toEnum())
				cPolicies = append(cPolicies, cPolicy)
			}

			var cMethod C.struct_ubus_method
			if len(cPolicies) == 0 {
				cMethod = C.init_ubus_method(cMethodName, nil, 0)
			} else {
				cPoliciesPtr := (*C.struct_blobmsg_policy)(C.calloc(C.ulong(len(method.Fields)), C.sizeof_struct_blobmsg_policy))
				// ^defer free
				freePtr.policiesPtrs = append(freePtr.policiesPtrs, cPoliciesPtr)
				C.memcpy(unsafe.Pointer(cPoliciesPtr), unsafe.Pointer(&cPolicies[0]), C.sizeof_struct_blobmsg_policy*C.ulong(len(method.Fields)))
				cMethod = C.init_ubus_method(cMethodName, cPoliciesPtr, C.int(len(cPolicies)))
			}

			cMethods = append(cMethods, cMethod)

			lang.AddDMapValue(ubusHandlerMap, obj.Name, method.Name, method.Handler)
		}

		cMethodPtr := (*C.struct_ubus_method)(C.calloc(C.ulong(len(cMethods)), C.sizeof_struct_ubus_method))
		// ^defer free
		freePtr.methodPtrs = append(freePtr.methodPtrs, cMethodPtr)
		C.memcpy(unsafe.Pointer(cMethodPtr), unsafe.Pointer(&cMethods[0]), C.sizeof_struct_ubus_method*C.ulong(len(cMethods)))

		cObjTypePtr.methods = cMethodPtr
		cObjTypePtr.n_methods = C.int(len(cMethods))

		cObjPtr.methods = cMethodPtr
		cObjPtr.n_methods = C.int(len(cMethods))
	}

	ret, err := C.ubus_add_object(ctx.ptr, cObjPtr)
	if err != nil {
		return err
	}
	if ret != C.UBUS_STATUS_OK {
		return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	freePtr.ready = true
	ctx.ubusObjPtrMap[obj.Name] = freePtr

	return nil
}

// encapsulate ubus_remove_object
func (ctx *UbusContext) RemoveObject(name string) error {
	if ptr, ok := ctx.ubusObjPtrMap[name]; ok {
		ret, err := C.ubus_remove_object(ctx.ptr, ptr.objPtr)
		if err != nil {
			return err
		}
		if ret != C.UBUS_STATUS_OK {
			return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
		}

		ptr.free()

		delete(ctx.ubusObjPtrMap, name)
	}

	return nil
}

// encapsulate ubus_send_reply
func (ctx *UbusContext) SendReply(req *UbusRequestData, msg any) error {
	buf := NewBlobBuf()
	defer buf.Free()

	buf.Init(0)
	if err := buf.AddJsonFrom(msg); err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()
	ret, err := C.ubus_send_reply(ctx.ptr, req.ptr, buf.ptr.head)
	if err != nil {
		return err
	}
	if ret != C.UBUS_STATUS_OK {
		return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	return nil
}

// encapsulate ubus_lookup_id
func (ctx *UbusContext) LookupId(path string) (uint32, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var id C.uint32_t
	ret, err := C.ubus_lookup_id(ctx.ptr, cpath, &id)

	if err != nil {
		return 0, err
	}
	if ret != C.UBUS_STATUS_OK {
		return 0, fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	return uint32(id), nil
}

//export ubus_data_handler_stub
func ubus_data_handler_stub(req *C.struct_ubus_request, typ C.int, msg *C.struct_blob_attr) {
	seq := int32(req.seq)

	if handler, ok := ubusDataHandlerMap[seq]; ok {
		defer delete(ubusDataHandlerMap, seq)

		str := C.blobmsg_format_json_indent(msg, C.bool(true), C.int(0))
		defer C.free(unsafe.Pointer(str))

		if err := handler(C.GoString(str)); err != nil {
			ubusDataHandlerErrors[seq] = err
		}
	}
}

// encapsulate ubus invoke
func (ctx *UbusContext) Invoke(id uint32, method string, param any, timeout int, cb UbusDataHandler) error {
	cmethod := C.CString(method)
	defer C.free(unsafe.Pointer(cmethod))

	buf := NewBlobBuf()
	defer buf.Free()

	buf.Init(0)
	buf.AddJsonFrom(param)

	req := (*C.struct_ubus_request)(C.calloc(1, C.sizeof_struct_ubus_request))
	defer C.free(unsafe.Pointer(req))

	var err error
	var ret C.int

	mutex.Lock()
	defer mutex.Unlock()
	if ret, err = C.ubus_invoke_async(ctx.ptr, C.uint32_t(id), cmethod, buf.ptr.head, req); err != nil {
		return err
	}
	if ret != C.UBUS_STATUS_OK {
		return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	req.data_cb = C.ubus_data_handler_t(C.ubus_data_handler_stub)
	seq := int32(req.seq)
	ubusDataHandlerMap[seq] = cb

	retry := 5
	for i := 0; i < retry; i++ {
		if ret, err = C.ubus_complete_request(ctx.ptr, req, C.int(timeout)); err != nil {
			// if err.Error() == "resource temporarily unavailable"
			continue
		}
		break
	}

	if err != nil {
		return err
	}
	if ret != C.UBUS_STATUS_OK {
		return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	if err, ok := ubusDataHandlerErrors[seq]; ok {
		defer delete(ubusDataHandlerErrors, seq)
		return err
	}

	return nil
}

//export ubus_event_handler_stub
func ubus_event_handler_stub(ctx *C.struct_ubus_context, ev *C.struct_ubus_event_handler, typ *C.char, msg *C.struct_blob_attr) {
	id := C.GoString(typ)
	str := C.blobmsg_format_json_indent(msg, C.bool(true), C.int(0))
	defer C.free(unsafe.Pointer(str))

	for pattern, handler := range ubusEventHandlerMap {
		matched, err := path.Match(pattern, id)
		if err != nil {
			continue
		}
		if matched {
			handler(id, C.GoString(str))
		}
	}
}

// encapsulate ubus_register_event_handler
func (ctx *UbusContext) RegisterEvent(pattern string, cb UbusEventHandler) error {
	listener := _UbusEventListener{
		ptr: (*C.struct_ubus_event_handler)(C.calloc(1, C.sizeof_struct_ubus_event_handler)),
	}

	cpattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cpattern))

	C.bind_ubus_event_handler(listener.ptr)

	ret, err := C.ubus_register_event_handler(ctx.ptr, listener.ptr, cpattern)
	if err != nil {
		return err
	}
	if ret != C.UBUS_STATUS_OK {
		return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	ubusEventHandlerMap[pattern] = cb
	ctx.listeners[pattern] = listener

	return nil
}

// encapsulate ubus_unregister_event_handler
func (ctx *UbusContext) UnregisterEvent(pattern string) error {
	if listener, ok := ctx.listeners[pattern]; ok {
		ret, err := C.ubus_unregister_event_handler(ctx.ptr, listener.ptr)
		if err != nil {
			return err
		}
		if ret != C.UBUS_STATUS_OK {
			return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
		}

		listener.free()
	}

	return nil
}

// encapsulate ubus_send_event
func (ctx *UbusContext) SendEvent(id string, msg any) error {
	buf := NewBlobBuf()
	defer buf.Free()

	buf.Init(0)
	if err := buf.AddJsonFrom(msg); err != nil {
		return err
	}

	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))

	mutex.Lock()
	defer mutex.Unlock()
	ret, err := C.ubus_send_event(ctx.ptr, cid, buf.ptr.head)
	if err != nil {
		return err
	}
	if ret != C.UBUS_STATUS_OK {
		return fmt.Errorf("%d: %s", int(ret), UbusErrorString(ret))
	}

	return nil
}

// encapsulate ubus_strerror
func UbusErrorString(code C.int) string {
	return C.GoString(C.ubus_strerror(code))
}
