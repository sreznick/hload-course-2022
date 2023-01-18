package main

import (
    "database/sql"
    "fmt"
    "net/http"
	_ "net/http"
    "bytes"
    "encoding/json"
    "io/ioutil"
    "strconv"
	"time"
	"os/exec"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)

const SQL_DRIVER = "postgres"
const SQL_CONNECT_URL = "postgres://postgres:postgres@localhost:8082/db_name?sslmode=disable"
const PORT_LOCAL = ":8081"
const HTTP_LOCAL = "http://127.0.0.1" + PORT_LOCAL + "/"

const CREATE   = "INSERT INTO urls (url) VALUES ($1) ON CONFLICT DO NOTHING"
const GETBYVAL = "SELECT id FROM urls WHERE url = $1"
const GETBYID  = "SELECT url FROM urls WHERE id = $1"

const DOMAINSIZE = 62
const SPACESIZE  = 62 ^ 7

type Response struct {
	longurl string
	tinyurl string
}

type RequestBody struct {
	longurl string `json:"longurl"`
}

func handleCreate(context *gin.Context, db *sql.DB) {
    fmt.Println("stressTest starts.")
	body := RequestBody{}
	err := context.BindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"response": "bad long url"})
		return
	}

	_, err = db.Exec(CREATE, body.longurl)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"response": "can't insert new long url: " + err.Error()})
		return
	}

	var id int
	err = db.QueryRow(GETBYVAL, body.longurl).Scan(&id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"response": "can't retrieve just inserted url id: " + err.Error()})
		return
	}
	tinyUrl, err := idToUrl(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"response": "all short urls are used"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"lognurl": body.longurl, "tinyurl": tinyUrl})
}

func idToUrl(id int) (string, error) {
	if id >= SPACESIZE || id < 0 {
		return "", fmt.Errorf("ERROR: id isn't right")
	}
	bytes := make([]byte, 7)
	for i, _ := range bytes {
		bytes[i] = intToByte(id % DOMAINSIZE)
		id /= DOMAINSIZE
	}
	return string(bytes), nil
}

func intToByte(id int) byte {
	if id < 10 {
		return byte('0' + id)
	} else if id < 36 {
		return byte('a' + (id - 10))
	} else {
		return byte('A' + (id - 36))
	}
}

func handleUrlGet(context *gin.Context, db *sql.DB) {
	shortUrl := context.Params.ByName("url")
	id, err := bytesToInt([]byte(shortUrl))
	if err != nil {
		context.Writer.WriteHeader(404)
		return
	}

	var longUrl string
	err = db.QueryRow(GETBYID, id).Scan(&longUrl)
	if err != nil {
		context.Writer.WriteHeader(404)
		return
	}

	context.Redirect(302, longUrl)
}

func bytesToInt(bytes []byte) (int, error) {
	if len(bytes) != 7 {
		return 0, fmt.Errorf("ERROR: wrong tinyurl length!")
	}
	result := 0
	for i := len(bytes) - 1; i >= 0; i-- {
		result *= DOMAINSIZE
		id, err := byteToInt(bytes[i])
		if err != nil {
			return 0, err
		}
		result += id
	}
	return result, nil
}

func byteToInt(symbol byte) (int, error) {
	if '0' <= symbol && symbol <= '9' {
		return int(symbol-'0'), nil
	}
	if 'a' <= symbol && symbol <= 'z' {
		return int(symbol - 'a') + 10, nil
	}
	if 'A' <= symbol && symbol <= 'Z' {
		return int(symbol-'A') + 36, nil
	}
	return 0, fmt.Errorf("ERROR: Fobiden symbol!")
}

var (
	createRequestCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_cnt",
		Help: "number of create",
	})
	getRequestCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_cnt",
		Help: "number of get",
	})

	createTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "create_time",
		Help: "time of create",
	})
	getTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_time",
		Help: "time of get",
	})
)


func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		createRequestCnt.Inc()
		start := time.Now()
		handleCreate(c, db)
		createTime.Observe(float64(time.Since(start).Milliseconds()))
	})

	r.GET("/:url", func(c *gin.Context) {
		getRequestCnt.Inc()
		start := time.Now()
		handleUrlGet(c, db)
		getTime.Observe(float64(time.Since(start).Milliseconds()))
	})

	return r
}

func openSQL() *sql.DB {
	cmd := exec.Command(`psql -U postgres -tc "DROP DATABASE IF EXISTS db_name; CREATE DATABASE db_name;"`)
    fmt.Println(cmd.Output())
	conn, err := sql.Open(SQL_DRIVER, SQL_CONNECT_URL)
    if err != nil {
        fmt.Println("ERROR: Failed to open sql file! ", err)
        panic("exit")
    }

    err = conn.Ping()
    if err != nil {
        fmt.Println("ERROR: Failed to ping database! ", err)
        panic("exit")
    }

    _, err = conn.Exec("CREATE TABLE IF NOT EXISTS urls (id SERIAL, url TEXT PRIMARY KEY)")
	if err != nil {
		fmt.Println("ERROR: Failed to open or create table! ", err)
		panic("exit")
	}

    return conn
}

func testCreateRequest(longurl string) Response {
	request, err := http.NewRequest("PUT",
            HTTP_LOCAL + "create",
            bytes.NewBuffer([]byte(`{"json": "` + longurl + `"}`))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var ans Response
	json.Unmarshal(body, &ans)

	return ans
}

func testGoodGetRequest(url string) {
	resp, err := http.Get(HTTP_LOCAL + url)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("ASSERT FAIL: Expected code 200. Got code " + strconv.Itoa(resp.StatusCode)))
	}
}

func testWrongGetRequest() {
	resp, err := http.Get(HTTP_LOCAL + "some_trash")
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 404 {
		panic(fmt.Errorf("ASSERT FAIL: Expected code 404. Got code " + strconv.Itoa(resp.StatusCode)))
	}
}

func stressTest() {
    fmt.Println("stressTest starts.")
	rightAnswer := testCreateRequest("https://www.youtube.com/watch?v=eBGIQ7ZuuiU")
	for i := 0; i <= 10000; i++ {
		newAnswer := testCreateRequest("https://www.youtube.com/watch?v=eBGIQ7ZuuiU")
		if newAnswer != rightAnswer {
			panic(fmt.Errorf("ASSERT FAIL: Answer changed!"))
		}
	}
    fmt.Println("Test 1 is ok.")
	for i := 0; i <= 100000; i++ {
		testGoodGetRequest(rightAnswer.tinyurl)
	}
    fmt.Println("Test 2 is ok.")
	for i := 0; i <= 100000; i++ {
		testWrongGetRequest()
	}
    fmt.Println("Test 3 is ok.")
}

func startServer() {
	db := openSQL()
	r := setupRouter(db)
	r.Run(PORT_LOCAL)
}

func main() {
	go startServer()
	time.Sleep(10 * time.Second);
    stressTest()
}
