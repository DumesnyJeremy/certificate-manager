package viper_fetcher

import (
	"errors"
	"github.com/DumesnyJeremy/lets-encrypt"
	"github.com/DumesnyJeremy/lets-encrypt/providers/dns"
	"github.com/DumesnyJeremy/notification-service"
	"github.com/spf13/viper"
	"os"

	"github.com/DumesnyJeremy/certificate-manager/manager"
	"github.com/DumesnyJeremy/certificate-manager/manager/fetcher"
	updater "github.com/DumesnyJeremy/certificate-manager/manager/updater"
)

type Config struct {
	CertManager     manager.CertManagerConfig             `mapstructure:"certificate_manager"`
	DNSServers      []dns.DNSServerConfig                 `mapstructure:"dns_servers"`
	Sites           []fetcher.CertificateFetchConfig      `mapstructure:"sites"`
	Updaters        []updater.CertificateUpdateConfig     `mapstructure:"updaters"`
	Notifiers       []notification_service.NotifierConfig `mapstructure:"notifiers"`
	LetsEncryptUser lets_encrypt.LetsEncryptUserConfig    `mapstructure:"lets_encrypt_user"`
	CertRootPath    string                                `mapstructure:"certificates_root_path"`
	RestartMinutes  int64                                 `mapstructure:"loop_restart_min"`
}

func ParseConfig(configFilePath string) (*Config, error) {
	fileType, err := findFileType(configFilePath)
	if err != nil {
			return nil, err
	}
	viper.SetConfigName("config")
	viper.SetConfigType(fileType)
	viper.AddConfigPath(configFilePath)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	configInfo, err := unmarshalServer()
	if err != nil {
		return nil, err
	}
	return &configInfo, nil
}

func findFileType(configFilePath string) (string, error) {
	filetypes := [3]string{"toml", "json", "yaml"}
	for _, filetype := range filetypes {
		_, err := os.Stat(configFilePath + "config." + filetype)
		if err == nil {
			return filetype, nil
		}
	}
	return "", errors.New("The configuration file wasn't found in " + configFilePath)
}

func unmarshalServer() (Config, error) {
	var configArray Config
	if err := viper.Unmarshal(&configArray); err != nil {
		return configArray, err
	}
	return configArray, nil
}
