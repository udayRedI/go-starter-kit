package lib

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

func GetSecrets(config *Config) {

	configFile, errReadFile := os.ReadFile("app/system_configs/config.local.json")
	if errReadFile != nil {
		log.Fatal(errReadFile.Error())
	}

	errUnMarshall := json.Unmarshal([]byte(configFile), config)
	if errUnMarshall != nil {
		log.Fatal("Error Unmarshal config.json during startup")
	}

	if config.IsValid() == false {
		errTxt := "Invalid config file please fix it"
		err := errors.New(errTxt)
		CheckFatal(err, errTxt)
	}

	if config.ENV != "LOCAL" {
		InitiateSentry(config)
	} else {
		for key, value := range (*config).AWSSecrets {
			os.Setenv(key, value)
		}
	}
}
