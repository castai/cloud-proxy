package e2etest

import (
	"log"
	"sync"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type Dispatcher struct {
	pendingRequests map[string]chan *proto.HttpResponse
	locker          sync.Mutex

	proxyRequestChan  chan<- *proto.HttpRequest
	proxyResponseChan <-chan *proto.HttpResponse

	logger *log.Logger
}

func NewDispatcher(requestChan chan<- *proto.HttpRequest, responseChan <-chan *proto.HttpResponse, logger *log.Logger) *Dispatcher {
	return &Dispatcher{
		pendingRequests:   make(map[string]chan *proto.HttpResponse),
		locker:            sync.Mutex{},
		proxyRequestChan:  requestChan,
		proxyResponseChan: responseChan,
		logger:            logger,
	}
}

func (d *Dispatcher) Run() {
	go func() {
		d.logger.Println("starting response returning loop")
		for {
			for resp := range d.proxyResponseChan {
				waiter := d.findWaiterForResponse(resp.RequestID)
				waiter <- resp
				d.logger.Println("Sent a response back to caller")
			}
		}
	}()
}

func (d *Dispatcher) SendRequest(req *proto.HttpRequest) (<-chan *proto.HttpResponse, error) {
	waiter := d.addRequestToWaitingList(req.RequestID)
	d.proxyRequestChan <- req
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
		d.logger.Panicln("Trying to send a response for non-existent request", requestID)
	}
	delete(d.pendingRequests, requestID)
	d.locker.Unlock()

	return val
}
