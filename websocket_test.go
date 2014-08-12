package turnpike

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

func newWebsocketServer(t *testing.T) (int, Router, io.Closer) {
	r := NewDefaultRouter()
	r.RegisterRealm(testRealm, NewDefaultRealm())
	s := NewWebsocketServer(r)
	s.RegisterProtocol(jsonWebsocketProtocol, websocket.TextMessage, new(JSONSerializer))
	s.RegisterProtocol(msgpackWebsocketProtocol, websocket.BinaryMessage, new(MessagePackSerializer))
	server := &http.Server{
		Handler: s,
	}

	var addr net.TCPAddr
	l, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(l)
	return l.Addr().(*net.TCPAddr).Port, r, l
}

func TestWSHandshakeJSON(t *testing.T) {
	port, r, closer := newWebsocketServer(t)
	defer closer.Close()

	client, err := NewWebsocketPeer(JSON, fmt.Sprintf("ws://localhost:%d/", port), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}

	client.Send(&Hello{Realm: testRealm})
	go r.Accept(client)

	if msg, ok := <-client.Receive(); !ok {
		t.Fatal("Receive buffer closed")
	} else if _, ok := msg.(*Welcome); !ok {
		t.Errorf("Message not Welcome message: %T, %+v", msg, msg)
	}
}

func TestWSHandshakeMsgpack(t *testing.T) {
	port, r, closer := newWebsocketServer(t)
	defer closer.Close()

	client, err := NewWebsocketPeer(MSGPACK, fmt.Sprintf("ws://localhost:%d/", port), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}

	client.Send(&Hello{Realm: testRealm})
	go r.Accept(client)

	if msg, ok := <-client.Receive(); !ok {
		t.Fatal("Receive buffer closed")
	} else if _, ok := msg.(*Welcome); !ok {
		t.Errorf("Message not Welcome message: %T, %+v", msg, msg)
	}
}