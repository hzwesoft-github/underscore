package openwrt

type UbusClient struct {
	Context *UbusContext
	Started bool

	Objects   map[string]UbusObject
	Listeners map[string]UbusEventHandler
}

func NewUbusClient(reconnect bool) (*UbusClient, error) {
	ctx, err := NewUbusContext(reconnect)
	if err != nil {
		return nil, err
	}

	return &UbusClient{ctx, false, nil, nil}, nil
}

func (client *UbusClient) Free() {
	client.Context.Free()
}

func (client *UbusClient) AddObject(obj *UbusObject) {
	if client.Objects == nil {
		client.Objects = make(map[string]UbusObject)
	}
	client.Objects[obj.Name] = *obj
}

func (client *UbusClient) RegisterEvent(event string, cb UbusEventHandler) {
	if client.Listeners == nil {
		client.Listeners = make(map[string]UbusEventHandler)
	}
	client.Listeners[event] = cb
}

func (client *UbusClient) RemoveObject(name string) error {
	delete(client.Objects, name)

	if client.Started {
		return client.Context.RemoveObject(name)
	}

	return nil
}

func (client *UbusClient) UnregisterEvent(event string) error {
	delete(client.Listeners, event)

	if client.Started {
		return client.Context.UnregisterEvent(event)
	}

	return nil
}

func (client *UbusClient) Start() (err error) {
	if len(client.Objects) > 0 {
		for name := range client.Objects {
			obj := client.Objects[name]
			if err = client.Context.AddObject(&obj); err != nil {
				return err
			}
		}
	}

	if len(client.Listeners) > 0 {
		for event := range client.Listeners {
			cb := client.Listeners[event]
			if err = client.Context.RegisterEvent(event, cb); err != nil {
				return err
			}
		}
	}

	if len(client.Objects) > 0 || len(client.Listeners) > 0 {
		client.Context.AddULoop()
	}

	return nil
}

func (client *UbusClient) SendReply(req *UbusRequestData, msg any) error {
	return client.Context.SendReply(req, msg)
}

func (client *UbusClient) Invoke(obj string, method string, param any, timeout int, cb UbusDataHandler) error {
	id, err := client.Context.LookupId(obj)
	if err != nil {
		return err
	}

	return client.Context.Invoke(id, method, param, timeout, cb)
}

func (client *UbusClient) SendEvent(id string, msg any) error {
	return client.Context.SendEvent(id, msg)
}
