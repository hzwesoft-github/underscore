package openwrt

type UbusClient struct {
	Context *UbusContext
	Started bool

	objs      map[string]UbusObject
	listeners map[string]UbusEventHandler
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
	if client.objs == nil {
		client.objs = make(map[string]UbusObject)
	}
	client.objs[obj.Name] = *obj
}

func (client *UbusClient) RegisterEvent(event string, cb UbusEventHandler) {
	if client.listeners == nil {
		client.listeners = make(map[string]UbusEventHandler)
	}
	client.listeners[event] = cb
}

func (client *UbusClient) RemoveObject(name string) error {
	delete(client.objs, name)

	if client.Started {
		return client.Context.RemoveObject(name)
	}

	return nil
}

func (client *UbusClient) UnregisterEvent(event string) error {
	delete(client.listeners, event)

	if client.Started {
		return client.Context.UnregisterEvent(event)
	}

	return nil
}

func (client *UbusClient) Start() (err error) {
	if len(client.objs) > 0 {
		for name := range client.objs {
			obj := client.objs[name]
			if err = client.Context.AddObject(&obj); err != nil {
				return err
			}
		}
	}

	if len(client.listeners) > 0 {
		for event := range client.listeners {
			cb := client.listeners[event]
			if err = client.Context.RegisterEvent(event, cb); err != nil {
				return err
			}
		}
	}

	if len(client.objs) > 0 || len(client.listeners) > 0 {
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
