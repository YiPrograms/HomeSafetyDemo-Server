package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type StationData struct {
	Temp  int
	Humid int
}

type GasData struct {
	PM25  int
	Smoke bool
}

type HomeData struct {
	S1  StationData
	S2  StationData
	Gas GasData
}

var data HomeData

func SetRoute() {
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&data)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

}

func main() {
	SetRoute()
	data = HomeData{StationData{-1, -1}, StationData{-1, -1}, GasData{-1, false}}

	log.Fatal(http.ListenAndServe(":80", nil))
}
