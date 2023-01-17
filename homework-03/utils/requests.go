package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ClientBD urlData

var ClientKafka *Kafka

var ClientRedis *Redis

type urlData interface {
	store(key string, url string)
	loadKey(url string) (key string, ok bool)
	loadURL(key string) (url string, ok bool)
}

var lastMap map[string]string

type GetURL struct {
	URL string `json:"longurl"`
}

type PingResponse struct {
	Topic string `json:"topic"`
	Type  string `json:"type"`
}

func GetURLFromKey(key string) (string, bool) {
	url, ok := ClientBD.loadURL(key)
	if !ok {
		return "", ok
	}

	return url, ok
}

func incrementKey(key string) error {
	result, err := ClientRedis.GetMapWithKey("metrics")

	if err != nil {
		return err
	}

	shouldSend := false

	val, ok := result[key]

	if ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return err
		}

		if intVal%100 == 0 {
			shouldSend = true
		}

		val = strconv.Itoa(intVal + 1)
	} else {
		val = "1"
	}

	err = ClientRedis.PutWithKey("metrics", key, val)

	if shouldSend {
		err = sendMetricUpdate(key, val)
	}

	return err
}

func sendMetricUpdate(key string, val string) error {
	kafkaClient := Kafka{
		Topic: "bmadzhuga-metrics",
		Type:  "master",
	}

	err := kafkaClient.Connect()

	if err != nil {
		return err
	}

	err = kafkaClient.Send(key, val, true)

	if err != nil {
		return err
	}

	return nil
}
func askMasterForURL(key string) (string, bool) {
	if lastMap != nil {
		val, ok := lastMap[key]

		if ok {
			return val, true
		}
	}

	kafkaMaster := Kafka{
		Topic: "bmadzhuga-events",
		Type:  "master",
	}

	err := kafkaMaster.Connect()

	if err != nil {
		return "", false
	}

	err = kafkaMaster.Send(ClientKafka.Topic, key, true)

	if err != nil {
		return "", false
	}

	url, ok := getURLFromRedis(key, 100)

	return url, ok
}

func getURLFromRedis(key string, timeout int) (string, bool) {
	run := true
	counter := 0
	for run {
		result, err := ClientRedis.GetMap()
		if err != nil {
			return "", false
		}

		lastMap = result

		val, ok := result[key]

		if ok {
			return val, true
		}

		counter++

		time.Sleep(time.Millisecond * 100)

		if counter >= timeout {
			run = false
		}
	}
	return "", false
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	start := time.Now()

	promReceivedLinkCount.Inc()

	key := r.URL.Path[1:]

	var url string
	var ok bool

	if ClientKafka.Type == "master" {
		url, ok = GetURLFromKey(key)
	} else {
		url, ok = askMasterForURL(key)
		if ok {
			incrementKey(key)
		}
	}

	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	http.Redirect(w, r, "https://"+url, 301)

	duration := time.Since(start)

	requestProcessingTimeSummaryMs.Observe(duration.Seconds())
	requestProcessingTimeHistogramMs.Observe(duration.Seconds())

	PrometheusPush()
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResp := PingResponse{Topic: ClientKafka.Topic, Type: ClientKafka.Type}
	resp, err := json.Marshal(jsonResp)

	if err != nil {
		http.Error(w, "JSON is invalid", 400)
		return
	}

	w.Write(resp)
}

func createResp(w http.ResponseWriter, key string, url string) {
	jsonResp := UrlDB{Key: key, URL: url}
	resp, err := json.Marshal(jsonResp)
	if err != nil {
		http.Error(w, "JSON is invalid", 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func HandlePut(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		return
	}

	fmt.Printf("PUT")

	start := time.Now()

	promRegisteredLinkCount.Inc()

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var jsonURL GetURL
	decoder.Decode(&jsonURL)

	url := jsonURL.URL

	if url == "" {
		w.WriteHeader(http.StatusOK)
		duration := time.Since(start)

		requestProcessingTimeSummaryMs.Observe(duration.Seconds())
		requestProcessingTimeHistogramMs.Observe(duration.Seconds())

		PrometheusPush()

		return
	}

	if key, ok := ClientBD.loadKey(url); ok {
		createResp(w, key, url)

		duration := time.Since(start)

		requestProcessingTimeSummaryMs.Observe(duration.Seconds())
		requestProcessingTimeHistogramMs.Observe(duration.Seconds())

		PrometheusPush()

		return
	}

	key, ok := "", true
	for ok {
		key = RandKey()
		_, ok = ClientBD.loadURL(key)
	}
	ClientBD.store(key, url)
	createResp(w, key, url)

	duration := time.Since(start)

	requestProcessingTimeSummaryMs.Observe(duration.Seconds())
	requestProcessingTimeHistogramMs.Observe(duration.Seconds())

	PrometheusPush()
}
