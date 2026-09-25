package topicwriter

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicwriterinternal"
)

type (
	Message = topicwriterinternal.PublicMessage
)

var (
	// Deprecated: the error will not be returned. Topic writer allow overflow queue for single call.
	// Will be removed after October 2025.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	ErrQueueLimitExceed                      = topicwriterinternal.ErrPublicQueueIsFull
	ErrMessagesPutToInternalQueueBeforeError = topicwriterinternal.ErrPublicMessagesPutToInternalQueueBeforeError
)

// Writer represent write session to topic
// It handles connection problems, reconnect to server when need and resend buffered messages
type Writer struct {
	inner *topicwriterinternal.WriterReconnector
}

// PublicInitialInfo is an information about writer after initialize
type PublicInitialInfo struct {
	LastSeqNum int64
}

// NewWriter create new writer from internal type. Used internally only.
func NewWriter(writer *topicwriterinternal.WriterReconnector) *Writer {
	return &Writer{
		inner: writer,
	}
}

// Write send messages to topic
// return after save messages into buffer in async mode (default) and after ack from server in sync mode.
// see topicoptions.WithSyncWrite
//
// The method will wait first initial connection even for async mode, that mean first write may be slower.
// especially when connection has problems.
//
// It returns ErrQueueLimitExceed (must be checked by errors.Is)
// if ctx cancelled before messages put to internal buffer or try to add more messages, that can be put to queue.
// If err != nil you can check errors.Is(err, ErrMessagesPutToInternalQueueBeforeError) for check if the messages
// put to buffer before error. It means that it is messages can be delivered to the server.
func (w *Writer) Write(ctx context.Context, messages ...Message) error {
	return w.inner.Write(ctx, messages)
}

// WaitInit waits until the reader is initialized
// or an error occurs, return PublicInitialInfo and err
func (w *Writer) WaitInit(ctx context.Context) (err error) {
	_, err = w.inner.WaitInit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// WaitInitInfo waits until the reader is initialized
// or an error occurs, return PublicInitialInfo and err
func (w *Writer) WaitInitInfo(ctx context.Context) (info PublicInitialInfo, err error) {
	privateInfo, err := w.inner.WaitInit(ctx)
	if err != nil {
		return PublicInitialInfo{}, err
	}
	publicInfo := PublicInitialInfo{LastSeqNum: privateInfo.LastSeqNum}

	return publicInfo, nil
}

// Close will flush rested messages from buffer and close the writer.
// You can't write new messages after call Close
func (w *Writer) Close(ctx context.Context) error {
	return w.inner.Close(ctx)
}

// Flush waits till all in-flight messages are acknowledged.
func (w *Writer) Flush(ctx context.Context) error {
	return w.inner.Flush(ctx)
}

// TxWriter used for send messages to the transaction
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
type TxWriter struct {
	inner *topicwriterinternal.WriterWithTransaction
}

func NewTxWriterInternal(w *topicwriterinternal.WriterWithTransaction) *TxWriter {
	return &TxWriter{inner: w}
}

// Write messages to the transaction
//
// It has not retries. If fails - needs to retry full transaction, as with any other
// error with table.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (w *TxWriter) Write(ctx context.Context, messages ...Message) error {
	return w.inner.Write(ctx, messages...)
}

// WaitInit waits until the reader is initialized
// or an error occurs, return PublicInitialInfo and err
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (w *TxWriter) WaitInit(ctx context.Context) (err error) {
	_, err = w.inner.WaitInit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// WaitInitInfo waits until the reader is initialized
// or an error occurs, return PublicInitialInfo and err
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (w *TxWriter) WaitInitInfo(ctx context.Context) (info PublicInitialInfo, err error) {
	privateInfo, err := w.inner.WaitInit(ctx)
	if err != nil {
		return PublicInitialInfo{}, err
	}
	publicInfo := PublicInitialInfo{LastSeqNum: privateInfo.LastSeqNum}

	return publicInfo, nil
}
