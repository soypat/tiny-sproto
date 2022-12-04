package sproto

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

var interestingPacket [256]byte

func TestMain(m *testing.M) {
	const plen = len(interestingPacket)
	for i := 0; i < plen; i++ {
		c := byte(i)
		if c == sof || c == esc {
			c = 0
		}
		interestingPacket[i] = c
	}

	// interestingPacket[0] = sof
	// interestingPacket[1] = esc
	// interestingPacket[19] = esc
	// interestingPacket[20] = sof ^ escxor
	// interestingPacket[plen-2] = esc
	// interestingPacket[plen-1] = eof
	os.Exit(m.Run())
}

func TestFrameLoopbackCopy(t *testing.T) {
	data := interestingPacket
	f := NewFrame(make([]byte, 300))
	err := f.SetData(data[:])
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	n, err := io.Copy(buffer, f)
	if err != nil || n == 0 {
		t.Fatal(n, err)
	}
	wire := buffer.Bytes()
	_, err = f.ParseNext(bufio.NewReader(bytes.NewReader(wire)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data[:], f.data) {
		t.Errorf("got    % x\nexpect % x", f.data, data)
	}
}

func TestFrameMultiShotLoopback(t *testing.T) {
	data := interestingPacket[:]
	f := NewFrame(make([]byte, 300))
	err := f.SetData(data[:])
	if err != nil {
		t.Fatal(err)
	}

	var buffer [64]byte
	var wire []byte
	for {
		n, err := f.Read(buffer[:])
		wire = append(wire, buffer[:n]...)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatal(err)
			break
		}
	}
	n, err := f.ParseNext(bufio.NewReader(bytes.NewReader(wire)))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(wire) {
		t.Error("parsed bytes not equal to sent over wire", n, len(wire))
	}
	if !bytes.Equal(data, f.data) {
		t.Errorf("got    % x\nexpect % x", f.data, data)
	}
}
