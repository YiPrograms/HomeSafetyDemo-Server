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

type Alert struct {
	Title string
	Body  string
}

type HomeData struct {
	S1     StationData
	S2     StationData
	Gas    GasData
	Alerts []Alert
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

		if req.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var dat HomeData
		err := json.NewDecoder(req.Body).Decode(&dat)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		fmt.Println("Get Update")
		data = dat
	})

}

func main() {
	SetRoute()
	data = HomeData{StationData{-1, -1}, StationData{-1, -1}, GasData{-1, false}, nil}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
