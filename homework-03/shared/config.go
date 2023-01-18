package shared

type Topics struct {
  Tinurls    string `yaml:"tinurls"`
  Clicks     string `yaml:"clicks"`
}

type MasterConfig struct {
  Kafka      string `yaml:"kafka"`
  Postgres   string `yaml:"postgres"`
  Socket     string `yaml:"socket"`
  Topics     Topics `yaml:"topics"`
}

type WorkerConfig struct {
  Kafka      string `yaml:"kafka"`
  Postgres   string `yaml:"postgres"`
  Redis      string `yaml:"redis"`
  Socket     string `yaml:"socket"`
  Topics     Topics `yaml:"topics"`
}

type Config struct {
  Master     MasterConfig `yaml:"master"`
  Worker     WorkerConfig `yaml:"worker"`
}
