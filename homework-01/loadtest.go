package main

import (
    "fmt"
    "bytes"
    "math/rand"
    "sync"
    "net/http"
    "log"
     "encoding/json"
    "github.com/brianvoe/gofakeit/v6"
)


type callable func() *http.Response

// Max simultaneous PG connections, also max runnable at the same time goroutines.
const MAX_PG_CONNECTIONS_COUNT = 5

// Target URIs.
const TINYURL_CREATION_URI = "http://localhost:8080/create"
const LONGURL_FETCH_URI_PREFIX = "http://localhost:8080/"

// Benchmark limits.
const HTTP_PUT_REQUESTS_COUNT = 10_000
const HTTP_GET_SUCCESS_COUNT = 100_000
const HTTP_GET_FAILURE_COUNT = 100_000

var data = make([]string, 0, HTTP_PUT_REQUESTS_COUNT)
var dataMutex = &sync.Mutex{}

var responseStatuses = make(map[int]int)
var responseStatusesMutex = &sync.Mutex{}

func main() {
  runRequests("PUT", HTTP_PUT_REQUESTS_COUNT, performTinyURLCreationRequest)
  runRequests("GET SUCCESS", HTTP_GET_SUCCESS_COUNT, performSuccessfulFetchRequests)
  runRequests("GET FAILURE", HTTP_GET_FAILURE_COUNT, performFailureFetchRequests)

  fmt.Println(responseStatuses)
}

func runRequests(task string, iterations int, f callable) {
  var waitGroup sync.WaitGroup
  waitGroup.Add(iterations)
  semaphore := make(chan int, MAX_PG_CONNECTIONS_COUNT)

  for i := 0; i < iterations; i++ {
    semaphore <- 1
    go func() {
      defer waitGroup.Done()
      if response := f(); response != nil {
        responseStatusesMutex.Lock()
        responseStatuses[response.StatusCode] += 1
        responseStatusesMutex.Unlock()
      }
      <- semaphore
    }()
  }

  waitGroup.Wait()
  log.Print(fmt.Sprintf("%s: OK", task))
}

type Result struct {
    Longurl string `json:"longurl"`
    Tinyurl string `json:"tinyurl"`
}

func performTinyURLCreationRequest() *http.Response {
  var requestBody = []byte(fmt.Sprintf(`{"longurl": "%s"}`, gofakeit.LetterN(100)))
  request, err := http.NewRequest(http.MethodPut, TINYURL_CREATION_URI, bytes.NewBuffer(requestBody))
  if err != nil {
    fmt.Println("requesterr", err, request)
  }

  client := &http.Client{}
  response, err2 := client.Do(request)

  var result Result
  err = json.NewDecoder(response.Body).Decode(&result)
  response.Body.Close()

  if err2 == nil && response.StatusCode == http.StatusOK {
    dataMutex.Lock()
    data = append(data, result.Tinyurl)
    dataMutex.Unlock()
  }

  defer response.Body.Close()

  return response
}

func performSuccessfulFetchRequests() *http.Response {
  getClient := &http.Client{
      CheckRedirect: func(req *http.Request, via []*http.Request) error {
          return http.ErrUseLastResponse
      },
  }

  var request *http.Request
  var response *http.Response
  var err error

  r := rand.Int() % len(data)
  tinyurl := data[r]

  if request, err = http.NewRequest("GET", LONGURL_FETCH_URI_PREFIX + tinyurl, nil); err != nil {
    log.Print(err, request)
    return nil
  }

  if response, err = getClient.Do(request); err != nil {
    log.Print("HTTP error: ", err, response)
  }

  defer response.Body.Close()

  return response
}

func performFailureFetchRequests() *http.Response {
  var request *http.Request
  var err error

  getClient := &http.Client{
      CheckRedirect: func(req *http.Request, via []*http.Request) error {
          return http.ErrUseLastResponse
      },
  }

  if request, err = http.NewRequest("GET", LONGURL_FETCH_URI_PREFIX + gofakeit.LetterN(100), nil); err != nil {
    log.Print(err, request)
    return nil
  }

  var response *http.Response

  if response, err = getClient.Do(request); err != nil {
    log.Print("HTTP error: ", err, response)
  }

  defer response.Body.Close()

  return response
}
