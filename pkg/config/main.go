package config

import (
	"github.com/spf13/viper"
	"log"
)

func Load(filename string) {
	cfg := viper.GetViper()

	cfg.SetConfigType("yaml")
	cfg.SetConfigFile(filename)

	err := cfg.ReadInConfig()
	if err != nil {
		log.Fatalf("Can't read config file: %s", err)
	}
	return
}

func Get(name string) interface{} {
	return viper.Get(name)
}

func GetString(name string) string {
	return viper.GetString(name)
}

func GetInt(name string) int {
	return viper.GetInt(name)
}
