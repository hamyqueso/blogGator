package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	pathToFile, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	var c Config

	data, err := os.Open(pathToFile)
	if err != nil {
		// fmt.Println("error opening json file")
		// fmt.Printf("%v", err)
		return Config{}, err
	}
	defer data.Close()

	byteValue, err := io.ReadAll(data)
	if err != nil {
		return Config{}, err
	}

	if err := json.Unmarshal(byteValue, &c); err != nil {
		// fmt.Println("Error unmarshalling")
		// fmt.Printf("%v", err)
		return Config{}, err
	}

	return c, nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home dir")
		return "", err
	}

	return filepath.Join(home, configFileName), nil
}

func (config Config) SetUser(username string) (Config, error) {
	return Config{}, nil
}
