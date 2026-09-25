package topicreadercommon

import (
	"io"
)

type oneTimeReader struct {
	err    error
	reader io.Reader

	// prevent early create decoder, because it can consume a lot of memory
	// https://github.com/ydb-platform/ydb-go-sdk/issues/1341
	readerMaker readerMaker
}

type readerMaker func() io.Reader

func newOneTimeReader(readerMaker readerMaker) oneTimeReader {
	return oneTimeReader{
		readerMaker: readerMaker,
	}
}

func newOneTimeReaderFromReader(reader io.Reader) oneTimeReader {
	maker := func() io.Reader { return reader }

	return newOneTimeReader(maker)
}

func (s *oneTimeReader) Read(p []byte) (n int, err error) {
	if s.err != nil {
		return 0, s.err
	}

	if s.reader == nil {
		s.reader = s.readerMaker()
	}

	n, err = s.reader.Read(p)
	if err != nil {
		s.err = err
		s.reader = nil
	}

	return n, err
}
