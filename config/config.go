package config

type ConfigServer struct {
	URL  string
	Port string
	Cors []string
}

type DBConfig struct {
	Type         string
	Path         string
	User         string
	Password     string
	Host         string
	Database     string
	ResetOnStart bool
}

type AppConfig struct {
	Server ConfigServer
	DB     DBConfig
}
