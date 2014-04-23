package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	addrFlag string
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1,
	WriteBufferSize: 1,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func terminalHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Websocket upgrade failed: %s\n", err)
	}
	defer conn.Close()

	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.Printf("conn.ReadMessage failed: %s\n", err)
				return
			}
		}

		log.Printf("Bytes: %s\n", string(payload))

		switch mt {
		case websocket.BinaryMessage:
		case websocket.TextMessage:
		case websocket.PingMessage:
		case websocket.PongMessage:
		default:
			log.Fatalf("Invalid message type %d\n", mt)
		}
	}
}

func init() {
	addrFlag = ":9000"
}

func main() {
	cwd, _ := os.Getwd() // TODO: flag

	// websocket that will wrap a pty
	http.HandleFunc("/terminal", terminalHandler)

	// serve html & javascript
	http.Handle("/", http.FileServer(http.Dir(cwd)))

	err := http.ListenAndServe(addrFlag, nil)
	if err != nil {
		log.Fatalf("net.http could not listen on address '%s': %s\n", addrFlag, err)
	}
}
