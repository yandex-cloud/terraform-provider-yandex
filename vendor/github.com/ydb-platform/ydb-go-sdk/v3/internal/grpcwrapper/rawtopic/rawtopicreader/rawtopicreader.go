package rawtopicreader

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var ErrUnexpectedMessageType = errors.New("unexpected message type")

type GrpcStream interface {
	Send(messageNew *Ydb_Topic.StreamReadMessage_FromClient) error
	Recv() (*Ydb_Topic.StreamReadMessage_FromServer, error)
	CloseSend() error
}

type StreamReader struct {
	Stream   GrpcStream
	ReaderID int64

	Tracer              *trace.Topic
	sessionID           string
	sentMessageCount    int
	receiveMessageCount int
}

func (s StreamReader) CloseSend() error {
	return s.Stream.CloseSend()
}

//nolint:funlen
func (s StreamReader) Recv() (_ ServerMessage, resErr error) {
	grpcMess, err := s.Stream.Recv()

	defer func() {
		s.receiveMessageCount++

		trace.TopicOnReaderReceiveGRPCMessage(
			s.Tracer,
			s.ReaderID,
			s.sessionID,
			s.receiveMessageCount,
			grpcMess,
			resErr,
		)
	}()

	if xerrors.Is(err, io.EOF) {
		return nil, err
	}
	if err != nil {
		if !xerrors.IsErrorFromServer(err) {
			err = xerrors.Transport(err)
		}

		return nil, err
	}

	var meta rawtopiccommon.ServerMessageMetadata
	if err = meta.MetaFromStatusAndIssues(grpcMess); err != nil {
		return nil, err
	}
	if !meta.Status.IsSuccess() {
		return nil, xerrors.WithStackTrace(fmt.Errorf("ydb: bad status from topic server: %v", meta.Status))
	}

	switch m := grpcMess.GetServerMessage().(type) {
	case *Ydb_Topic.StreamReadMessage_FromServer_InitResponse:
		s.sessionID = m.InitResponse.GetSessionId()

		resp := &InitResponse{}
		resp.ServerMessageMetadata = meta
		resp.fromProto(m.InitResponse)

		return resp, nil
	case *Ydb_Topic.StreamReadMessage_FromServer_ReadResponse:
		resp := &ReadResponse{}
		resp.ServerMessageMetadata = meta
		if err = resp.fromProto(m.ReadResponse); err != nil {
			return nil, err
		}

		return resp, nil
	case *Ydb_Topic.StreamReadMessage_FromServer_StartPartitionSessionRequest:
		resp := &StartPartitionSessionRequest{}
		resp.ServerMessageMetadata = meta
		if err = resp.fromProto(m.StartPartitionSessionRequest); err != nil {
			return nil, err
		}

		return resp, nil
	case *Ydb_Topic.StreamReadMessage_FromServer_StopPartitionSessionRequest:
		req := &StopPartitionSessionRequest{}
		req.ServerMessageMetadata = meta
		if err = req.fromProto(m.StopPartitionSessionRequest); err != nil {
			return nil, err
		}

		return req, nil

	case *Ydb_Topic.StreamReadMessage_FromServer_EndPartitionSession:
		req := &EndPartitionSession{}
		req.ServerMessageMetadata = meta
		if err = req.fromProto(m.EndPartitionSession); err != nil {
			return nil, err
		}

		return req, nil

	case *Ydb_Topic.StreamReadMessage_FromServer_CommitOffsetResponse:
		resp := &CommitOffsetResponse{}
		resp.ServerMessageMetadata = meta
		if err = resp.fromProto(m.CommitOffsetResponse); err != nil {
			return nil, err
		}

		return resp, nil
	case *Ydb_Topic.StreamReadMessage_FromServer_PartitionSessionStatusResponse:
		resp := &PartitionSessionStatusResponse{}
		resp.ServerMessageMetadata = meta
		if err = resp.fromProto(m.PartitionSessionStatusResponse); err != nil {
			return nil, err
		}

		return resp, nil
	case *Ydb_Topic.StreamReadMessage_FromServer_UpdateTokenResponse:
		resp := &UpdateTokenResponse{}
		resp.ServerMessageMetadata = meta
		resp.MustFromProto(m.UpdateTokenResponse)

		return resp, nil
	default:
		return nil, xerrors.WithStackTrace(fmt.Errorf(
			"ydb: receive unexpected message (%v): %w",
			reflect.TypeOf(grpcMess.GetServerMessage()),
			ErrUnexpectedMessageType,
		))
	}
}

//nolint:funlen
func (s StreamReader) Send(msg ClientMessage) (resErr error) {
	defer func() {
		resErr = xerrors.Transport(resErr)
	}()

	var grpcMess *Ydb_Topic.StreamReadMessage_FromClient
	switch m := msg.(type) {
	case *InitRequest:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_InitRequest{InitRequest: m.toProto()},
		}

	case *ReadRequest:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_ReadRequest{ReadRequest: m.toProto()},
		}
	case *StartPartitionSessionResponse:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_StartPartitionSessionResponse{
				StartPartitionSessionResponse: m.toProto(),
			},
		}
	case *StopPartitionSessionResponse:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_StopPartitionSessionResponse{
				StopPartitionSessionResponse: m.toProto(),
			},
		}

	case *CommitOffsetRequest:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_CommitOffsetRequest{
				CommitOffsetRequest: m.toProto(),
			},
		}

	case *PartitionSessionStatusRequest:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_PartitionSessionStatusRequest{
				PartitionSessionStatusRequest: m.toProto(),
			},
		}

	case *UpdateTokenRequest:
		grpcMess = &Ydb_Topic.StreamReadMessage_FromClient{
			ClientMessage: &Ydb_Topic.StreamReadMessage_FromClient_UpdateTokenRequest{
				UpdateTokenRequest: m.ToProto(),
			},
		}
	}

	if grpcMess == nil {
		resErr = xerrors.WithStackTrace(fmt.Errorf("ydb: send unexpected message type: %v", reflect.TypeOf(msg)))
	} else {
		resErr = s.Stream.Send(grpcMess)
	}

	s.sentMessageCount++
	trace.TopicOnReaderSentGRPCMessage(
		s.Tracer,
		s.ReaderID,
		s.sessionID,
		s.sentMessageCount,
		grpcMess,
		resErr,
	)

	return resErr
}

type ClientMessage interface {
	isClientMessage()
}

type clientMessageImpl struct{}

func (*clientMessageImpl) isClientMessage() {}

type ServerMessage interface {
	isServerMessage()
	StatusData() rawtopiccommon.ServerMessageMetadata
	SetStatus(status rawydb.StatusCode)
}

type serverMessageImpl struct{}

func (*serverMessageImpl) isServerMessage() {}
