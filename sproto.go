package sproto

import (
	"errors"
	"fmt"
	"io"
)

type state uint8

const (
	START state = iota
	READ
	ESCAPE
	END
)

const (
	sof    byte = 0x7d
	eof    byte = sof
	esc    byte = 0x7e
	escxor byte = 0x20
)

func NewFrame(buffer []byte) *Frame {
	return &Frame{
		data: buffer[:0:len(buffer)],
	}
}

func (f *Frame) SetData(data []byte) error {
	if len(data) > f.Size() {
		return errors.New("data length greater than buffer capacity")
	}
	f.setLen(len(data))
	copy(f.data, data)
	return nil
}

func (f *Frame) Size() int { return cap(f.data) }

func (f *Frame) setLen(l int) { f.data = f.data[:l] }

type Frame struct {
	ptr  int
	data []byte
}

func (f *Frame) ParseNext(r io.ByteReader) (n int, err error) {
	status := START
	f.setLen(f.Size()) // Max out size.
readloop:
	for dataPtr := 0; dataPtr < f.Size(); {
		c, err := r.ReadByte()
		if err != nil {
			f.setLen(0)
			break
		}
		n++
		switch status {
		case START:
			if c == sof {
				status = READ
			}

		case READ:
			if c == esc {
				status = ESCAPE
			} else if c == eof {
				// Succesfully read entire frame.
				status = END
				f.setLen(dataPtr)
				f.ptr = 0
				break readloop
			} else {
				f.data[dataPtr] = c
				dataPtr++
			}

		case ESCAPE:
			f.data[dataPtr] = c ^ escxor
			dataPtr++
			status = READ
		default:
			panic("unexpected state")
		}
	}
	if status != END {
		if len(f.data) == 0 {
			err = fmt.Errorf("buffer too small (%d) for message", f.Size())
		} else {
			err = fmt.Errorf("did not reach END byte(%d):%w", status, err)
		}
	}
	return n, err
}

func (f *Frame) Read(b []byte) (n int, err error) {
	frameLen := len(f.data)
	dataPtr := f.ptr
	toWrite := frameLen - dataPtr
	if toWrite == 0 {
		return 0, io.EOF
	}
	if dataPtr == 0 {
		b[0] = sof
		n++
	}
	for ; dataPtr < toWrite; dataPtr++ {
		c := f.data[dataPtr]
		remaining := len(b) - n
		if remaining < 2 {
			break
		}
		switch c {
		case sof, esc:
			b[n] = esc
			b[n+1] = c ^ escxor
			n += 2
		default:
			b[n] = c
			n += 1
		}
	}
	if dataPtr == toWrite {
		b[n] = eof
		n++
	}
	f.ptr = dataPtr
	return n, nil
}
