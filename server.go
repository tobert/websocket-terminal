package main

import (
	"encoding/base64"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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

type wsPty struct {
	Cmd *exec.Cmd // pty builds on os.exec
	Pty *os.File  // a pty is simply an os.File
}

func (wp *wsPty) Start() {
	var err error
	wp.Cmd = exec.Command("/bin/bash", "--login")
	wp.Pty, err = pty.Start(wp.Cmd)
	if err != nil {
		log.Fatalf("Failed to start command: %s\n", err)
	}
}

func (wp *wsPty) Stop() {
	wp.Pty.Close()
	wp.Cmd.Wait()
}

func terminalHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Websocket upgrade failed: %s\n", err)
	}
	defer conn.Close()

	wp := wsPty{}
	// TODO: check for errors, return 500 on fail
	wp.Start()

	go func() {
		buf := make([]byte, 128)
		for {
			n, err := wp.Pty.Read(buf)
			if err != nil {
				log.Printf("Failed to read from pty master: %s", err)
				return
			}

			out := make([]byte, base64.StdEncoding.EncodedLen(n))
			base64.StdEncoding.Encode(out, buf[0:n])
			log.Printf("forwarding %d bytes of data as %d bytes of base64\n", len(buf), len(out))

			err = conn.WriteMessage(websocket.TextMessage, out)

			if err != nil {
				log.Printf("Failed to send %d bytes on websocket: %s", n, err)
				return
			}
		}
	}()

	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.Printf("conn.ReadMessage failed: %s\n", err)
				return
			}
		}

		log.Printf("Received: '%s'\n", string(payload))

		switch mt {
		case websocket.BinaryMessage:
			log.Printf("Ignoring binary message...\n")
		case websocket.TextMessage:
			log.Printf("Base64: %s\n", payload)
			buf := make([]byte, base64.StdEncoding.DecodedLen(len(payload)))
			n, err := base64.StdEncoding.Decode(buf, payload)
			log.Printf("Decoded %n bytes msg: %s\n", n, err)
			log.Printf("Decode: %s\n", buf)
			wp.Pty.Write(buf)
		default:
			log.Printf("Invalid message type %d\n", mt)
			return
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
