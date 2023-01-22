package config

var (
	// pod info
	IsMaster = true
	PodID    = "mpereskokova-1" // kafka groupID
	BaseURL  = "localhost:8080"

	// postgres
	DatabaseURL = "postgres://user:password@localhost:5432/postgres"

	// redis
	RedisAddr   = "localhost:26379"
	RedisPrefix = "mpereskokova:"

	// kafka
	KafkaBrokers = []string{"localhost:9092"}
	ClicksTopic  = "mpereskokova-clicks"
	CreateTopic  = "mpereskokova-urls"
	ClicksSend   = int64(100)
)
