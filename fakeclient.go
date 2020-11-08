package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

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

	for {
		time.Sleep(time.Second * 3)

		log := handler.Log{
			Type:       logTypes[rand.Intn(3)],
			Msg:        "this is a message: " + handler.RandString(6),
			LogTime:    time.Now().Format(time.RFC3339),
			FileName:   "fakeclient.go",
			FuncName:   "SendFake",
			LineNumber: rand.Intn(1000),
			AppID:      id,
		}

		req, err := json.Marshal(log)
		if err != nil {
			ppt.Errorln("failed to marshal fake log:\n\t", err.Error())
			return
		}

		str := "http://127.0.0.1:3000/new/log?id=" + id

		resp, err = http.Post(str, "application/json", bytes.NewBuffer(req))
		if err != nil {
			ppt.Errorln("failed to create fake log:\n\t", err.Error())
			return
		}
		resp.Body.Close()
	}
}
