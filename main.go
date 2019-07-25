
package main

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"
	"strconv"
)

type StationData struct {
	Temp int
	Humid int
}

type GasData struct {
	
}

type HomeData struct {
	S1 StationData
	S2 StationData
	Gas GasData
}


var sd [3]StationData
var gasd GasData

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
		fmt.Println(b)
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
		id , _:= strconv.Atoi(req.URL.Query().Get("id"))
		sd[id] = dat
	})

}

func main() {
	SetRoute();
	sd[1] = StationData{100,50}
	sd[2] = StationData{1000,550}
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
