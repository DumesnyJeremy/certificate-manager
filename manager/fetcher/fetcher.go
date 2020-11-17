package fetcher

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"strconv"
	"time"
)

type SiteCertProber interface {
	DaysLeft() int
	RefreshCertifAndGetDaysLeft() (int, error)
	IsSiteValid() bool
	Refresh() error
	GetConfig() Config
	GetDomain() string
}

type Config struct {
	Server   string         `mapstructure:"server"`
	URL      string         `mapstructure:"url"`
	Port     int            `mapstructure:"port"`
	Location LocationConfig `mapstructure:"location"`
}

type LocationConfig struct {
	PrivateKey  string `mapstructure:"private_key"`
	Certificate string `mapstructure:"certificate"`
}

// Main client containing an X.509 certificate and the analyzer config.
type Client struct {
	Domain      string
	Certificate *x509.Certificate
	Config      Config
}

// Receives multi analyzer configs, parse site by site and
// build an array of clients (extracted certificate).
func InitMulti(configSites []Config) []SiteCertProber {
	sitesCertificates := make([]SiteCertProber, 0)
	for _, configSite := range configSites {
		siteCertificate, err := Init(configSite)
		if err == nil {
			sitesCertificates = append(sitesCertificates, siteCertificate)
		} else {
			log.Error(err.Error())
		}
	}
	return sitesCertificates
}

// Build a client by extracting certificate from site.
func Init(siteConfig Config) (SiteCertProber, error) {
	domain, err := publicsuffix.Domain(siteConfig.URL)
	if err != nil {
		log.Error("For [", siteConfig.URL, "]; can't found the domain; ", err)
	}
	extract, err := extractCertificate(siteConfig)
	if err != nil {
		return nil, err

	}
	return &Client{
		Domain:      domain,
		Certificate: extract,
		Config:      siteConfig,
	}, nil
}

// Method to refresh the certificate of a site, used to be sure the days left change to 90 days left.
func (certifExtract *Client) Refresh() error {
	extract, err := extractCertificate(certifExtract.Config)
	if err != nil {
		return err
	}
	certifExtract.Certificate = extract
	return nil
}

// Method who return the certificate validity days left.
func (certifExtract *Client) DaysLeft() int {
	return int(ComputeDaysLeft(certifExtract.Certificate))
}

// Method to check if this certificate belongs to the site request.
func (certifExtract *Client) IsSiteValid() bool {
	return certifExtract.Certificate.Subject.CommonName == certifExtract.Config.URL
}

// Regroup the 3 methods above and return the number of days remaining.
func (certifExtract *Client) RefreshCertifAndGetDaysLeft() (int, error) {
	// Refresh certificates of the siteConfig
	err := certifExtract.Refresh()
	if err != nil {
		return 0, err
	}

	// Check how many day remaining for the certificates
	day := certifExtract.DaysLeft()
	if day == 0 {
		return 0, nil
	}

	// Look if the CommonName match with the SiteUrl
	if !certifExtract.IsSiteValid() {
		return day, errors.New("The domain name is wrong for this siteConfig: " + certifExtract.Config.URL)
	}
	return day, err
}

func (certifExtract *Client) GetConfig() Config {
	return certifExtract.Config
}

func (certifExtract *Client) GetDomain() string {
	return certifExtract.Domain
}

// Dial connects to the given network address using net.Dial,
// is a valid certificate for the named host
func extractCertificate(site Config) (*x509.Certificate, error) {
	fullSite := site.URL + ":" + strconv.Itoa(site.Port)
	conn, err := tls.Dial("tcp", fullSite, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	for _, peerCertificate := range conn.ConnectionState().PeerCertificates {
		err := peerCertificate.VerifyHostname(site.URL)
		if err == nil {
			return peerCertificate, nil
		}
	}
	return nil, errors.New("No valid certificate found for [" + site.URL + "].")
}

func ComputeDaysLeft(certificate *x509.Certificate) int64 {
	return int64(certificate.NotAfter.Local().Sub(time.Now().Local()).Hours() / 24)
}
