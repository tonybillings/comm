package stream

import (
	"io"
)

type DataReader struct {
	data      []byte
	dataIndex int
}

func (r *DataReader) Read(buffer []byte) (int, error) {
	dataLen := len(r.data)
	if r.dataIndex >= dataLen {
		return -1, io.EOF
	}

	bufLen := len(buffer)
	endIndex := r.dataIndex + bufLen
	if endIndex > dataLen {
		endIndex = dataLen
	}

	index := 0
	for i := r.dataIndex; i < endIndex; i++ {
		buffer[index] = r.data[i]
		index++
	}

	count := endIndex - r.dataIndex
	r.dataIndex += count
	return count, nil
}

func NewDataReader(data []byte) *DataReader {
	r := &DataReader{}
	r.data = data
	return r
}
