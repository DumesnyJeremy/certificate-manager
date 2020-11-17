package ssh

import (
	"errors"
	"github.com/hnakamur/go-scp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/DumesnyJeremy/certificate-manager/manager/fetcher"
	"github.com/DumesnyJeremy/certificate-manager/manager/updater"
)

type SSH struct {
	Client         *ssh.Client
	Config         certificate_updater.CertificateUpdateConfig
	CertifRootPath string
}

// Receives the config from the file, read and parse the id_rasa
// and returns a Signer from the encoded private key.
// Used the Signer to create an SSH client, and create the tcp connection with ssh.Dial .
func InitCertifUpdater(config certificate_updater.CertificateUpdateConfig, certifRootPath string) (certificate_updater.CertificateUpdater, error) {
	cu, err := initSSHUpdater(config, certifRootPath)
	if err != nil {
		return nil, err
	}
	return cu, nil
}

func initSSHUpdater(config certificate_updater.CertificateUpdateConfig, certifRootPath string) (*SSH, error) {
	privateKey, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, errors.New("Read private key for this config: " + config.RemoteConnection.Hostname + " fail: " + err.Error())
	}
	client, err := initSSHClient(config, signer)
	if err != nil {
		return nil, err
	}
	return &SSH{
		Client:         client,
		Config:         config,
		CertifRootPath: certifRootPath,
	}, nil
}

func initSSHClient(config certificate_updater.CertificateUpdateConfig, signer ssh.Signer) (*ssh.Client, error) {
	// Fill up client
	clientConfig := &ssh.ClientConfig{
		User: config.CertificatesOwner,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Create client connection
	client, err := ssh.Dial("tcp",
		config.RemoteConnection.Hostname+":"+strconv.Itoa(config.RemoteConnection.Port),
		clientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Send with the client create in the InitMulti,
// the Certificate and the Private key to the right place, given in the site configuration.
func (scu *SSH) UpdateCertificate(site fetcher.SiteCertProber) error {
	err := scp.NewSCP(scu.Client).SendFile(scu.CertifRootPath+"/"+site.GetConfig().URL+"/"+
		site.GetConfig().URL+".crt", site.GetConfig().Location.Certificate)
	if err != nil {
		return err
	}

	err = scp.NewSCP(scu.Client).SendFile(scu.CertifRootPath+"/"+site.GetConfig().URL+"/"+
		site.GetConfig().URL+".key", site.GetConfig().Location.PrivateKey)
	if err != nil {
		return err
	}
	return nil
}

// NewSession opens  for this client and execute the restart command.
func (scu *SSH) ReloadHTTPServer() error {
	defer scu.Client.Close()
	session, err := scu.Client.NewSession()
	if err != nil {
		return err
	}
	// Reload the scu
	if err := session.Run(scu.Config.RestartCMD); err != nil {
		return err
	}
	return nil
}

// Get the name of the updater used.
func (scu *SSH) GetName() string {
	return scu.Config.Name
}
