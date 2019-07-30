package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type StationData struct {
	Temp  int
	Humid int
}

type GasData struct {
	PM25  int
	Smoke bool
}

type Alert struct {
	Title string
	Body  string
	Time  int64
}

type HomeData struct {
	S1     StationData
	S2     StationData
	Gas    GasData
	Alerts []Alert
}

var data HomeData

var int ClientCount = 0

func SetRoute() {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	HaveUpdate := make(chan int)

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Client Connected!")
		ClientCount++
		fmt.Println("Client Count:", ClientCount)
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			fmt.Println("Upgrade:", err)
			return
		}
		defer func() {
			fmt.Println("Client Disconnected!")
			ClientCount--
			c.Close()
		}()

		for {
			msg, err := json.Marshal(data)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Send: %s\n", msg)
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Write to Client:", err)
				return
			}
			<-HaveUpdate
		}

	})

	http.HandleFunc("/update", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("RPi Connected!")
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			fmt.Println("Upgrade:", err)
			return
		}
		defer func() {
			fmt.Println("RPi Disconnected!")
			c.Close()
		}()

		for {
			c.SetReadDeadline(time.Now().Add(time.Second * 20))
			_, msg, errr := c.ReadMessage()

			var dat HomeData
			err := json.Unmarshal(msg, &dat)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			fmt.Println("Get Update")
			data = dat
			for i := 0; i < ClientCount; i++ {
				HaveUpdate <- 1
			}
		}

	})

}

func main() {
	SetRoute()
	data = HomeData{StationData{-1, -1}, StationData{-1, -1}, GasData{-1, false}, nil}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
