package openwrt

/*
#include <libubox/uloop.h>
*/
import "C"

func UloopInit() error {
	_, err := C.uloop_init()
	return err
}

func UloopRun() error {
	_, err := C.uloop_run()
	return err
}

func UloopDone() error {
	_, err := C.uloop_done()
	return err
}
