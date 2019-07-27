package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type StationData struct {
	Temp  int
	Humid int
}

type GasData struct {
	PM25 int
}

type HomeData struct {
	S1  StationData
	S2  StationData
	Gas GasData
}

var sd [3]StationData
var gasd GasData

var online [3]bool

func GetHomeData() HomeData {
	fmt.Println("get")
	return HomeData{S1: sd[1], S2: sd[2], Gas: gasd}
}

func SetRoute() {
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		dat := GetHomeData()
		b, err := json.Marshal(dat)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})

	http.HandleFunc("/stationupdate", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var dat StationData
		err := decoder.Decode(&dat)
		if err != nil {
			fmt.Println(err)
			return
		}
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		sd[id] = dat
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/airupdate", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var dat GasData
		err := decoder.Decode(&dat)
		if err != nil {
			fmt.Println(err)
			return
		}
		gasd = dat
		w.WriteHeader(http.StatusOK)
	})

}

func main() {
	SetRoute()
	sd[1] = StationData{-1, -1}
	sd[2] = StationData{-1, -1}
	gasd = GasData{-1}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
