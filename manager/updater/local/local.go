package local

import (
	"os/exec"

	"github.com/DumesnyJeremy/certificate-manager/manager/fetcher"
	updater "github.com/DumesnyJeremy/certificate-manager/manager/updater"
)

type Local struct {
	Config         updater.CertificateUpdateConfig
	CertifRootPath string
}

// Receives the config in argument and and create an object with the config and the certificate root path.
func InitCertifUpdater(config updater.CertificateUpdateConfig, certifRootPath string) (updater.CertificateUpdater, error) {
	cu, err := initLocalUpdater(config, certifRootPath)
	if err != nil {
		return nil, err
	}
	return cu, nil
}

func initLocalUpdater(config updater.CertificateUpdateConfig, certifRootPath string) (*Local, error) {
	return &Local{
		Config:         config,
		CertifRootPath: certifRootPath,
	}, nil
}

// Send with the client create in the InitMulti,
// the Certificate and the Private key to the right place, given in the site configuration.
func (lcu *Local) UpdateCertificate(site fetcher.SiteCertProber) error {
	// Copy the Certificate to the right place given in site config
	_, err := exec.Command("cp " + lcu.CertifRootPath + "/" + site.GetConfig().URL + "/" +
		site.GetConfig().URL + ".crt " + site.GetConfig().Location.Certificate).Output()
	if err != nil {
		return err
	}
	// Copy the Private Key to the right place given in site config
	_, err = exec.Command("cp " + lcu.CertifRootPath + "/" + site.GetConfig().URL + "/" +
		site.GetConfig().URL + ".key " + site.GetConfig().Location.PrivateKey).Output()
	if err != nil {
		return err
	}
	_, err = exec.Command("chown " + lcu.Config.CertificatesOwner + ":" + lcu.Config.CertificatesOwner + " " +
		site.GetConfig().Location.Certificate).Output()
	if err != nil {
		return err
	}
	_, err = exec.Command("chown " + lcu.Config.CertificatesOwner + ":" + lcu.Config.CertificatesOwner + " " +
		site.GetConfig().Location.PrivateKey).Output()
	if err != nil {
		return err
	}
	_, err = exec.Command("chown " + lcu.Config.CertificatesOwner + " " + site.GetConfig().Location.Certificate).Output()
	if err != nil {
		return err
	}
	_, err = exec.Command("chown " + lcu.Config.CertificatesOwner + " " + site.GetConfig().Location.PrivateKey).Output()
	if err != nil {
		return err
	}
	return nil
}

// Retrieve the restart command and execute the named program with
// the given arguments.
func (lcu *Local) ReloadHTTPServer() error {
	_, err := exec.Command(lcu.Config.RestartCMD).Output()
	if err != nil {
		return err
	}
	return nil
}

// Get the name of the updater used.
func (lcu *Local) GetName() string {
	return lcu.Config.Name
}
