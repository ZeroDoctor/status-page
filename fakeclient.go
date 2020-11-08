package main

import "time"

func (wh *WebHandler) SendFake() {

	for {
		time.Sleep(time.Second * 3)
		wh.Broadcast <- WebMsg{
			Type: "log",
			Data: "another one",
		}
	}

}
