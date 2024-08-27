package dummy

import (
	"log"
	"sync"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type Dispatcher struct {
	pendingRequests map[string]chan *proto.HttpResponse
	locker          sync.Mutex

	ProxyRequestChan  chan<- *proto.HttpRequest
	ProxyResponseChan <-chan *proto.HttpResponse
}

func NewDispatcher(requestChan chan<- *proto.HttpRequest, responseChan <-chan *proto.HttpResponse) *Dispatcher {
	return &Dispatcher{
		pendingRequests:   make(map[string]chan *proto.HttpResponse),
		locker:            sync.Mutex{},
		ProxyRequestChan:  requestChan,
		ProxyResponseChan: responseChan,
	}
}

func (d *Dispatcher) Run() {
	go func() {
		log.Println("starting response returning loop")
		for {
			for resp := range d.ProxyResponseChan {
				waiter := d.findWaiterForResponse(resp.RequestID)
				waiter <- resp
				log.Println("Sent a response back to caller")
			}
		}
	}()
}

func (d *Dispatcher) SendRequest(req *proto.HttpRequest) (<-chan *proto.HttpResponse, error) {
	waiter := d.addRequestToWaitingList(req.RequestID)
	d.ProxyRequestChan <- req
	return waiter, nil
}

func (d *Dispatcher) addRequestToWaitingList(requestID string) <-chan *proto.HttpResponse {
	waiter := make(chan *proto.HttpResponse, 1)
	d.locker.Lock()
	d.pendingRequests[requestID] = waiter
	d.locker.Unlock()
	return waiter
}

func (d *Dispatcher) findWaiterForResponse(requestID string) chan *proto.HttpResponse {
	d.locker.Lock()
	val, ok := d.pendingRequests[requestID]
	if !ok {
		log.Panicln("Trying to send a response for non-existent request", requestID)
	}
	delete(d.pendingRequests, requestID)
	d.locker.Unlock()

	return val
}
