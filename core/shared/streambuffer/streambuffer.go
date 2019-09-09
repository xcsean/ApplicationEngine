package streambuffer

import (
	"fmt"
	"io"
)

// StreamBuffer stream buffer for read and write
type StreamBuffer struct {
	reader io.Reader
	buffer []byte
	start  int
	end    int
}

// New new a buffer, attach to a Reader
func New(reader io.Reader, len int) *StreamBuffer {
	buffer := make([]byte, len)
	return &StreamBuffer{reader, buffer, 0, 0}
}

// Read read from the Reader attached, return the bytes read
func (sb *StreamBuffer) Read() (int, error) {
	n, err := sb.reader.Read(sb.buffer[sb.end:])
	if err != nil {
		return n, err
	}
	sb.end += n
	return n, nil
}

// Length return length of current buffer
func (sb *StreamBuffer) Length() int {
	return sb.end - sb.start
}

// Peek try to get the data for length, has no side-effect
func (sb *StreamBuffer) Peek(n int) ([]byte, error) {
	if sb.end-sb.start >= n {
		buffer := sb.buffer[sb.start:(sb.start + n)]
		return buffer, nil
	}
	return nil, fmt.Errorf("not enough data to peek")
}

// Fetch try to fetch 'n' bytes with skipping 'offset' bytes, has side-effect
func (sb *StreamBuffer) Fetch(offset, n int) []byte {
	sb.start += offset
	buffer := sb.buffer[sb.start:(sb.start + n)]
	sb.start += n
	return buffer
}

// Shift shift data to the start direction
func (sb *StreamBuffer) Shift() {
	if sb.start == 0 {
		return
	}
	copy(sb.buffer, sb.buffer[sb.start:sb.end])
	sb.end -= sb.start
	sb.start = 0
}
