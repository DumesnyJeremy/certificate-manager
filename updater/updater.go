package certificate_updater

import (
	"github.com/DumesnyJeremy/certificate-manager/fetcher"
)

// The 2 types of implementation to upload a certificate.
const LocalAccessType = "local"
const RemoteAccessType = "remote"

type CertificateUpdater interface {
	UpdateCertificate(site fetcher.SiteCertProber) error
	ReloadHTTPServer() error
	GetName() string
}

type CertificateUpdateConfig struct {
	Name              string               `mapstructure:"name"`
	Type              string               `mapstructure:"type"`
	CertificatesOwner string               `mapstructure:"certificates_owner"`
	RemoteConnection  infoRemoteConnection `mapstructure:"remote_connection"`
	RestartCMD        string               `mapstructure:"reload_cmd"`
}

type infoRemoteConnection struct {
	Protocol string `mapstructure:"protocol"`
	Port     int    `mapstructure:"port"`
	Hostname string `mapstructure:"hostname"`
}
