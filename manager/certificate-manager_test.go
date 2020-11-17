package certificate_manager

import (
	"errors"
	"github.com/DumesnyJeremy/lets-encrypt"
	"github.com/DumesnyJeremy/lets-encrypt/providers/dns"
	"github.com/DumesnyJeremy/notification-service"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/DumesnyJeremy/certificate-manager/manager"
	"github.com/DumesnyJeremy/certificate-manager/manager/fetcher"
	"github.com/DumesnyJeremy/certificate-manager/manager/updater"
)

type ClientMock struct {
	mock.Mock
}

func (_m *ClientMock) DaysLeft() int {
	return 30
}
func (_m *ClientMock) GetConfig() fetcher.Config {
	return fetcher.Config{
		Server:   "Test server",
		URL:      "1.serv.io",
		Port:     443,
		Location: fetcher.LocationConfig{},
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

func FakeSitesCertificates(numberOfSites, dayLeftSites int) []fetcher.SiteCertProber {
	sitesCertificates := make([]fetcher.SiteCertProber, 0)
	for i := 1; i <= numberOfSites; i++ {
		siteCertificate := FakeSiteCertificate()
		sitesCertificates = append(sitesCertificates, siteCertificate)
	}
	return sitesCertificates
}

func FakeSiteCertificate() fetcher.SiteCertProber {
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
	manager.InitCertificateManager(manager.CertManagerConfig{},
		[]certificate_updater.CertificateUpdater{},
		[]fetcher.SitesPerDomain{},
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
	CertManager := manager.CertManager{Config: manager.CertManagerConfig{
		Recipients: []manager.RecipientConfig{{
			Notifier:   "Rocket",
			Categories: []string{"HIGH"},
			Dest:       []string{"toto"},
		}},
	},
		Notifiers: []notification_service.Notifier{tmp}}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
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
	CertManager := manager.CertManager{Config: manager.CertManagerConfig{
		Recipients: []manager.RecipientConfig{{
			Notifier:   "Rocket",
			Categories: []string{"HIGH"},
			Dest:       []string{"toto"},
		}},
	},
		Notifiers: []notification_service.Notifier{tmp}}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
	sitesToRenew := CertManager.GetSitesToRenew()
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, nil
	}
	if err := CertManager.ForceRenewForSite(sitesToRenew[0]); err == nil {
		t.Error("Error: ", err)
	}
}

func TestSuccessFailTriggerAlertForSite(t *testing.T) {
	CertManager := manager.CertManager{}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
	sitesToRenew := CertManager.GetSitesToRenew()
	RefreshCertifAndSendDayLeftMocked = func() (int, error) {
		return 50, errors.New("Forced error")
	}
	if err := CertManager.ForceRenewForSite(sitesToRenew[0]); err == nil {
		t.Error("Error: ", err)
	}
}

func TestFailTriggerAlertForSite(t *testing.T) {
	CertManager := manager.CertManager{}
	sites := FakeSitesCertificates(1, 30)

	CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
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
	CertManager := manager.CertManager{}
	for _, argument := range arguments {
		sites := FakeSitesCertificates(argument.numberOfSites, argument.dayLeftSites)

		CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
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
	CertManager := manager.CertManager{}
	for _, argument := range arguments {
		sites := FakeSitesCertificates(argument.numberOfSites, argument.dayLeftSites)

		CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
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
	CertManager := manager.CertManager{}
	for _, argument := range arguments {
		sites := FakeSitesCertificates(argument.numberOfSites, argument.dayLeftSites)

		CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
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
	CertManager := manager.CertManager{Config: manager.CertManagerConfig{
		Recipients: []manager.RecipientConfig{{
			Notifier:   "Rocket",
			Categories: []string{"ERROR"},
			Dest:       []string{"toto"},
		}},
	},
		Notifiers: []notification_service.Notifier{tmp}}
	CertManager.IndexedSites = fetcher.IndexSitesPerDomains(sites)
	CertManager.ParseSites()
}
