package main

import (
	"fmt"
	"net"
	"flag"
	"net/http"
	"crypto/tls"
	"code.google.com/p/go.net/websocket"
	"github.com/huin/mqtt"
	"bufio"
	"bytes"
)

var (
	bs = flag.Bool("bs", false, "SSL/TLS broker connection")
	bhost = flag.String("bhost", "localhost", "Broker host")
	bport = flag.String("bport", "1883", "Broker port")
	bsinsec = flag.Bool("skip-verify", false, "Insecure Skip Verify")
	bcert = flag.String("bcert", "", "SSL/TLS broker certificate")
	bkey = flag.String("bkey", "", "SSL/TLS broker key")

	ws = flag.Bool("ws", false, "SSL/TLS websocket connection")
	wsport = flag.String("wport", "9999", "Websocket listening port")
	wscert = flag.String("wscert", "", "CA file for secure server")
	wskey = flag.String("wskey", "", "Key file for secure server")
)

func wshandler(ws *websocket.Conn) {
	flag.Parse()
	var mqcon net.Conn
	var err error
	if *bs {
		conf := tls.Config{InsecureSkipVerify: *bsinsec}
		if *bcert != "" && *bkey != "" {
			Cert, err := tls.LoadX509KeyPair(*bcert, *bkey)
			if err != nil {
				fmt.Println("LoadX509KeyPair:", err)
				return
			}
			conf.Certificates = []tls.Certificate{Cert}
		}
		mqcon, err = tls.Dial("tcp", *bhost + ":" + *bport, &conf)
	} else {
		mqcon, err = net.Dial("tcp", *bhost + ":" + *bport)
	}

	if err != nil {
		fmt.Println("mqcon error:", err.Error())
		ws.Close()
		return
	}
	ws.PayloadType = websocket.BinaryFrame

	bmqcon := bufio.NewReadWriter(bufio.NewReader(mqcon), bufio.NewWriter(mqcon))
	bws := bufio.NewReadWriter(bufio.NewReader(ws), bufio.NewWriter(ws))

	go func() {
		for {
			msg, err := mqtt.DecodeOneMessage(bmqcon, nil)
			fmt.Println("brok->", msg)
			if err != nil {
				mqcon.Close()
				return
			}
			wbuffer := new(bytes.Buffer)
			msg.Encode(wbuffer)
			bws.Write(wbuffer.Bytes())
			bws.Flush()
			wbuffer.Truncate(wbuffer.Len())
		}
	}()
	for {
		msg, err := mqtt.DecodeOneMessage(bws, nil)
		fmt.Println("webs->", msg)
		if err != nil {
			ws.Close()
			return
		}
		msg.Encode(bmqcon)
		bmqcon.Flush()
	}
}

func main() {
	flag.Parse()
	http.Handle("/mqtt", websocket.Handler(wshandler))
	var err error
	if *ws {
		if *wscert == "" || *wskey == "" {
			fmt.Println("-ws need a certificate and a key specified")
			return
		}
		err = http.ListenAndServeTLS(":" + *wsport, *wscert, *wskey, nil)
	} else {
		err = http.ListenAndServe(":" + *wsport, nil)
	}
	if err != nil {
		fmt.Println("ListenAndserve:", err)
		return
	}
}

