package topicwriterinternal

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type WriterWithTransaction struct {
	streamWriter *WriterReconnector
	tx           tx.Transaction
	tracer       *trace.Topic
}

func NewTopicWriterTransaction(w *WriterReconnector, tx tx.Transaction, tracer *trace.Topic) *WriterWithTransaction {
	res := &WriterWithTransaction{
		streamWriter: w,
		tx:           tx,
	}
	if tracer == nil {
		res.tracer = &trace.Topic{}
	} else {
		res.tracer = tracer
	}

	tx.OnBeforeCommit(res.onBeforeCommitTransaction)
	tx.OnCompleted(res.onTransactionCompleted)

	return res
}

func (w *WriterWithTransaction) onBeforeCommitTransaction(ctx context.Context) (err error) {
	traceCtx := ctx
	onDone := trace.TopicOnWriterBeforeCommitTransaction(
		w.tracer,
		&traceCtx,
		w.tx.SessionID(),
		w.streamWriter.GetSessionID(),
		w.tx.ID(),
	)

	defer func() {
		onDone(err, w.streamWriter.GetSessionID())
		ctx = traceCtx
	}()

	// wait message flushing
	return w.Close(ctx)
}

func (w *WriterWithTransaction) WaitInit(ctx context.Context) (info InitialInfo, err error) {
	return w.streamWriter.WaitInit(ctx)
}

func (w *WriterWithTransaction) Write(ctx context.Context, messages ...PublicMessage) error {
	for i := range messages {
		messages[i].tx = w.tx
	}

	return w.streamWriter.Write(ctx, messages)
}

func (w *WriterWithTransaction) Close(ctx context.Context) error {
	return w.streamWriter.Close(ctx)
}

func (w *WriterWithTransaction) onTransactionCompleted(err error) {
	// after transaction finished by any reason - the writer closed without flush
	noNeedFlushCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = w.Close(noNeedFlushCtx)
}
