package lib

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
)

func GetSecretConfig() *Config {

	filePtr := flag.String("file", "config/local.json", "config file location")
	flag.Parse()

	config := &Config{}
	configFile, errReadFile := os.ReadFile(*filePtr)
	if errReadFile != nil {
		log.Fatal(errReadFile.Error())
	}

	errUnMarshall := json.Unmarshal([]byte(configFile), config)
	if errUnMarshall != nil {
		log.Fatalf("Startup failed with error %s while unmarshling %s", errUnMarshall, configFile)
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

	return config
}
