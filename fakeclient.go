package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zerodoctor/go-status/handler"
	ppt "github.com/zerodoctor/goprettyprinter"
)

var logTypes = []string{
	"info", "warn", "error",
}

// SendFake :
func SendFake(wh *handler.WebHandler) {

	time.Sleep(time.Second * 3)
	resp, err := http.Post("http://127.0.0.1:3000/new/app?name=fakeClient", "", nil)
	if err != nil {
		ppt.Errorln("failed to create fake client program:\n\t", err.Error())
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	id := string(body)

	resp.Body.Close()

	socket, err := createClient("127.0.0.1:3000", "/ws")
	if err != nil {
		ppt.Errorln("failed to create client:\n\t", err.Error())
		return
	}

	msg := handler.WebMsg{
		Type: "client",
		Data: handler.RandString(12),
	}
	err = socket.WriteJSON(msg)
	if err != nil {
		ppt.Errorln("failed to send msg:\n\t", err)
		return
	}

	count := 0

	for {
		time.Sleep(time.Second * 3)

		logs := []handler.Log{
			{
				Type:       logTypes[rand.Intn(3)],
				Msg:        "this is a message: " + handler.RandString(6),
				LogTime:    time.Now().Format(time.RFC3339),
				FileName:   "fakeclient.go",
				FuncName:   "SendFake",
				LineNumber: rand.Intn(1000),
				AppID:      id,
				AppName:    "fakeClient",
				Index:      count,
			},
		}
		count++

		msg = handler.WebMsg{
			Type: "logs",
			Logs: logs,
		}

		err = socket.WriteJSON(msg)
		if err != nil {
			ppt.Errorln("failed to send logs:\n\t", err)
		}
	}
}

func createClient(address, path string) (*websocket.Conn, error) {

	u := url.URL{
		Scheme: "ws",
		Host:   address,
		Path:   path,
	}

	socket, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dail %s: %s - %+v", u.String(), err.Error(), resp)
	}

	return socket, nil
}
