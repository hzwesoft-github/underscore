package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/hzwesoft-github/underscore/openwrt"
)

var (
	isInvokeServer  = flag.Bool("is", false, "invoke 被调用端")
	isInvokeClient  = flag.Bool("ic", false, "invoke 调用端")
	isEventSender   = flag.Bool("es", false, "event 发送端")
	isEventListener = flag.Bool("el", false, "event 监听端")

	isclient   *openwrt.UbusClient
	icclient   *openwrt.UbusClient
	esclient   *openwrt.UbusClient
	elclient   *openwrt.UbusClient
	ubusObject openwrt.UbusObject

	invoke = make(chan bool)
	event  = make(chan bool)
)

type TestMessage struct {
	F1 string   `json:"f1"`
	F2 int32    `json:"f2"`
	F3 []string `json:"array"`
}

func ubusHandler(obj string, method string, req *openwrt.UbusRequestData, msg string) {
	fmt.Printf("server received: %s\n", msg)
	isclient.SendReply(req, msg)
}

func ubusDataHandler(msg string) {
	fmt.Printf("client received: %s\n", msg)
	fmt.Println()
}

func ubusEventHandler(event string, msg string) {
	fmt.Printf("event %s received: %s\n", event, msg)
}

func init() {
	ubusObject = openwrt.UbusObject{
		Name: "test_obj",
	}

	field1 := openwrt.UbusMethodField{
		Name: "f1",
		Type: openwrt.BLOBMSG_TYPE_STRING,
	}
	field2 := openwrt.UbusMethodField{
		Name: "f2",
		Type: openwrt.BLOBMSG_TYPE_INT32,
	}

	ubusObject.AddMethod("method1", ubusHandler, field1, field2)
}

func invokeServer(client *openwrt.UbusClient) {
	client.AddObject(&ubusObject)
	fmt.Println("obj added")
	invoke <- true
}

func invokeClient(client *openwrt.UbusClient) {
	for i := 0; i < 1000; i++ {
		msg := TestMessage{strconv.Itoa(i), int32(i)}
		fmt.Printf("client invoke %s %s\n", ubusObject.Name, ubusObject.Methods[0].Name)
		client.Invoke(ubusObject.Name, ubusObject.Methods[0].Name, msg, 2000, ubusDataHandler)
		<-time.After(5 * time.Second)
	}
}

func eventListener(client *openwrt.UbusClient) {
	client.RegisterEvent("test_event", ubusEventHandler)
	fmt.Println("event registered")
	event <- true
}

func eventSender(client *openwrt.UbusClient) {
	for i := 0; i < 1000; i++ {
		msg := TestMessage{strconv.Itoa(i), int32(i)}
		fmt.Printf("event %s send\n", "test_event")
		client.SendEvent("test_event", msg)
		<-time.After(5 * time.Second)
	}
}

func main() {
	flag.Parse()

	var err error

	openwrt.UloopInit()

	if *isInvokeServer {
		isclient, err = openwrt.NewUbusClient(true)
		if err != nil {
			panic(err)
		}
		go invokeServer(isclient)
	}

	if *isInvokeClient {
		icclient, err = openwrt.NewUbusClient(true)
		if err != nil {
			panic(err)
		}
		go invokeClient(icclient)
	}

	if *isEventSender {
		esclient, err = openwrt.NewUbusClient(true)
		if err != nil {
			panic(err)
		}
		go eventSender(esclient)
	}

	if *isEventListener {
		elclient, err = openwrt.NewUbusClient(true)
		if err != nil {
			panic(err)
		}
		go eventListener(elclient)
	}

	if *isInvokeServer {
		<-invoke
		isclient.Start()
		fmt.Println("started")
	}

	if *isEventListener {
		<-event
		elclient.Start()
		fmt.Println("started")
	}

	openwrt.UloopRun()
}
