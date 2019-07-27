package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NaySoftware/go-fcm"
)

type Config struct {
	Token     string
	ServerKey string
}

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

var conf Config

var sd [3]StationData
var gasd GasData

var stalastupdate [3]int64
var gaslastupdate int64

var AlarmBadAir bool = false
var AlarmSmoke bool = false

func GetHomeData() HomeData {
	CurTime := time.Now().Unix()
	if CurTime-stalastupdate[1] > 6 {
		sd[1] = StationData{-1, -1}
	}
	if CurTime-stalastupdate[1] > 6 {
		sd[2] = StationData{-1, -1}
	}
	if CurTime-gaslastupdate > 12 {
		gasd = GasData{-1, false}
	}
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
		stalastupdate[id] = time.Now().Unix()
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
		gaslastupdate = time.Now().Unix()
		w.WriteHeader(http.StatusOK)

		if gasd.PM25 > 2000 {
			if !AlarmBadAir {
				AlarmBadAir = true
				SendPush("Alert: Bad Air", "PM2.5: "+strconv.Itoa(gasd.PM25))
			} else {
				AlarmBadAir = false
			}
		}

		if gasd.Smoke {
			if !AlarmSmoke {
				AlarmSmoke = true
				SendPush("Alert: Smoke", "Smoke sensor triggered")
			} else {
				AlarmSmoke = false
			}
		}
	})

}

func SendPush(Title string, Body string) {
	var NP fcm.NotificationPayload
	NP.Title = Title
	NP.Body = Body

	data := map[string]string{}

	ids := []string{
		conf.Token,
	}

	c := fcm.NewFcmClient(conf.ServerKey)
	c.NewFcmRegIdsMsg(ids, data)
	c.SetNotificationPayload(&NP)
	status, err := c.Send()
	if err == nil {
		status.PrintResults()
	} else {
		fmt.Println(err)
	}
}

func LoadConfiguration(file string) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
}

func main() {
	LoadConfiguration("config.json")
	SetRoute()
	sd[1] = StationData{-1, -1}
	sd[2] = StationData{-1, -1}
	gasd = GasData{-1, false}
	stalastupdate[1] = time.Now().Unix()
	stalastupdate[2] = time.Now().Unix()
	gaslastupdate = time.Now().Unix()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
