package stream

import (
	"io"
	"strings"
	"sync/atomic"
)

const (
	defaultChanBufferSize = 1024
	readBufferSize        = 1500
)

// StringStream Produces instances of string from a byte stream containing
// C strings (null-terminated/delimited strings).
type StringStream struct {
	reader      io.Reader
	streamChan  chan string
	processChan chan bool
	err         atomic.Value
}

func (s *StringStream) Stream(reader io.Reader, bufferSize ...int) <-chan string {
	if s.reader == nil {
		s.reader = reader
		if len(bufferSize) > 0 {
			s.streamChan = make(chan string, bufferSize[0])
		} else {
			s.streamChan = make(chan string, defaultChanBufferSize)
		}

		go s.process()
	}

	return s.streamChan
}

func (s *StringStream) Error() error {
	e := s.err.Load()
	if e != nil {
		return e.(error)
	}
	return nil
}

func (s *StringStream) Close() error {
	if s.processChan != nil {
		select {
		case <-s.processChan:
			break
		default:
			close(s.processChan)
		}
	}
	return s.Error()
}

func (s *StringStream) process() {
	s.processChan = make(chan bool)
	buffer := make([]byte, readBufferSize)
	sb := strings.Builder{}

	for {
		select {
		case <-s.processChan:
			s.processChan = nil
			close(s.streamChan)
			return
		default:
		}

		count, err := s.reader.Read(buffer)
		if err != nil {
			s.err.Store(err)

			close(s.processChan)
			s.processChan = nil

			close(s.streamChan)
			return
		}

		for i := 0; i < count; i++ {
			b := buffer[i]
			if b > 0 {
				sb.WriteByte(buffer[i])
			} else {
				if sb.Len() > 0 {
					select {
					case s.streamChan <- sb.String():
						break
					default:
					}
					sb.Reset()
				}
			}

		}
	}
}

func String(reader io.Reader, bufferSize ...int) <-chan string {
	s := &StringStream{}
	return s.Stream(reader, bufferSize...)
}
