package minissdpd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

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

func TestTooManyLengthBytes(t *testing.T) {
	buf := &bytes.Buffer{}
	for i := 0; i <= MaxLengthBytes+1; i++ {
		buf.WriteByte(0x80)
	}
	_, err := DecodeStringLength(buf)
	if err != errTooLong {
		t.Fatalf("expected errTooLong, got: %v", err)
	}
}

func TestServiceEncode(t *testing.T) {
	s := Service{
		Type:     "dummytype",
		USN:      strings.Repeat("1", 128),
		Server:   "dummy 1.0",
		Location: "http://127.0.0.1/setup.xml",
	}

	expect := make([]byte, 0)
	expect = append(expect, byte(len(s.Type)))
	expect = append(expect, s.Type...)

	// USN needs 2 bytes for length of 128
	usnLen := []byte{0x81, byte(len(s.USN)) & 0x7f}
	expect = append(expect, usnLen...)
	expect = append(expect, s.USN...)

	expect = append(expect, byte(len(s.Server)))
	expect = append(expect, s.Server...)
	expect = append(expect, byte(len(s.Location)))
	expect = append(expect, s.Location...)

	buf := &bytes.Buffer{}
	b, err := s.EncodeTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	if b != len(expect) {
		t.Fatalf("expected %d bytes, got %d", len(expect), b)
	}

	out := buf.Bytes()
	if !reflect.DeepEqual(out, expect) {
		t.Logf("expected: %v\n", expect)
		t.Logf("     got: %v\n", out)
		t.Fatal("mismatched bytes after encode")
	}
}

func TestDecodeServices(t *testing.T) {

	stream := []byte{
		0x03,
		0x15, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x38, 0x30, 0x30, 0x31, 0x1d, 0x75, 0x72, 0x6e, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x31, 0x3a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x65, 0x3a, 0x31, 0x18, 0x75, 0x75, 0x69, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x31,
		0x15, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x38, 0x30, 0x30, 0x32, 0x1d, 0x75, 0x72, 0x6e, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x32, 0x3a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x65, 0x3a, 0x31, 0x18, 0x75, 0x75, 0x69, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x32,
		0x15, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x38, 0x30, 0x30, 0x33, 0x1d, 0x75, 0x72, 0x6e, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x33, 0x3a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x65, 0x3a, 0x31, 0x18, 0x75, 0x75, 0x69, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x33,
	}

	expect := []Service{
		{"urn:Type1:device:controllee:1", "uuid:0000-0000-0000-0001", "", "http://127.0.0.1:8001"},
		{"urn:Type2:device:controllee:1", "uuid:0000-0000-0000-0002", "", "http://127.0.0.1:8002"},
		{"urn:Type3:device:controllee:1", "uuid:0000-0000-0000-0003", "", "http://127.0.0.1:8003"},
	}

	buf := bytes.NewReader(stream)
	out, err := decodeServices(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(out, expect) {
		t.Logf("expected: %v\n", expect)
		t.Logf("     got: %v\n", out)
		t.Fatal("mismatched services after decode")
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
