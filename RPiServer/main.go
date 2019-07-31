package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	API_KEY string
	APP_ID  string
}

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

var conf Config

var sd [3]StationData
var gasd GasData

var connected [3]bool

var AlarmBadAir bool = false
var AlarmSmoke bool = false

var AlertsHist []Alert

func GetHomeData() HomeData {
	return HomeData{sd[1], sd[2], gasd, AlertsHist}
}

func SendAirData(c *websocket.Conn, id int, AirUpdate chan int) {
	for {
		msg, err := json.Marshal(gasd)
		fmt.Printf("Send %d: %s\n", id, msg)
		err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Write to Station:", err)
			break
		}
		<-AirUpdate
	}
}

func SetRoute(HaveUpdate chan int) {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	AirUpdate := make(chan int)

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
		id, _ := strconv.Atoi(req.URL.Query().Get("id"))
		if connected[id] {
			fmt.Println("Station", id, "Already connected!")
			return
		}
		fmt.Println("Station", id, "Connected!")
		connected[id] = true
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			fmt.Println("Upgrade:", err)
			return
		}
		defer func() {
			fmt.Println("Station", id, "Disconnected!")
			connected[id] = false
			sd[id] = StationData{-1, -1}
			c.Close()
		}()

		go SendAirData(c, id, AirUpdate)

		for {
			c.SetReadDeadline(time.Now().Add(time.Second * 5))
			_, msg, errr := c.ReadMessage()
			if errr != nil {
				fmt.Println("Read:", err)
				return
			}
			fmt.Printf("Receive %d: %s\n", id, msg)

			var dat StationData
			err := json.Unmarshal(msg, &dat)
			if err != nil {
				fmt.Println(err)
				return
			}

			sd[id] = dat
			HaveUpdate <- 1
		}
	})

	http.HandleFunc("/airupdate", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Air Connected!")
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			fmt.Println("Upgrade:", err)
			return
		}
		defer func() {
			fmt.Println("Air Disconnected!")
			c.Close()
		}()

		for {
			c.SetReadDeadline(time.Now().Add(time.Second * 12))
			_, msg, errr := c.ReadMessage()
			if errr != nil {
				fmt.Println("Read:", errr)
				return
			}
			fmt.Printf("Receive Air: %s\n", msg)

			var dat GasData
			err := json.Unmarshal(msg, &dat)
			if err != nil {
				fmt.Println(err)
				return
			}
			gasd = dat

			if connected[1] {
				AirUpdate <- 1
			}
			if connected[2] {
				AirUpdate <- 2
			}

			if gasd.PM25 > 2000 {
				if !AlarmBadAir {
					AlarmBadAir = true
					SendPush("Alert: Bad Air", "PM2.5: "+strconv.Itoa(gasd.PM25))
					AlertsHist = append(AlertsHist, Alert{"Bad Air", "PM2.5: " + strconv.Itoa(gasd.PM25), time.Now().Unix()})
				}
			} else {
				AlarmBadAir = false
			}

			if gasd.Smoke {
				if !AlarmSmoke {
					AlarmSmoke = true
					SendPush("Alert: Smoke", "Smoke sensor triggered")
					AlertsHist = append(AlertsHist, Alert{"Smoke", "Smoke sensor triggered", time.Now().Unix()})
				}
			} else {
				AlarmSmoke = false
			}

			HaveUpdate <- 1
		}
	})

}

func SendPush(Title string, Body string) {
	url := "https://onesignal.com/api/v1/notifications"
	var jsonStr = []byte(`{
		"app_id": "` + conf.APP_ID + `",
		"included_segments": ["Active Users"],
		"contents": {"en": "` + Body + `"},
		"headings": {"en": "` + Title + `"}
	}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Basic "+conf.API_KEY)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("Push HTTP Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Push Response:", string(body))
}

func ConnectToRouter(HaveUpdate chan int) {
	for {
		SendToRouter(HaveUpdate)
		time.Sleep(2 * time.Second)
	}
}

func SendToRouter(HaveUpdate chan int) {
	url := "ws://homesafetydemo.ml:8080/update"

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial:", err)
		return
	}
	defer func() {
		fmt.Printf("Disconnected to %s", url)
		c.Close()
	}()
	fmt.Printf("Connected to %s", url)

	for {
		fmt.Println("Send Update")

		dat := GetHomeData()
		msg, _ := json.Marshal(dat)

		err := c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Write to Station:", err)
			break
		}
		<-HaveUpdate
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
	HaveUpdate := make(chan int)
	SetRoute(HaveUpdate)
	sd[1] = StationData{-1, -1}
	sd[2] = StationData{-1, -1}
	gasd = GasData{-1, false}
	AlertsHist = append(AlertsHist, Alert{"Earthquake", "Scale: 7.6", 937849636})
	go ConnectToRouter(HaveUpdate)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
