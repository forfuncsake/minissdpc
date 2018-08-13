package minissdpd

import (
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	MockRequestTypeString = iota
	MockRequestTypeGet
	MockRequestTypeRegister

	testSocket = "test.sock"
)

type mockServer struct {
	conn   net.Listener
	dir    string
	dialog map[string]string

	done chan struct{}
}

// Close will close the mock server.
// Only call close once per mockServer
func (m *mockServer) Close() error {
	defer os.RemoveAll(m.dir)
	close(m.done)
	return m.conn.Close()
}

func newMockServer(t *testing.T, dialog map[string]string) *mockServer {
	dir, err := ioutil.TempDir("", "ssdpc")
	if err != nil {
		t.Fatalf("could not create temp socket dir: %v", err)
	}

	sock := filepath.Join(dir, testSocket)
	conn, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("could not open test socket: %v", err)
	}

	t.Logf("Creating mock server socket at %s\n", sock)

	server := &mockServer{
		conn:   conn,
		dir:    dir,
		dialog: dialog,
		done:   make(chan struct{}),
	}
	go server.handle(t)

	return server
}

func (m *mockServer) handle(t *testing.T) {
	for {
		select {
		case <-m.done:
			break
		default:
		}

		conn, err := m.conn.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			t.Fatalf("mock server could not accept connection: %v", err)
		}

		t.Log("mock server accepted connection")

		// The first byte is the request type
		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			t.Fatalf("could not read request type from start of request: %v", err)
		}
		reqType := int(buf[0])

		switch reqType {
		case MockRequestTypeString:
			n, err := DecodeStringLength(conn)
			if err != nil {
				t.Fatalf("mock server could not decode string length: %v", err)
			}
			buf = make([]byte, n)
			_, err = conn.Read(buf)
			if err != nil {
				t.Fatalf("could not read request string: %v", err)
			}
			t.Logf("mock server read %d bytes\n", n)

			if resp, ok := m.dialog[string(buf)]; ok {
				_, err := conn.Write([]byte(resp))
				if err != nil {
					t.Fatalf("mock server could not write response: %v", err)
				}
			}
		case MockRequestTypeGet:
		case MockRequestTypeRegister:
			// Register doesn't send a response
			time.Sleep(1 * time.Second)
		}

		conn.Close()
	}
}

func TestClientConnections(t *testing.T) {
	s := newMockServer(t, nil)
	defer s.Close()

	c := Client{
		SocketPath: filepath.Join(s.dir, testSocket),
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

	dialog := map[string]string{
		testString: testString,
	}

	s := newMockServer(t, dialog)
	defer s.Close()

	c := Client{
		SocketPath: filepath.Join(s.dir, testSocket),
	}

	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	_, err := c.Write([]byte{MockRequestTypeString})
	if err != nil {
		t.Fatalf("could not write mock request type byte: %v", err)
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

	buf := make([]byte, len(testString))
	nn, err := c.conn.Read(buf)
	if err != nil {
		t.Fatalf("error reading from mock server: %v", err)
	}

	if int(nn) != len(testString) {
		t.Fatalf("expected to read %d bytes, got %d", len(testString), nn)
	}

	if string(buf) != testString {
		t.Fatalf("unexpected response from mock server: %#v", string(buf))
	}
}

func TestClientRegister(t *testing.T) {
	s := newMockServer(t, nil)
	defer s.Close()

	c := &Client{
		SocketPath: filepath.Join(s.dir, testSocket),
	}

	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	_, err := c.Write([]byte{MockRequestTypeRegister})
	if err != nil {
		t.Fatalf("could not write mock request type byte: %v", err)
	}

	err = c.RegisterService(Service{
		Type:     "urn:Dummy:device:controllee:1",
		USN:      "1234-1234-1234-1234",
		Server:   "Dummy 1.0",
		Location: "http://127.0.0.1/setup.xml",
	})
	if err != nil {
		t.Fatal(err)
	}
}
