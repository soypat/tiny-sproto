package sproto

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"
)

var interestingPacket [256]byte

func TestMain(m *testing.M) {
	const plen = len(interestingPacket)
	for i := 0; i < plen; i++ {
		interestingPacket[i] = byte(i)
	}
	interestingPacket[0] = sof
	interestingPacket[1] = esc
	interestingPacket[19] = esc
	interestingPacket[20] = sof ^ escxor
	interestingPacket[plen-2] = esc
	interestingPacket[plen-1] = eof
	os.Exit(m.Run())
}

func TestFrameOneshotLoopback(t *testing.T) {
	data := interestingPacket
	f := NewFrame(make([]byte, 300))
	err := f.SetData(data[:])
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	n, err := io.Copy(buffer, f)
	// n, err := f.Read()
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

func TestFrameLoopback(t *testing.T) {
	data := []byte{0x00, 0x01, sof, 0x03, esc}
	f := NewFrame(make([]byte, 256))
	err := f.SetData(data)
	if err != nil {
		t.Fatal(err)
	}
	var buffer [256]byte
	n, err := f.Read(buffer[:])
	if err != nil {
		t.Fatal(err)
	}
	wire := buffer[:n]
	_, err = f.ParseNext(bufio.NewReader(bytes.NewReader(wire)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, f.data) {
		t.Errorf("got    % x\nexpect % x", f.data, data)
	}
}
