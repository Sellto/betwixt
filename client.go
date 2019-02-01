package betwixt

import (
	"errors"
	"log"
	"strings"
	"github.com/lgln/canopus"
)

// NewLwm2mClient instantiates a new instance of LWM2M Client
func NewLwm2mClient(name, local, remote string, registry Registry) LWM2MClient {
	server := canopus.NewServer()
	conn, _ := canopus.Dial(remote)


	// Create Mandatory
	c := &DefaultLWM2MClient{
		coapServer:     server,
		coapConn:       conn,
		enabledObjects: make(map[LWM2MObjectType]Object),
		registry:       registry,
	}

	mandatory := registry.GetMandatory()
	for _, o := range mandatory {
		c.EnableObject(o.GetType(), NewNullEnabler())
	}

	return c
}

// DefaultLWM2MClient ;
type DefaultLWM2MClient struct {
	coapServer     canopus.CoapServer
	coapConn       canopus.Connection
	registry       Registry
	enabledObjects map[LWM2MObjectType]Object
	path           string

	// Events
	evtOnStartup FnOnStartup
	evtOnRead    FnOnRead
	evtOnWrite   FnOnWrite
	evtOnExecute FnOnExecute
	evtOnError   FnOnError
	evtOnObserve FnOnObserve
}

// Register this client to a LWM2M Server instance
// name must be unique and be less than 10 characers
func (c *DefaultLWM2MClient) Register(name string) (string, error) {
	if len(name) > 50 {
		return "", errors.New("Client name can not exceed 10 characters")
	}

	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Post)
	req.SetStringPayload(BuildModelResourceStringPayload(c.enabledObjects))
	req.SetRequestURI("/rd")
	req.SetURIQuery("ep", name)

	resp, err := c.coapConn.Send(req)
	path := ""
	if err != nil {
		return "", err
	} else {
		path = resp.GetMessage().GetLocationPath()
	}
	c.path = path

	c.coapConn.Close()
	return path, nil
}

// SetEnabler ; Sets/Defines an Enabler for a given LWM2M Object Type
func (c *DefaultLWM2MClient) SetEnabler(t LWM2MObjectType, e ObjectEnabler) {
	_, ok := c.enabledObjects[t]
	if ok {
		c.enabledObjects[t].SetEnabler(e)
	}
}

// GetEnabledObjects ; Returns a list of LWM2M Enabled Objects
func (c *DefaultLWM2MClient) GetEnabledObjects() map[LWM2MObjectType]Object {
	return c.enabledObjects
}

// GetRegistry ; Returns the registry used for looking up LWM2M object type definitions
func (c *DefaultLWM2MClient) GetRegistry() Registry {
	return c.registry
}

// Deregister ; Unregisters this client from a LWM2M server which was previously registered
func (c *DefaultLWM2MClient) Deregister() {
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Delete)

	req.SetRequestURI(c.path)
	_, err := c.coapConn.Send(req)

	if err != nil {
		log.Println(err)
	}
}

// Update ;
func (c *DefaultLWM2MClient) Update() {

}

func (c *DefaultLWM2MClient) Notify(payload []byte) {
	s := c.coapServer
	s.Notify("/3200/0", payload, true)
}

// AddResource ;
func (c *DefaultLWM2MClient) AddResource() {

}

// AddObject ;
func (c *DefaultLWM2MClient) AddObject() {

}

// UseRegistry ;
func (c *DefaultLWM2MClient) UseRegistry(reg Registry) {
	c.registry = reg
}

// EnableObject ; Registes an object enabler for a given LWM2M object type
func (c *DefaultLWM2MClient) EnableObject(t LWM2MObjectType, e ObjectEnabler) error {
	_, ok := c.enabledObjects[t]
	if !ok {
		if c.registry == nil {
			return errors.New("No registry found/set")
		}
		c.enabledObjects[t] = NewObject(t, e, c.registry)

		return nil
	} else {
		return errors.New("Object already enabled")
	}
}

// AddObjectInstance ; Adds a new object instance for a previously enabled LWM2M object type
func (c *DefaultLWM2MClient) AddObjectInstance(t LWM2MObjectType, instance int) error {
	o := c.enabledObjects[t]
	if o != nil {
		o.AddInstance(instance)

		return nil
	}
	return errors.New("Attempting to add a nil instance")
}

// AddObjectInstances ; Adds a list of object instance for a previously enabled LWM2M object type
func (c *DefaultLWM2MClient) AddObjectInstances(t LWM2MObjectType, instances ...int) {
	for _, o := range instances {
		c.AddObjectInstance(t, o)
	}
}

// GetObject ;
func (c *DefaultLWM2MClient) GetObject(n LWM2MObjectType) Object {
	return c.enabledObjects[n]
}

func (c *DefaultLWM2MClient) validate() {

}

// Start up the LWM2M client, listens to incoming requests and fires the OnStart event
func (c *DefaultLWM2MClient) Start() {
	c.validate()

	s := c.coapServer

	s.OnStart(func(server canopus.CoapServer) {
		if c.evtOnStartup != nil {
			c.evtOnStartup()
		}
	})

	s.OnObserve(func(resource string, msg canopus.Message) {
		log.Printf("Observe Requested on %s\n", resource)
		if c.evtOnObserve != nil {
			c.evtOnObserve(resource)
		}
	})


	s.Get("/:obj/:inst/:rsrc", c.handleReadRequest)
	s.Get("/:obj/:inst", c.handleReadRequest)
	s.Get("/:obj", c.handleReadRequest)

	s.Put("/:obj/:inst/:rsrc", c.handleWriteRequest)
	s.Put("/:obj/:inst", c.handleWriteRequest)

	s.Delete("/:obj/:inst", c.handleDeleteRequest)

	s.Post("/:obj/:inst/:rsrc", c.handleExecuteRequest)
	s.Post("/:obj/:inst", c.handleCreateRequest)

	// sooskim c.coapServer.Start()
	con := c.coapConn.(*canopus.UDPConnection).GetConn()
	a := strings.Split(con.LocalAddr().String(), ":")[1]

	c.coapServer.ListenAndServe(a)
}

// Handles LWM2M Create Requests (not to be mistaken for/not the same as  CoAP PUT)
func (c *DefaultLWM2MClient) handleCreateRequest(req canopus.Request) canopus.Response {
	log.Println("Create Request")
	attrResource := req.GetAttribute("rsrc")
	objectID := req.GetAttributeAsInt("obj")
	instanceID := req.GetAttributeAsInt("inst")

	var resourceID = -1

	if attrResource != "" {
		resourceID = req.GetAttributeAsInt("rsrc")
	}

	t := LWM2MObjectType(objectID)
	obj := c.GetObject(t)
	enabler := obj.GetEnabler()

	msg := canopus.NewMessageOfType(canopus.MessageAcknowledgment, req.GetMessage().GetMessageId(), canopus.NewEmptyPayload()).(*canopus.CoapMessage)
	msg.Token = req.GetMessage().GetToken()
	msg.Payload = canopus.NewEmptyPayload()

	if enabler != nil {
		lwReq := Default(req, OPERATIONTYPE_CREATE)
		response := enabler.OnCreate(instanceID, resourceID, lwReq)
		msg.Code = response.GetResponseCode()
	} else {
		msg.Code = canopus.CoapCodeMethodNotAllowed
	}
	return canopus.NewResponseWithMessage(msg)
}

// Handles LWM2M Read Requests (not to be mistaken for/not the same as  CoAP GET)
func (c *DefaultLWM2MClient) handleReadRequest(req canopus.Request) canopus.Response {
	attrResource := req.GetAttribute("rsrc")
	objectID := req.GetAttributeAsInt("obj")
	instanceID := req.GetAttributeAsInt("inst")

	var resourceID = -1

	if attrResource != "" {
		resourceID = req.GetAttributeAsInt("rsrc")
	}
	
	t := LWM2MObjectType(objectID)
	obj := c.GetObject(t)
	enabler := obj.GetEnabler()

	msg := canopus.NewMessageOfType(canopus.MessageAcknowledgment, req.GetMessage().GetMessageId(), canopus.NewEmptyPayload()).(*canopus.CoapMessage)
	msg.Token = req.GetMessage().GetToken()

	if enabler != nil {
		model := obj.GetDefinition()
		resource := model.GetResource(LWM2MResourceType(resourceID))

		if resource == nil {
			// TODO: Return TLV of Object Instance
			msg.Code = canopus.CoapCodeNotFound
		} else {
			if !IsReadableResource(resource) {
				msg.Code = canopus.CoapCodeMethodNotAllowed
			} else {
				lwReq := Default(req, OPERATIONTYPE_READ)
				response := enabler.OnRead(instanceID, resourceID, lwReq)

				val := response.GetResponseValue()
				msg.Code = response.GetResponseCode()

				msg.AddOption(canopus.OptionContentFormat, MediaTypeFromValue(val))
				b := EncodeValue(resource.GetId(), resource.MultipleValuesAllowed(), val)
				msg.Payload = canopus.NewBytesPayload(b)
			}
		}
	} else {
		msg.Code = canopus.CoapCodeMethodNotAllowed
	}
	return canopus.NewResponseWithMessage(msg)
}

// Handles LWM2M Delete Requests (not to be mistaken for/not the same as  CoAP DELETE)
func (c *DefaultLWM2MClient) handleDeleteRequest(req canopus.Request) canopus.Response {
	log.Println("Delete Request")
	objectID := req.GetAttributeAsInt("obj")
	instanceID := req.GetAttributeAsInt("inst")

	t := LWM2MObjectType(objectID)
	enabler := c.GetObject(t).GetEnabler()

	msg := canopus.NewMessageOfType(canopus.MessageAcknowledgment, req.GetMessage().GetMessageId(), canopus.NewEmptyPayload()).(*canopus.CoapMessage)
	msg.Token = req.GetMessage().GetToken()

	if enabler != nil {
		lwReq := Default(req, OPERATIONTYPE_DELETE)

		response := enabler.OnDelete(instanceID, lwReq)
		msg.Code = response.GetResponseCode()
	} else {
		msg.Code = canopus.CoapCodeMethodNotAllowed
	}
	return canopus.NewResponseWithMessage(msg)
}

func (c *DefaultLWM2MClient) handleDiscoverRequest() {
	log.Println("Discovery Request")
}

func (c *DefaultLWM2MClient) handleObserveRequest() {
	log.Println("Observe Request")
}

// Handles LWM2M Write Requests (not to be mistaken for/not the same as  CoAP POST)
func (c *DefaultLWM2MClient) handleWriteRequest(req canopus.Request) canopus.Response {
	log.Println("Write Request")
	attrResource := req.GetAttribute("rsrc")
	objectID := req.GetAttributeAsInt("obj")
	instanceID := req.GetAttributeAsInt("inst")

	var resourceID = -1

	if attrResource != "" {
		resourceID = req.GetAttributeAsInt("rsrc")
	}

	t := LWM2MObjectType(objectID)
	obj := c.GetObject(t)
	enabler := obj.GetEnabler()

	msg := canopus.NewMessageOfType(canopus.MessageAcknowledgment, req.GetMessage().GetMessageId(), canopus.NewEmptyPayload()).(*canopus.CoapMessage)
	msg.Token = req.GetMessage().GetToken()

	if enabler != nil {
		model := obj.GetDefinition()
		resource := model.GetResource(LWM2MResourceType(resourceID))
		if resource == nil {
			// TODO Write to Object Instance
			msg.Code = canopus.CoapCodeNotFound
		} else {
			if !IsWritableResource(resource) {
				msg.Code = canopus.CoapCodeMethodNotAllowed
			} else {
				lwReq := Default(req, OPERATIONTYPE_WRITE)
				response := enabler.OnWrite(instanceID, resourceID, lwReq)
				msg.Code = response.GetResponseCode()
			}
		}
	} else {
		msg.Code = canopus.CoapCodeNotFound
	}
	return canopus.NewResponseWithMessage(msg)
}

// Handles LWM2M Execute Requests
func (c *DefaultLWM2MClient) handleExecuteRequest(req canopus.Request) canopus.Response {
	log.Println("Execute Request")
	attrResource := req.GetAttribute("rsrc")
	objectID := req.GetAttributeAsInt("obj")
	instanceID := req.GetAttributeAsInt("inst")

	var resourceID = -1

	if attrResource != "" {
		resourceID = req.GetAttributeAsInt("rsrc")
	}

	t := LWM2MObjectType(objectID)
	obj := c.GetObject(t)
	enabler := obj.GetEnabler()

	msg := canopus.NewMessageOfType(canopus.MessageAcknowledgment, req.GetMessage().GetMessageId(), canopus.NewEmptyPayload()).(*canopus.CoapMessage)
	msg.Token = req.GetMessage().GetToken()

	if enabler != nil {
		model := obj.GetDefinition()
		resource := model.GetResource(LWM2MResourceType(resourceID))
		if resource == nil {
			msg.Code = canopus.CoapCodeNotFound
		}

		if !IsExecutableResource(resource) {
			msg.Code = canopus.CoapCodeMethodNotAllowed
		} else {
			lwReq := Default(req, OPERATIONTYPE_EXECUTE)
			response := enabler.OnExecute(instanceID, resourceID, lwReq)
			msg.Code = response.GetResponseCode()
		}
	} else {
		msg.Code = canopus.CoapCodeNotFound
	}
	return canopus.NewResponseWithMessage(msg)
}

// Events
func (c *DefaultLWM2MClient) OnStartup(fn FnOnStartup) {
	c.evtOnStartup = fn
}

func (c *DefaultLWM2MClient) OnRead(fn FnOnRead) {
	c.evtOnRead = fn
}

func (c *DefaultLWM2MClient) OnWrite(fn FnOnWrite) {
	c.evtOnWrite = fn
}

func (c *DefaultLWM2MClient) OnExecute(fn FnOnExecute) {
	c.evtOnExecute = fn
}

func (c *DefaultLWM2MClient) OnError(fn FnOnError) {
	c.evtOnError = fn
}

func (c *DefaultLWM2MClient) OnObserve(fn FnOnObserve) {
	c.evtOnObserve = fn
}
