// Package conversation contains coordination session internal code that helps implement a typical conversation-like
// session protocol based on a bidirectional gRPC stream.
package conversation

import (
	"context"
	"sync"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Coordination"

	"github.com/ydb-platform/ydb-go-sdk/v3/coordination"
)

// Controller provides a simple mechanism to work with a session protocol using a gRPC bidirectional stream. Creating a
// bidirectional stream client may be quite tricky because messages are usually being processed independently and in
// parallel. Moreover, the gRPC client library puts strict limitations on an implementation of the client, e.g. multiple
// calls of the Send or Recv methods of the stub client must not be performed from different goroutines. Also, there are
// no guarantees that a message successfully dispatched by the Send method will actually reach the server, neither does
// the server enjoy same guarantees when delivering messages to the client. This usually ends up having two goroutines
// (one for sending outgoing messages and another one for receiving incoming ones) and a queue where messages are
// published to be eventually delivered to the server. The Controller simplifies working with this model providing
// generic implementation of the message queue and related routines, handling retries of sent and pending operations
// when the underlying gRPC stream needs to be reconnected.
//
// A typical coordination session looks like this (we are going to skip for now how the gRPC stream is created, handled
// and kept alive, you can find the details on that in the Session, and focus on the protocol):
//
//  1. The client opens a new gRPC bidirectional stream.
//  2. The client sends the SessionStart request and wait until the Failure or the SessionStarted reply.
//  3. The server sends the SessionStarted response with the SessionID. At this point the session is started. If the
//     client needs to reconnect the gRPC stream in the future, it will use that SessionID to attach to the previously
//     created session in the SessionStart request.
//  4. The client sends the AcquireSemaphore request to acquire a permit to the semaphore in this session with count 5.
//  5. After a moment, the client decides to acquire another semaphore, it sends one more AcquireSemaphore request with
//     count 4.
//  6. The server replies with the AcquireSemaphoreResult response to the second AcquireSemaphore request to inform the
//     client that the semaphore was successfully acquired.
//  7. The server replies with the AcquireSemaphorePending response in order to inform the client that the semaphore
//     from the first request has been acquired by another session.
//  8. After a while, the server sends the AcquireSemaphoreResult response which implies that the semaphore from the
//     first request is acquired in the current session.
//  9. Then the client sends the ReleaseSemaphore request in order to release the acquired semaphore.
//  10. The server replies with the ReleaseSemaphoreResult.
//  11. The client terminates the session with the SessionStop request.
//  12. The server let the client know that the session is over sending the SessionStopped response and closing the gRPC
//     stream.
//
// We can notice five independent conversations here:
//
// 1. StartSession, SessionStarted — points 2–3;
// 2. AcquireSemaphore, AcquireSemaphoreResult — points 4, 6;
// 3. AcquireSemaphore, AcquireSemaphorePending, AcquireSemaphoreResult — points 5, 7 and 8;
// 4. ReleaseSemaphore, ReleaseSemaphoreResult — points 9–10;
// 5. SessionStop, SessionStopped — points 11–12.
//
// If at any time the client encounters an unrecoverable error (for example, the underlying gRPC stream becomes
// disconnected), the client will have to replay every conversation from their very beginning. Let us see why it is
// actually the case. But before we go into that, let us look at the grpc.ClientStream SendMsg method:
//
// "…SendMsg does not wait until the message is received by the server. An untimely stream closure may result in lost
// messages. To ensure delivery, users should ensure the RPC completed successfully using RecvMsg…"
//
// This is true for both, the client and the server. So when the server replies to the client it does not really know if
// the response is received by the client. And vice versa, when the client sends a request to the server it has no way
// to know if the request was delivered to the server unless the server sends another message to the client in reply.
//
// That is why conversation-like protocols typically use idempotent requests. Idempotent requests can be safely retried
// as long as you keep the original order of the conversations. For example, if the gRPC stream is terminated before
// the point 6, we cannot know if the server gets the requests. There may be one, two or none AcquireSemaphore requests
// successfully delivered to and handled by the server. Moreover, the server may have already sent to the client the
// corresponding responses. Nevertheless, if the requests are idempotent, we can safely retry them all in the newly
// created gRPC stream and get the same results as we would have got if we had sent them without stream termination.
// Note that if the stream is terminated before the point 8, we still need to replay the first AcquireSemaphore
// conversation because we have no knowledge if the server replied with the AcquireSemaphoreResult in the terminated
// stream.
//
// However, sometimes even idempotent requests cannot be safely retried. Consider the case wherein the point 5 from the
// original list is:
//
//  5. After a moment, the client decides to modify the AcquireSemaphore request and sends another one with the same
//     semaphore but with count 4.
//
// If then the gRPC stream terminates, there are two likely outcomes:
//
//  1. The server received the first request but the second one was not delivered. The current semaphore count is 5.
//  2. The server received and processed the both requests. The current semaphore permit count is 4.
//
// If we retry the both requests, the observed result will be different depending on which outcome occurs:
//
//  1. The first retry will be a noop, the second one will decrease the semaphore count to 4. This is expected behavior.
//  2. The first retry will try to increase the semaphore count to 5, it causes an error. This is unexpected.
//
// In order to avoid that we could postpone a conversation if there is another one for the same semaphore which has been
// sent but has not been yet delivered to the server. For more details, see the WithConflictKey option.
type Controller struct {
	mutex sync.Mutex // guards access to the fields below

	queue     []*Conversation // the message queue, the front is in the end of the slice
	conflicts map[string]struct{}

	notifyChan chan struct{}
	closed     bool
}

// ResponseFilter defines the filter function called by the controller to know if a received message relates to the
// conversation. If a ResponseFilter returns true, the message is considered to be part of the conversation.
type ResponseFilter func(request *Ydb_Coordination.SessionRequest, response *Ydb_Coordination.SessionResponse) bool

// Conversation is a core concept of the conversation package. It is an ordered sequence of connected request/reply
// messages. For example, the acquiring semaphore conversation may look like this:
//
// 1. The client sends the AcquireSemaphore request.
// 2. The server replies with the AcquireSemaphorePending response.
// 3. After a while, the server replies with the AcquireSemaphoreResult response. The conversation is ended.
//
// There may be many different conversations carried out simultaneously in one session, so the exact order of all the
// messages in the session is unspecified. In the example above, there may be other messages (from different
// conversations) between points 1 and 2, or 2 and 3.
type Conversation struct {
	message           func() *Ydb_Coordination.SessionRequest
	responseFilter    ResponseFilter
	acknowledgeFilter ResponseFilter
	cancelMessage     func(req *Ydb_Coordination.SessionRequest) *Ydb_Coordination.SessionRequest
	cancelFilter      ResponseFilter
	conflictKey       string
	requestSent       *Ydb_Coordination.SessionRequest
	cancelRequestSent *Ydb_Coordination.SessionRequest
	response          *Ydb_Coordination.SessionResponse
	responseErr       error
	done              chan struct{}
	idempotent        bool
	canceled          bool
}

// NewController creates a new conversation controller. You usually have one controller per one session.
func NewController() *Controller {
	return &Controller{
		notifyChan: make(chan struct{}, 1),
		conflicts:  make(map[string]struct{}),
	}
}

// WithResponseFilter returns an Option that specifies the filter function that is used to detect the last response
// message in the conversation. If such a message was found, the conversation is immediately ended and the response
// becomes available by the Conversation.Await method.
func WithResponseFilter(filter ResponseFilter) Option {
	return func(c *Conversation) {
		c.responseFilter = filter
	}
}

// WithAcknowledgeFilter returns an Option that specifies the filter function that is used to detect an intermediate
// response message in the conversation. If such a message was found, the conversation continues, but it lets the client
// know that the server successfully consumed the first request of the conversation.
func WithAcknowledgeFilter(filter ResponseFilter) Option {
	return func(c *Conversation) {
		c.acknowledgeFilter = filter
	}
}

// WithCancelMessage returns an Option that specifies the message and filter functions that are used to cancel the
// conversation which has been already sent. This message is sent in the background when the caller cancels the context
// of the Controller.Await function. The response is never received by the caller and is only used to end the
// conversation and remove it from the queue.
func WithCancelMessage(
	message func(req *Ydb_Coordination.SessionRequest) *Ydb_Coordination.SessionRequest,
	filter ResponseFilter,
) Option {
	return func(c *Conversation) {
		c.cancelMessage = message
		c.cancelFilter = filter
	}
}

// WithConflictKey returns an Option that specifies the key that is used to find out messages that cannot be delivered
// to the server until the server acknowledged the request. If there is a conversation with the same conflict key in the
// queue that has not been yet delivered to the server, the controller will temporarily suspend other conversations with
// the same conflict key until the first one is acknowledged.
func WithConflictKey(key string) Option {
	return func(c *Conversation) {
		c.conflictKey = key
	}
}

// WithIdempotence returns an Option that enabled retries for this conversation when the underlying gRPC stream
// reconnects. The controller will replay the whole conversation from scratch unless it is not ended.
func WithIdempotence(idempotent bool) Option {
	return func(c *Conversation) {
		c.idempotent = idempotent
	}
}

// Option configures how we create a new conversation.
type Option func(c *Conversation)

// NewConversation creates a new conversation that starts with a specified message.
func NewConversation(request func() *Ydb_Coordination.SessionRequest, opts ...Option) *Conversation {
	conversation := Conversation{message: request}
	for _, o := range opts {
		if o != nil {
			o(&conversation)
		}
	}

	return &conversation
}

func (c *Controller) notify() {
	select {
	case c.notifyChan <- struct{}{}:
	default:
	}
}

// PushBack puts a new conversation at the end of the queue.
func (c *Controller) PushBack(conversation *Conversation) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return coordination.ErrSessionClosed
	}

	conversation.enqueue()
	c.queue = append([]*Conversation{conversation}, c.queue...)
	c.notify()

	return nil
}

// PushFront puts a new conversation at the beginning of the queue.
func (c *Controller) PushFront(conversation *Conversation) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return coordination.ErrSessionClosed
	}

	conversation.enqueue()
	c.queue = append(c.queue, conversation)
	c.notify()

	return nil
}

func (c *Controller) sendFront() *Ydb_Coordination.SessionRequest {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// We are notified but there are no conversations in the queue. Return nil to make the loop in OnSend wait.
	if len(c.queue) == 0 {
		return nil
	}

	for i := len(c.queue) - 1; i >= 0; i-- {
		req := c.queue[i]

		if req.canceled && req.cancelRequestSent == nil {
			req.sendCancel()
			c.notify()

			return req.cancelRequestSent
		}

		if req.requestSent != nil {
			continue
		}

		if _, ok := c.conflicts[req.conflictKey]; ok {
			continue
		}

		req.send()

		if req.conflictKey != "" {
			c.conflicts[req.conflictKey] = struct{}{}
		}
		if req.responseFilter == nil && req.acknowledgeFilter == nil {
			c.queue = append(c.queue[:i], c.queue[i+1:]...)
		}
		c.notify()

		return req.requestSent
	}

	return nil
}

// OnSend blocks until a new conversation request becomes available at the end of the queue. You should call this method
// in the goroutine that handles gRPC stream Send method. ctx can be used to cancel the call.
func (c *Controller) OnSend(ctx context.Context) (*Ydb_Coordination.SessionRequest, error) {
	var req *Ydb_Coordination.SessionRequest
	for {
		select {
		case <-ctx.Done():
		case <-c.notifyChan:
			req = c.sendFront()
		}

		// Process ctx.Done() first to make sure we cancel the call if conversations are too chatty.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// We were notified but there were no messages in the queue. Just wait for more messages become available.
		if req != nil {
			break
		}
	}

	return req, nil
}

// OnRecv consumes a new conversation response and process with the corresponding conversation if any exists for it. The
// returned value indicates if any conversation considers the incoming message part of it or the controller is closed.
// You should call this method in the goroutine that handles gRPC stream Recv method.
func (c *Controller) OnRecv(resp *Ydb_Coordination.SessionResponse) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	notify := false
	handled := false
	for i := len(c.queue) - 1; i >= 0; i-- {
		req := c.queue[i]
		if req.requestSent == nil {
			continue
		}

		switch {
		case req.responseFilter != nil && req.responseFilter(req.requestSent, resp):
			if !req.canceled {
				req.succeed(resp)

				if req.conflictKey != "" {
					delete(c.conflicts, req.conflictKey)
					notify = true
				}

				c.queue = append(c.queue[:i], c.queue[i+1:]...)
			}

			handled = true
		case req.acknowledgeFilter != nil && req.acknowledgeFilter(req.requestSent, resp):
			if !req.canceled {
				if req.conflictKey != "" {
					delete(c.conflicts, req.conflictKey)
					notify = true
				}
			}

			handled = true
		case req.cancelRequestSent != nil && req.cancelFilter(req.cancelRequestSent, resp):
			if req.conflictKey != "" {
				delete(c.conflicts, req.conflictKey)
				notify = true
			}
			c.queue = append(c.queue[:i], c.queue[i+1:]...)
			handled = true
		}
	}

	if notify {
		c.notify()
	}

	return c.closed || handled
}

// OnDetach fails all non-idempotent conversations if there are any in the queue. You should call this method when the
// underlying gRPC stream of the session is closed.
func (c *Controller) OnDetach() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i := len(c.queue) - 1; i >= 0; i-- {
		req := c.queue[i]
		if !req.idempotent {
			req.fail(coordination.ErrOperationStatusUnknown)

			if req.requestSent != nil && req.conflictKey != "" {
				delete(c.conflicts, req.conflictKey)
			}

			c.queue = append(c.queue[:i], c.queue[i+1:]...)
		}
	}
}

// Close fails all conversations if there are any in the queue. It also does not allow pushing more conversations to the
// queue anymore. You may optionally specify the final conversation if needed.
func (c *Controller) Close(byeConversation *Conversation) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closed = true

	for i := len(c.queue) - 1; i >= 0; i-- {
		req := c.queue[i]
		if !req.canceled {
			req.fail(coordination.ErrSessionClosed)
		}
	}

	if byeConversation != nil {
		byeConversation.enqueue()
		c.queue = []*Conversation{byeConversation}
	}

	c.notify()
}

// OnAttach retries all idempotent conversations if there are any in the queue. You should call this method when the
// underlying gRPC stream of the session is connected.
func (c *Controller) OnAttach() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	notify := false
	for i := len(c.queue) - 1; i >= 0; i-- {
		req := c.queue[i]
		if req.idempotent && req.requestSent != nil {
			if req.conflictKey != "" {
				delete(c.conflicts, req.conflictKey)
			}

			// If the request has been canceled, re-send the cancellation message, otherwise re-send the original one.
			if req.canceled {
				req.cancelRequestSent = nil
			} else {
				req.requestSent = nil
			}
			notify = true
		}
	}

	if notify {
		c.notify()
	}
}

// Cancel the conversation if it has been sent and there is no response ready. This returns false if the response is
// ready and the caller may safely return it instead of canceling the conversation.
func (c *Controller) cancel(conversation *Conversation) bool {
	if conversation.cancelMessage == nil {
		return true
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// The context is canceled but the response is ready, return it anyway.
	if conversation.response != nil || conversation.responseErr != nil {
		return false
	}

	if conversation.requestSent != nil {
		conversation.cancel()
		c.notify()
	} else {
		// If the response has not been sent, just remove it from the queue.
		for i := len(c.queue) - 1; i >= 0; i-- {
			req := c.queue[i]
			if req == conversation {
				c.queue = append(c.queue[:i], c.queue[i+1:]...)

				break
			}
		}
	}

	return true
}

// Await waits until the conversation ends. ctx can be used to cancel the call.
func (c *Controller) Await(
	ctx context.Context,
	conversation *Conversation,
) (*Ydb_Coordination.SessionResponse, error) {
	select {
	case <-conversation.done:
	case <-ctx.Done():
	}

	if ctx.Err() != nil && c.cancel(conversation) {
		return nil, ctx.Err()
	}

	if conversation.responseErr != nil {
		return nil, conversation.responseErr
	}

	return conversation.response, nil
}

func (c *Conversation) enqueue() {
	c.requestSent = nil
	c.done = make(chan struct{})
}

func (c *Conversation) send() {
	c.requestSent = c.message()
}

func (c *Conversation) sendCancel() {
	c.cancelRequestSent = c.cancelMessage(c.requestSent)
}

func (c *Conversation) succeed(response *Ydb_Coordination.SessionResponse) {
	c.response = response
	close(c.done)
}

func (c *Conversation) fail(err error) {
	c.responseErr = err
	close(c.done)
}

func (c *Conversation) cancel() {
	c.canceled = true
	close(c.done)
}
