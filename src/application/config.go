package application

import (
	"encoding/json"
	"os"

	"github.com/jqwel/xlipboard/src/utils"
)

const ConfigFile = "Config.json"
const LogFile = "log.txt"

type Config struct {
	Port         string        `json:"Port"`
	Authkey      string        `json:"Authkey"`
	Certificate  string        `json:"-"`
	PrivateKey   string        `json:"-"`
	Mount        string        `json:"-"`
	NtpAddress   string        `json:"NtpAddress"` // 同步时间服务器地址
	SyncSettings []SyncSetting `json:"SyncSettings"`
}
type SyncSetting struct {
	Target string `json:"Target"`
}

var DefaultConfig = Config{
	Port:        "3216",
	Authkey:     utils.RandStringBytes(64),
	Certificate: "",
	PrivateKey:  "",
	NtpAddress:  "ntp.aliyun.com",
	SyncSettings: []SyncSetting{
		{
			Target: "192.168.200.101:3216",
		},
	},
}

func init() {
	_, data, _ := utils.GenerateCertificate(nil)
	DefaultConfig.Certificate = data.Certificate
	DefaultConfig.PrivateKey = data.PrivateKey
}

func LoadConfig(path string) (*Config, error) {
	if utils.IsExistFile(path) {
		return loadConfigFromFile(path)
	}
	if err := createConfigFile(path); err != nil {
		return nil, err
	}
	return &DefaultConfig, nil
}

func loadConfigFromFile(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configBytes, &DefaultConfig); err != nil {
		return nil, err
	}
	return &DefaultConfig, nil
}

func createConfigFile(path string) error {
	defaultConfigJSON, err := json.MarshalIndent(DefaultConfig, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, defaultConfigJSON, 0744); err != nil {
		return err
	}
	return nil
}
