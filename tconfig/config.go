package tconfig

type App struct {
}

type MySql struct {
	Username     string
	Password     string
	Address      string
	DatabaseName string
}

type Cors struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type HTTP struct {
	Port string
}

type MQTT struct {
	Broker   string
	Port     int
	ClientID string
	Username string
	Password string
}
