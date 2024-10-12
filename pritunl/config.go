package pritunl

import (
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/xlzd/gotp"
)

const (
	configFilePath = ".autovpn/pritunl"
)

type config struct {
	ID  string
	key string
	pin string
}

func (c *config) OTP() string {
	totp := gotp.NewDefaultTOTP(c.key)
	return c.pin + totp.Now()
}

type answer struct {
	Key string
	Pin string
}

func initConfig(id string) {
	viper.SetConfigType("json")
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, configFilePath, id)
	viper.SetConfigFile(path)
}

func getConfig(id string) (*config, error) {
	initConfig(id)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			color.Yellow("Config not found. Please configure OTP.")
			return configureOTP(id)
		}
	}

	return &config{
		ID:  id,
		key: viper.GetString("key"),
		pin: viper.GetString("pin"),
	}, nil
}

func updateOTP(id string) (*config, error) {
	return configureOTP(id)
}

func getAnswer() (*answer, error) {
	var qs = []*survey.Question{
		{
			Name:   "key",
			Prompt: &survey.Password{Message: "Enter OTP key"},
		},
		{
			Name:   "pin",
			Prompt: &survey.Password{Message: "Enter Pin"},
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
	viper.Set("pin", ans.Pin)
	if err := viper.WriteConfig(); err != nil {
		return nil, err
	}

	color.Yellow("Config saved to %s\n\n", viper.ConfigFileUsed())
	return &config{
		ID:  id,
		key: ans.Key,
		pin: ans.Pin,
	}, nil
}
