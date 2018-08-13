package minissdpd

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const testSocket = "test.sock"

func newSocket(t *testing.T) (path string, close func() error) {
	dir, err := ioutil.TempDir("", "ssdpc")
	if err != nil {
		t.Fatalf("could not create temp socket dir: %v", err)
	}

	path = filepath.Join(dir, testSocket)
	conn, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("could not open test socket: %v", err)
	}

	close = func() error {
		os.RemoveAll(dir)
		return conn.Close()
	}

	return path, close
}

func TestClientConnections(t *testing.T) {
	sock, close := newSocket(t)
	defer close()

	c := Client{
		SocketPath: sock,
	}

	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}

	if err := c.Connect(); err != errOpen {
		t.Fatalf("expected errOpen connceting to open server, got %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("client close error: %v", err)
	}

	if err := c.Connect(); err != nil {
		t.Fatalf("connection reuse failed with error: %v", err)
	}
	c.Close()
}

func TestClientNilConn(t *testing.T) {
	c := &Client{}

	_, err := c.Write(nil)
	if err != errNilConn {
		t.Fatalf("expected errNilCon, got %v", err)
	}
}

func TestClientWriteString(t *testing.T) {
	testString := "minissdp"

	expect := []byte{byte(len(testString))}
	expect = append(expect, testString...)

	client, server := net.Pipe()
	reader := make(chan []byte)

	go func() {
		err := server.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Fatalf("could not set server read deadline: %v", err)
		}
		buf := make([]byte, len(expect))
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}
		reader <- buf
	}()

	c := Client{
		conn: client,
	}

	n, err := c.WriteString(testString)
	if err != nil {
		t.Fatal(err)
	}

	// WriteString will encode the string, so we need to add a byte
	// for the string length prefix
	if n != len(testString)+1 {
		t.Fatalf("expected to write %d bytes, wrote %d", len(testString)+1, n)
	}

	out := <-reader

	if len(out) != len(expect) {
		t.Fatalf("expected to read %d bytes, got %d", len(expect), len(out))
	}

	if string(out[1:]) != testString {
		t.Fatalf("unexpected response from mock server: %#v", string(out[1:]))
	}
}

func TestClientRegister(t *testing.T) {
	service := Service{
		Type:     "urn:Dummy:device:controllee:1",
		USN:      "1234-1234-1234-1234",
		Server:   "Dummy 1.0",
		Location: "http://127.0.0.1/setup.xml",
	}

	expect := []byte{RequestTypeRegister}
	expect = append(expect, byte(len(service.Type)))
	expect = append(expect, service.Type...)
	expect = append(expect, byte(len(service.USN)))
	expect = append(expect, service.USN...)
	expect = append(expect, byte(len(service.Server)))
	expect = append(expect, service.Server...)
	expect = append(expect, byte(len(service.Location)))
	expect = append(expect, service.Location...)

	client, server := net.Pipe()
	reader := make(chan []byte)

	go func() {
		err := server.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Fatalf("could not set server read deadline: %v", err)
		}
		buf := make([]byte, len(expect))
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}
		reader <- buf
	}()

	c := Client{
		conn: client,
	}

	err := c.RegisterService(service)
	if err != nil {
		t.Fatal(err)
	}

	out := <-reader

	if len(out) != len(expect) {
		t.Fatalf("expected to read %d bytes, got %d", len(expect), len(out))
	}

	if !reflect.DeepEqual(out, expect) {
		t.Fatalf("unexpected response from mock server: %#v", string(out))
	}
}

func TestClientGetByType(t *testing.T) {
	filter := "urn:Dummy:device:controllee:1"
	expect := []byte{byte(len(filter))}
	expect = append(expect, filter...)

	client, server := net.Pipe()
	defer server.Close()
	defer client.Close()

	reader := make(chan []byte)

	go func() {
		err := server.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Fatalf("could not set server read deadline: %v", err)
		}

		// First read the RequestType byte (because net.Pipe Connections don't buffer
		// and we use multiple calls to Write())
		buf := make([]byte, 1)
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}
		if buf[0] != RequestTypeByType {
			t.Fatalf("Expected first byte to be %x, got %x", RequestTypeByType, buf[0])
		}

		// Then read the encoded request string
		buf = make([]byte, len(expect))
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}

		_, err = server.Write([]byte{0})
		if err != nil {
			t.Fatalf("server write error: %v", err)
		}

		reader <- buf
	}()

	c := Client{
		conn: client,
	}

	services, err := c.GetServicesByType(filter)
	if err != nil {
		t.Fatal(err)
	}

	out := <-reader

	if !reflect.DeepEqual(out, expect) {
		t.Fatalf("unexpected response from mock server: %#v", string(out))
	}

	if len(services) != 0 {
		t.Fatal("unexpected services returned from mock server")
	}
}

func TestClientGetByUSN(t *testing.T) {
	filter := "uuid:1111-2222-3333-4444"
	expect := []byte{byte(len(filter))}
	expect = append(expect, filter...)

	client, server := net.Pipe()
	defer server.Close()
	defer client.Close()

	reader := make(chan []byte)

	go func() {
		err := server.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Fatalf("could not set server read deadline: %v", err)
		}

		// First read the RequestType byte (because net.Pipe Connections don't buffer
		// and we use multiple calls to Write())
		buf := make([]byte, 1)
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}
		if buf[0] != RequestTypeByUSN {
			t.Fatalf("Expected first byte to be %x, got %x", RequestTypeByUSN, buf[0])
		}

		// Then read the encoded request string
		buf = make([]byte, len(expect))
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}

		_, err = server.Write([]byte{0})
		if err != nil {
			t.Fatalf("server write error: %v", err)
		}

		reader <- buf
	}()

	c := Client{
		conn: client,
	}

	services, err := c.GetServicesByUSN(filter)
	if err != nil {
		t.Fatal(err)
	}

	out := <-reader

	if !reflect.DeepEqual(out, expect) {
		t.Fatalf("unexpected response from mock server: %#v", string(out))
	}

	if len(services) != 0 {
		t.Fatal("unexpected services returned from mock server")
	}
}

// This test is basically a copy of TestDecideServices, with the addition of
// validating that GetServicesAll will send the correct bytes to the server
func TestClientGetAll(t *testing.T) {
	stream := []byte{
		0x03,
		0x15, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x38, 0x30, 0x30, 0x31, 0x1d, 0x75, 0x72, 0x6e, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x31, 0x3a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x65, 0x3a, 0x31, 0x18, 0x75, 0x75, 0x69, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x31,
		0x15, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x38, 0x30, 0x30, 0x32, 0x1d, 0x75, 0x72, 0x6e, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x32, 0x3a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x65, 0x3a, 0x31, 0x18, 0x75, 0x75, 0x69, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x32,
		0x15, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x38, 0x30, 0x30, 0x33, 0x1d, 0x75, 0x72, 0x6e, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x33, 0x3a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x65, 0x3a, 0x31, 0x18, 0x75, 0x75, 0x69, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x30, 0x2d, 0x30, 0x30, 0x30, 0x33,
	}

	services := []Service{
		{"urn:Type1:device:controllee:1", "uuid:0000-0000-0000-0001", "", "http://127.0.0.1:8001"},
		{"urn:Type2:device:controllee:1", "uuid:0000-0000-0000-0002", "", "http://127.0.0.1:8002"},
		{"urn:Type3:device:controllee:1", "uuid:0000-0000-0000-0003", "", "http://127.0.0.1:8003"},
	}

	expect := []byte{RequestTypeAll, 1, 0}

	client, server := net.Pipe()
	defer server.Close()
	defer client.Close()

	reader := make(chan []byte)

	go func() {
		err := server.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Fatalf("could not set server read deadline: %v", err)
		}

		// Then read the encoded request
		buf := make([]byte, len(expect))
		_, err = server.Read(buf)
		if err != nil {
			t.Fatalf("server read error: %v", err)
		}

		_, err = server.Write(stream)
		if err != nil {
			t.Fatalf("server write error: %v", err)
		}

		reader <- buf
	}()

	c := Client{
		conn: client,
	}

	resp, err := c.GetServicesAll()
	if err != nil {
		t.Fatal(err)
	}

	out := <-reader

	if !reflect.DeepEqual(out, expect) {
		t.Fatalf("unexpected response from mock server: %#v", string(out))
	}

	if !reflect.DeepEqual(resp, services) {
		t.Fatal("unexpected services returned from mock server")
	}
}
