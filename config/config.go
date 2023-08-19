package config

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

const STORE string = "$HOME/.local/share/tracker"

var Config *viper.Viper

type Server struct {
	HostName string
	Port     string
}

func (server *Server) URL() string {
	return fmt.Sprintf("%s:%s", server.HostName, server.Port)
}

func Init() error {
	Config = viper.New()
	Config.SetDefault("hostname", "localhost")
	Config.SetDefault("port", "8080")
	Config.SetDefault("store.dir", os.ExpandEnv(STORE))
	Config.SetDefault("store.runs", "runs")
	Config.SetDefault("store.db", "runs.db")

	Config.SetConfigFile("tracker.yaml")
	Config.AddConfigPath(os.ExpandEnv("$HOME/.config"))
	Config.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}

		if err := viper.SafeWriteConfig(); err != nil {
			log.Println("Error writing config", err)
		}

		log.Println("Config file not found, using default values")
	}

	return nil
}

func GetDBPath() string {
	return path.Join(Config.GetString("store.dir"), Config.GetString("store.db"))
}

func GetDataPath() string {
	return path.Join(Config.GetString("store.dir"), Config.GetString("store.runs"))
}
