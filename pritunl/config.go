package pritunl

import (
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

const (
	configFilePath = ".autovpn/pritunl"
)

type config struct {
	ID  string
	key string
}

type answer struct {
	Key string `survey:"key"`
}

func init() {
	// check if config folder exists
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, configFilePath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			color.Red("Error creating config folder: %v", err)
			os.Exit(1)
		}
	}
}

func initConfig(id string) {
	viper.SetConfigType("json")
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, configFilePath, id)
	viper.SetConfigFile(path)
}

func getConfig(id string) (*config, error) {
	initConfig(id)

	// check if config file exists
	if _, err := os.Stat(viper.ConfigFileUsed()); os.IsNotExist(err) {
		color.Yellow("Config not found. Please configure TOTP.")
		return configureOTP(id)
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return &config{
		ID:  id,
		key: viper.GetString("key"),
	}, nil
}

func updateOTP(id string) (*config, error) {
	return configureOTP(id)
}

func getAnswer() (*answer, error) {
	var qs = []*survey.Question{
		{
			Name:   "key",
			Prompt: &survey.Password{Message: "Enter TOTP key"},
		},
	}
	var ans answer
	if err := survey.Ask(qs, &ans); err != nil {
		return nil, err
	}
	return &ans, nil
}

func configureOTP(id string) (*config, error) {
	initConfig(id)

	ans, err := getAnswer()
	if err != nil {
		return nil, err
	}

	viper.Set("key", ans.Key)
	if err := viper.WriteConfig(); err != nil {
		return nil, err
	}

	color.Yellow("Config saved to %s\n\n", viper.ConfigFileUsed())
	return &config{
		ID:  id,
		key: ans.Key,
	}, nil
}
