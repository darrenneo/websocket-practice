package settings

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Settings struct {
	Secrets *Secrets
	Configs *Configs
}

type Secrets struct {
	Username string `json:",omitempty"`
	Password string `json:",omitempty"`
}

type Configs struct{}

var settings = Settings{}

func Get() Settings {
	return settings
}

func LoadSettings() error {
	l := log.New(os.Stdout, "settings: ", log.LstdFlags)

	l.Println("Loading settings...")

	secrets, err := getSecrets()
	if err != nil {
		l.Fatal(err)
	}
	settings.Secrets = secrets

	configs, err := getConfigs()
	if err != nil {
		l.Fatal(err)
	}
	settings.Configs = configs

	return nil
}

func getSecrets() (*Secrets, error) {
	results := &Secrets{}
	var secretsContent []byte

	secretFile := fmt.Sprintf("internal/settings/secrets.json")

	content, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file [%s]: %w", secretFile, err)
	}

	secretsContent = content
	if jsonErr := json.Unmarshal(secretsContent, results); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets: %w", jsonErr)
	}

	return results, nil
}

func getConfigs() (*Configs, error) {
	result := &Configs{}
	var configsContent []byte

	configFile := fmt.Sprintf("internal/settings/configs.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configsContent = []byte("{}")
	} else {
		content, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("error while reading config file [%s]: %w", configFile, err)
		}
		configsContent = content
	}

	if jsonErr := json.Unmarshal(configsContent, result); jsonErr != nil {
		return nil, fmt.Errorf(
			"error while parsing configsContent: %w",
			jsonErr,
		)
	}

	return result, nil
}
