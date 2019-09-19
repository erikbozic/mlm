package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const (
	configDirName  = "mesos-monitor"
	configFileName = "mesos-monitor.json"
)

func init() {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		configDirPath = path.Join(v, configDirName)
	} else if v := os.Getenv("HOME"); v != "" {
		configDirPath = path.Join(v, ".config", configDirName)
	}
	configFilePath = path.Join(configDirPath, configFileName)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		err := os.MkdirAll(configDirPath, 0744)
		if err != nil {
			log.Println("error creating config dir:", err.Error())
		}

		f, err := os.Create(configFilePath)
		if err != nil {
			log.Println("error creating config file:", err.Error())
		}
		defer f.Close()
	}
}

var (
	configFilePath string
	configDirPath  string
)

type MonitorConfig struct {
	MesosMasterUrl string `json:"mesosMasterUrl"`
}

func ReadConfig() (cfg MonitorConfig) {
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("error reading config file:", err.Error())
		return cfg
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Println("error unmarshaling json to config:", err.Error())
	}
	return cfg
}

func (cfg *MonitorConfig) SaveConfig() {
	bytes, err := json.Marshal(cfg)
	if err != nil {
		log.Println("error marshaling config to json:", err.Error())
	}

	err = ioutil.WriteFile(configFilePath, bytes, 0644)
	if err != nil {
		log.Println("error writing config file:", err.Error())
	}
}
