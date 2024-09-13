package e2etest

import (
	"log"
	"sync"

	cloudproxyv1alpha "github.com/castai/cloud-proxy/proto/gen/proto/v1alpha"
)

type Dispatcher struct {
	pendingRequests map[string]chan *cloudproxyv1alpha.StreamCloudProxyRequest
	locker          sync.Mutex

	proxyRequestChan  chan<- *cloudproxyv1alpha.StreamCloudProxyResponse
	proxyResponseChan <-chan *cloudproxyv1alpha.StreamCloudProxyRequest

	logger *log.Logger
}

func NewDispatcher(requestChan chan<- *cloudproxyv1alpha.StreamCloudProxyResponse, responseChan <-chan *cloudproxyv1alpha.StreamCloudProxyRequest, logger *log.Logger) *Dispatcher {
	return &Dispatcher{
		pendingRequests:   make(map[string]chan *cloudproxyv1alpha.StreamCloudProxyRequest),
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
				waiter := d.findWaiterForResponse(resp.GetResponse().GetMessageId())
				waiter <- resp
				d.logger.Println("Sent a response back to caller")
			}
		}
	}()
}

func (d *Dispatcher) SendRequest(req *cloudproxyv1alpha.StreamCloudProxyResponse) (<-chan *cloudproxyv1alpha.StreamCloudProxyRequest, error) {
	waiter := d.addRequestToWaitingList(req.MessageId)
	d.proxyRequestChan <- req
	return waiter, nil
}

func (d *Dispatcher) addRequestToWaitingList(requestID string) <-chan *cloudproxyv1alpha.StreamCloudProxyRequest {
	waiter := make(chan *cloudproxyv1alpha.StreamCloudProxyRequest, 1)
	d.locker.Lock()
	d.pendingRequests[requestID] = waiter
	d.locker.Unlock()
	return waiter
}

func (d *Dispatcher) findWaiterForResponse(requestID string) chan *cloudproxyv1alpha.StreamCloudProxyRequest {
	d.locker.Lock()
	val, ok := d.pendingRequests[requestID]
	if !ok {
		d.logger.Panicln("Trying to send a response for non-existent request", requestID)
	}
	delete(d.pendingRequests, requestID)
	d.locker.Unlock()

	return val
}
