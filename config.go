package tofu

type Config struct {
}

type MySqlConfig struct {
	Username     string
	Password     string
	Address      string
	DatabaseName string
}

type HTTPConfig struct {
	Port string
}

type MQTTConfig struct {
	Broker   string
	Port     int
	ClientID string
	Username string
	Password string
}
