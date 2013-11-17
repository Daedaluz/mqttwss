package main

import (
	"fmt"
	"net"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"github.com/huin/mqtt"
)

func wshandler(ws *websocket.Conn) {
	mqcon, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		fmt.Println("mqcon error:", err.Error())
		ws.Close()
		return
	}
	ws.PayloadType = websocket.BinaryFrame
	go func() {
		for {
			msg, err := mqtt.DecodeOneMessage(mqcon, nil)
			if err != nil {
				mqcon.Close()
				return
			}
			msg.Encode(ws)
		}
	}()
	for {
		msg, err := mqtt.DecodeOneMessage(ws, nil)
		if err != nil {
			ws.Close()
			return
		}
		msg.Encode(mqcon)
	}
}

func main() {
	http.Handle("/mqtt", websocket.Handler(wshandler))
	http.ListenAndServe(":8080", nil)
}

