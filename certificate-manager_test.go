package certificate_manager

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"testing"

	"ssl-ar/certificate-updater"
	"ssl-ar/lets-encrypt"
	"ssl-ar/lets-encrypt/providers/dns"
	"ssl-ar/notification-service"
	"ssl-ar/certificate-prober"
)

type ClientMock struct {
	mock.Mock
}

func (_m *ClientMock) DaysLeft() int {
	return 30
}
func (_m *ClientMock) GetConfig() certificate_prober.Config {
	return certificate_prober.Config{
		Server:   "Test server",
		URL:      "1.serv.io",
		Port:     443,
		Location: certificate_prober.LocationConfig{},
	}
}
func (_m *ClientMock) IsSiteValid() bool {
	return true
}
func (_m *ClientMock) Refresh() error {
	return nil
}
func (_m *ClientMock) GetDomain() string {
	return ""
}
func (_m *ClientMock) RefreshCertifAndGetDaysLeft() (int, error) {
	return RefreshCertifAndSendDayLeftMocked()
}

var RefreshCertifAndSendDayLeftMocked func() (int, error)

func FakeSitesCertificates(numberOfSites, dayLeftSites int) []certificate_prober.SiteCertProber {
	sitesCertificates := make([]certificate_prober.SiteCertProber, 0)
	for i := 1; i <= numberOfSites; i++ {
		siteCertificate := FakeSiteCertificate()
		sitesCertificates = append(sitesCertificates, siteCertificate)
	}
	return sitesCertificates
}

func FakeSiteCertificate() certificate_prober.SiteCertProber {
	return &ClientMock{}
}

type Notif struct{}

func (notifier *Notif) SendMessage(msg string, dest string) (string, error) {
	return SendMessageMocked()
}

var SendMessageMocked func() (string, error)

func (notifier *Notif) GetName() string {
	return "Rocket"
}

func TestInitCertManager(t *testing.T) {
	InitCertificateManager(CertManagerConfig{},
		[]certificate_updater.CertificateUpdater{},
		[]certificate_prober.SitesPerDomain{},
		[]notification_service.Notifier{},
		[]dns.DNSServer{},
		lets_encrypt.LetsEncrypt{},
		"test")
}

func TestSuccessForceRenew(t *testing.T) {
	SendMessageMocked = func() (string, error) {
		return "", nil
	}
	tmp := new(Notif)
	CertManager := CertManager{Config: CertManagerConfig{
		Recipients: []RecipientConfig{{
			Notifier:   "Rocket",
			Categories: []string{"HIGH"},
			Dest:       []string{"toto"},
		}},
	},
		Notifiers: []notification_service.Notifier{tmp}}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
	sitesToRenew := CertManager.GetSitesToRenew()
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, nil
	}
	if err := CertManager.ForceRenewForSite(sitesToRenew[0]); err == nil {
		t.Error("Error: ", err)
	}
}

func TestSendMessageErrorTriggerAlertForSite(t *testing.T) {
	SendMessageMocked = func() (string, error) {
		return "", errors.New("Fake error.")
	}
	tmp := new(Notif)
	CertManager := CertManager{Config: CertManagerConfig{
		Recipients: []RecipientConfig{{
			Notifier:   "Rocket",
			Categories: []string{"HIGH"},
			Dest:       []string{"toto"},
		}},
	},
		Notifiers: []notification_service.Notifier{tmp}}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
	sitesToRenew := CertManager.GetSitesToRenew()
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, nil
	}
	if err := CertManager.ForceRenewForSite(sitesToRenew[0]); err == nil {
		t.Error("Error: ", err)
	}
}

func TestSuccessFailTriggerAlertForSite(t *testing.T) {
	CertManager := CertManager{}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
	sitesToRenew := CertManager.GetSitesToRenew()
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, errors.New("Forced error")
	}
	if err := CertManager.ForceRenewForSite(sitesToRenew[0]); err == nil {
		t.Error("Error: ", err)
	}
}

func TestFailTriggerAlertForSite(t *testing.T) {
	CertManager := CertManager{}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, errors.New("Forced error")
	}
	sitesToRenew := CertManager.GetSitesToRenew()
	if err := CertManager.ForceRenewForSite(sitesToRenew[0]); err == nil {
		t.Error("Error: ", err)
	}
}

func TestCheckRenewReturnRenew(t *testing.T) {
	var arguments = []struct {
		numberOfSites int
		dayLeftSites  int
	}{
		// It will begin the renew when the remaining days are under 30.
		{1, 40},
		// More than 50 site to renew in less than one week
		{1, 5},
	}
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 10, nil
	}
	CertManager := CertManager{}
	for _, argument := range arguments {
		sites := FakeSitesCertificates(argument.numberOfSites, argument.dayLeftSites)

		CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
		sitesToRenew := CertManager.GetSitesToRenew()
		if err := CertManager.Renew(sitesToRenew[0]); err == nil {
			t.Error("Error: ", err)
		}
	}
}

func TestCheckRenewReturnHigh(t *testing.T) {
	var arguments = []struct {
		numberOfSites int
		dayLeftSites  int
	}{
		// It will begin the renew when the remaining days are under 30.
		{1, 40},
		// More than 50 site to renew in less than one week
		{1, 5},
	}
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 10, nil
	}
	CertManager := CertManager{}
	for _, argument := range arguments {
		sites := FakeSitesCertificates(argument.numberOfSites, argument.dayLeftSites)

		CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
		sitesToRenew := CertManager.GetSitesToRenew()
		if err := CertManager.Renew(sitesToRenew[0]); err == nil {
			t.Error("Error: ", err)
		}
	}
}

func TestSuccessCheckRenew(t *testing.T) {
	var arguments = []struct {
		numberOfSites int
		dayLeftSites  int
	}{
		// It will begin the renew when the remaining days are under 30.
		{1, 40},
		// More than 50 site to renew in less than one week
		{1, 5},
	}
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 10, nil
	}
	CertManager := CertManager{}
	for _, argument := range arguments {
		sites := FakeSitesCertificates(argument.numberOfSites, argument.dayLeftSites)

		CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
		sitesToRenew := CertManager.GetSitesToRenew()
		if err := CertManager.Renew(sitesToRenew[0]); err == nil {
			t.Error("Error: ", err)
		}
	}
}

func TestParseSites(t *testing.T) {
	sites := FakeSitesCertificates(1, 20)
	SendMessageMocked = func() (string, error) {
		return "", errors.New("Fake error.")
	}
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, errors.New("Forced error")
	}
	tmp := new(Notif)
	CertManager := CertManager{Config: CertManagerConfig{
		Recipients: []RecipientConfig{{
			Notifier:   "Rocket",
			Categories: []string{"ERROR"},
			Dest:       []string{"toto"},
		}},
	},
		Notifiers: []notification_service.Notifier{tmp}}
	CertManager.IndexedSites = certificate_prober.IndexSitesPerDomains(sites)
	CertManager.ParseSites()
}
