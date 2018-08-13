package minissdpd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
)

func EncodeLengthReverse(length int, w io.Writer) error {
	if length < 0 {
		return errInvalidLength
	}

	n := uint(length)
	b := make([]byte, 0, MaxLengthBytes)

	for i := uint(MaxLengthBytes - 1); i > 0; i-- {
		x := pow(128, i)
		if n >= x {
			v := byte(n>>(7*i) | 0x80)
			b = append(b, v)
		}
	}
	b = append(b, byte(n&0x7f))

	_, err := w.Write(b)
	if err != nil {
		return fmt.Errorf("could not write to buffer: %v", err)
	}
	return nil
}

func EncodeLengthNoLoop(length int, w io.Writer) error {
	if length < 0 {
		return errInvalidLength
	}

	n := uint(length)
	b := make([]byte, 0, 5)

	if n >= 268435456 {
		b = append(b, byte(n>>28|0x80))
	}
	if n >= 2097152 {
		b = append(b, byte(n>>21|0x80))
	}
	if n >= 16384 {
		b = append(b, byte(n>>14|0x80))
	}
	if n >= 128 {
		b = append(b, byte(n>>7|0x80))
	}
	b = append(b, byte(n&0x7f))

	_, err := w.Write(b)
	if err != nil {
		return fmt.Errorf("could not write to buffer: %v", err)
	}
	return nil
}

func EncodeLengthNoAppend(length int, w io.Writer) error {
	if length < 0 {
		return errInvalidLength
	}

	n := uint(length)
	b := []byte{
		byte(n>>28 | 0x80),
		byte(n>>21 | 0x80),
		byte(n>>14 | 0x80),
		byte(n>>7 | 0x80),
		byte(n & 0x7f),
	}

	var i int
	if n >= 268435456 {
		i = 0
	} else if n >= 2097152 {
		i = 1
	} else if n >= 16384 {
		i = 2
	} else if n >= 128 {
		i = 3
	} else {
		i = 4
	}

	_, err := w.Write(b[i:])
	if err != nil {
		return fmt.Errorf("could not write to buffer: %v", err)
	}
	return nil
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		length  int
		encoded []byte
	}{
		{0, []byte{0}},
		{1, []byte{1}},
		{127, []byte{127}},
		{128, []byte{129, 0}},
		{268435456, []byte{129, 128, 128, 128, 0}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Encode %d", test.length), func(t *testing.T) {
			b := &bytes.Buffer{}
			err := EncodeStringLength(test.length, b)
			if err != nil {
				t.Fatal(err)
			}
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(out, test.encoded) {
				t.Errorf("encode expected %#v, got %#v", test.encoded, out)
			}
		})

		t.Run(fmt.Sprintf("Decode %d", test.length), func(t *testing.T) {
			b := bytes.NewBuffer(test.encoded)
			n, err := DecodeStringLength(b)
			if err != nil {
				t.Fatal(err)
			}

			if n != test.length {
				t.Fatalf("expected length of %d, got %d", test.length, n)
			}
		})
	}
}

func TestNegativeLength(t *testing.T) {
	err := EncodeStringLength(-1, nil)
	if err == nil {
		t.Fatal("expected error for negative length")
	}
	if err != errInvalidLength {
		t.Fatalf("Expected ErrInvalidLength, got: %v", err)
	}
}

func TestNilWriter(t *testing.T) {
	err := EncodeStringLength(1, nil)
	if err == nil {
		t.Fatal("expected error for nil writer")
	}
}

func BenchmarkEncodeShort(b *testing.B) {
	bb := make([]byte, 1)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeStringLength(0, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeLong(b *testing.B) {
	bb := make([]byte, 5)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeStringLength(268435456, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeReverseShort(b *testing.B) {
	bb := make([]byte, 1)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeLengthReverse(0, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeReverseLong(b *testing.B) {
	bb := make([]byte, 5)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeLengthReverse(268435456, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeNoLoopShort(b *testing.B) {
	bb := make([]byte, 1)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeLengthNoLoop(0, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeNoLoopLong(b *testing.B) {
	bb := make([]byte, 5)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeLengthNoLoop(268435456, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeNoAppendShort(b *testing.B) {
	bb := make([]byte, 1)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeLengthNoAppend(0, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeNoAppendLong(b *testing.B) {
	bb := make([]byte, 5)
	buf := bytes.NewBuffer(bb)

	for i := 0; i < b.N; i++ {
		err := EncodeLengthNoAppend(268435456, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
