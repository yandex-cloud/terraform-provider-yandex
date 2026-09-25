package rawtopicreader

//go:generate mockgen -destination ../rawtopicreadermock/stream_reader_stream_interface_mock.go --typed -package rawtopicreadermock -write_package_comment=false --typed . TopicReaderStreamInterface

type TopicReaderStreamInterface interface {
	Recv() (ServerMessage, error)
	Send(msg ClientMessage) error
	CloseSend() error
}
