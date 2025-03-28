package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Address string `yaml:"address"`
}

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	DBUrl      string `yaml:"db_url" env:"DB_URL" env-required:"true"`
	HTTPServer `yaml:"http_server"`
}

func MustLoad() *Config {
	var configpath string

	configpath = os.Getenv("CONFIG_PATH")

	if configpath == "" {
		flags := flag.String("config", "", "Path to the configuration file")
		flag.Parse()

		configpath = *flags

		if configpath == ""  {
			log.Fatal("Config path is not set")
		}
	}

	if _, err := os.Stat(configpath); os.IsNotExist(err) {
		log.Fatal("Congif file does not exist on path:", configpath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configpath, &cfg); err != nil {
		log.Fatal("Cannot read config file:", err.Error())
	}

	return &cfg
}