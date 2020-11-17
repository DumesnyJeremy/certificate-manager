package manager

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/DumesnyJeremy/lets-encrypt"
	"github.com/DumesnyJeremy/lets-encrypt/providers/dns"
	"github.com/DumesnyJeremy/notification-service"

	"github.com/DumesnyJeremy/certificate-manager/manager/fetcher"
	"github.com/DumesnyJeremy/certificate-manager/manager/updater"
)

// Used to respect the Let's Encrypt rate limits. See:
// https://letsencrypt.org/docs/rate-limits/ .
const (
	MaxSitesPerDomain        = 200
	DaysAfterRenew           = 90
	MaxRenewPerDomainPerWeek = 50
	DaysLimitToRenew         = 30
)

type CertManagerConfig struct {
	Recipients []RecipientConfig `mapstructure:"recipients"`
}

type RecipientConfig struct {
	Notifier   string   `mapstructure:"notifier"`
	Categories []string `mapstructure:"categories"`
	Dest       []string `mapstructure:"dest"`
}

// Interface with 3 methods used to detect if all the sites in one domain
// respect the limits that Let's Encrypt has established.
type SitesLimitInfo interface {
	remainQueriesForNextWeek(domain fetcher.SitesPerDomain)
	lastMonthValiditySitesNumber(domain fetcher.SitesPerDomain) int
	lastWeekValiditySitesNumber(domain fetcher.SitesPerDomain) int
}

// CertManager receives all the object creat from the configuration files
// with all the InitMulti methods in the main.
// This object will be used to implement all the CertManager class methods.
//
// We will be able to:
//
// * Update the certificate with the 'CertificateUpdater'
// by using the ssh implementation or the local one.
//
// * Verify the remaining validity of this certificate with 'Client'.
//
// * Can receive a notification by mail or rocket about what's happening with 'Notifier'.
//
// * Validate the challenge provided by Let's Encrypt by the implementation of the DNS Challenge with 'DNSServer'.
//
// * Communicate with Let's Encrypt to ask a Challenge, and and get the certificate with 'LetsEncrypt'.
type CertManager struct {
	Config              CertManagerConfig                        // Configuration files.
	ConfDirPath         string                                   // Path of the configuration directory.
	CertificateUpdaters []certificate_updater.CertificateUpdater // Methods used to renew certificates.
	IndexedSites        []fetcher.SitesPerDomain                 // Methods used to upload certificates.
	Notifiers           []notification_service.Notifier          // Methods used to send notification.
	DNSServers          []dns.DNSServer                          // Methods used to accomplish DNS Challenges.
	LetsEncrypt         lets_encrypt.LetsEncrypt                 // Used to communicate with Let's Encrypt.
}

//Initialization of the Certificate Manager structure.
func InitCertificateManager(CertificateManager CertManagerConfig,
	certificateUpdaters []certificate_updater.CertificateUpdater,
	sitesPerDomain []fetcher.SitesPerDomain,
	notifiers []notification_service.Notifier,
	dnsServers []dns.DNSServer,
	LetsEncrypt lets_encrypt.LetsEncrypt,
	confDirPath string) (*CertManager, error) {
	return &CertManager{
		Config:              CertificateManager,
		IndexedSites:        sitesPerDomain,
		CertificateUpdaters: certificateUpdaters,
		Notifiers:           notifiers,
		DNSServers:          dnsServers,
		LetsEncrypt:         LetsEncrypt,
		ConfDirPath:         confDirPath,
	}, nil
}

// Use an array of site that need a renew this between this week and this month,
// it will only take 50 sites per domain every week to do respect the Let's Encrypt rate limits.
//
// If an error occurs during the renew, it will send to the recipients who have the ERROR categories
// in the configuration file.
func (CertManager *CertManager) ParseSites() {
	sitesToRenew := CertManager.GetSitesToRenew()
	for _, site := range sitesToRenew {
		err := CertManager.Renew(site)
		if err != nil {
			if err := CertManager.sendToRecipientsByCategories(
				"["+site.GetConfig().URL+"] "+"Error: "+err.Error()+";",
				"ERROR"); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

// Method who receive a site in parameter and set the DNS Provider to accomplish the
// Let's Encrypt challenge receive the new Certificate.
// After that, we will update the certificate and reload it by SSH.
func (CertManager *CertManager) Renew(site fetcher.SiteCertProber) error {
	// Find the DNS Server of the site given.
	DNSServer, err := CertManager.GetDNSProviderForSite(site.GetConfig().URL)
	if err != nil {
		return err
	}
	// Set the DNS Provider, let's encrypt challenge.
	if err := CertManager.LetsEncrypt.SetDNSProvider(dns.DNSProvider{DNSServer: DNSServer}); err != nil {
		return err
	}
	if err := CertManager.LetsEncrypt.AskCertificate(site.GetConfig().URL); err != nil {
		return err
	}
	// Use the certificate for the correct server.
	for _, CertificateUpdater := range CertManager.CertificateUpdaters {
		if CertificateUpdater.GetName() == site.GetConfig().Server {
			if err := CertificateUpdater.UpdateCertificate(site); err != nil {
				return err
			}
			if err := CertificateUpdater.ReloadHTTPServer(); err != nil {
				return err
			}
		}
	}
	if err := CertManager.sendToRecipientsByCategories(
		"["+site.GetConfig().URL+"] "+"New certificate upload;",
		"RENEW"); err != nil {
		return err
	}
	return nil
}

// Find the authoritative DNS Server for the given site.
func (CertManager *CertManager) GetDNSProviderForSite(siteURL string) (dns.DNSServer, error) {
	for _, DNSServer := range CertManager.DNSServers {
		if DNSServer.IsAuthoritativeForDomain(siteURL) {
			return DNSServer, nil
		}
	}
	return nil, errors.New("Didn't found it's DNS server")
}

func (CertManager *CertManager) sendToRecipientsByCategories(msg string, renewOrError string) error {
	// Parse recipients of the configuration file
	for _, recipient := range CertManager.Config.Recipients {
		for _, recipientCategories := range recipient.Categories {
			// Find the recipients categories match
			if renewOrError == recipientCategories {
				// Retrieve the notifier
				if err := CertManager.parseAllNotifiers(recipient, msg, renewOrError); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Receives recipients with the message.
// Parse all of them and check with one is corresponding.
func (CertManager *CertManager) parseAllNotifiers(recipient RecipientConfig, msg string, renewOrError string) error {
	for _, notifier := range CertManager.Notifiers {
		// If the recipient match with the current notifier
		if strings.EqualFold(notifier.GetName(), recipient.Notifier) {
			if err := sendToAllRecipients(recipient, notifier, msg, renewOrError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Receives recipients, the type of notification with the message and it type.
// With all these information, the methods will parse every recipient and Send the message.
func sendToAllRecipients(recipient RecipientConfig, notifier notification_service.Notifier, msg string, renewOrError string) error {
	// Parse all the recipients to send the message.
	for _, dest := range recipient.Dest {
		typeOfSend, err := notifier.SendMessage(msg, dest)
		if err != nil {
			return err
		}
		// Log what's going on
		if renewOrError == "ERROR" {
			log.Error(msg, " ", typeOfSend+" to ", dest)
		} else {
			log.Info(msg, " ", typeOfSend+" to ", dest)
		}
	}
	return nil
}

// Get a domain and a number of days,
// and return the amount of certificates that need to be renewed before the given day's.
func (CertManager *CertManager) GetSitesQtyToRenewBefore(days int, domain fetcher.SitesPerDomain) int {
	var SitesToRenewThisWeek int
	for _, site := range domain.Sites {
		if site.DaysLeft() <= days {
			SitesToRenewThisWeek += 1
		}
	}
	return SitesToRenewThisWeek
}

// Get a domain and a number of day's,
// and return the amount of remaining queries in the next given day's.
func (CertManager *CertManager) GetRemainingLEQueriesUntil(days int, domain fetcher.SitesPerDomain) int {
	QueriesMAXForNexWeek := MaxRenewPerDomainPerWeek
	for _, site := range domain.Sites {
		if site.DaysLeft() > DaysAfterRenew-days {
			QueriesMAXForNexWeek -= 1
		}
	}
	return QueriesMAXForNexWeek
}

// Get an indexed list of sites for a specific domain and return only the sites to renew.
// Only the sites that can be renewed in one shot (LE limit requests) will be returned.
func (CertManager *CertManager) tookOfSitesToRenew(domain fetcher.SitesPerDomain) []fetcher.SiteCertProber {
	siteToRenew := make([]fetcher.SiteCertProber, 0)
	sitesUnder30DaysLeft := CertManager.GetSitesQtyToRenewBefore(30, domain)
	availableQueries := CertManager.GetRemainingLEQueriesUntil(7, domain)
	if sitesUnder30DaysLeft > CertManager.GetSitesQtyToRenewBefore(7, domain) {
		log.Warn("For [", domain.Name, "]; only the most dangerous sites will be renew.")
	}
	queriesToPerform := 0
	for _, site := range domain.Sites {
		if queriesToPerform >= availableQueries {
			return siteToRenew
		}
		if site.DaysLeft() <= DaysLimitToRenew {
			siteToRenew = append(siteToRenew, site)
			queriesToPerform += 1
		}
	}
	return siteToRenew
}

// Discard a domain if it doesn't respect the Let's Encrypt rate limits.
//
// No more than 200 sites per domain.
// No more than 50 site with a certificate days left less than 7.
// No more site to renew this week than queries left for this domain.
func (CertManager *CertManager) discardNonRenewableDomains() {
	indexedSitesCopies := CertManager.IndexedSites
	for index, domain := range indexedSitesCopies {
		if CertManager.GetSitesQtyToRenewBefore(30, domain) > MaxSitesPerDomain {
			indexedSitesCopies = append(indexedSitesCopies[:index], indexedSitesCopies[index+1:]...)
			log.Error("Discard [", domain.Name, "]; Number of site for this domains > 200: 'https://letsencrypt.org/docs/rate-limits/'")
		} else if CertManager.GetSitesQtyToRenewBefore(7, domain) > MaxRenewPerDomainPerWeek {
			indexedSitesCopies = append(indexedSitesCopies[:index], indexedSitesCopies[index+1:]...)
			log.Error("Discard [", domain.Name, "]; Number of site to renew in a week for this domains > 50: 'https://letsencrypt.org/docs/rate-limits/' ")
		} else if CertManager.GetRemainingLEQueriesUntil(7, domain) < CertManager.GetSitesQtyToRenewBefore(7, domain) {
			indexedSitesCopies = append(indexedSitesCopies[:index], indexedSitesCopies[index+1:]...)
			log.Error("Discard [", domain.Name, "]; Number of site to renew in a week for this domains > queries week left: 'https://letsencrypt.org/docs/rate-limits/'")
		}
	}
	CertManager.IndexedSites = indexedSitesCopies
}

// Receives the indexed domains and return an array with only the sites which need to be renew.
func (CertManager *CertManager) GetSitesToRenew() []fetcher.SiteCertProber {
	siteToRenew := make([]fetcher.SiteCertProber, 0)
	CertManager.discardNonRenewableDomains()
	for _, domain := range CertManager.IndexedSites {
		siteToRenew = append(siteToRenew, CertManager.tookOfSitesToRenew(domain)...)
	}
	return siteToRenew
}

// Force a renew for a specific site_cert_prober.SiteCertProber (regardless of its actual day's left)
// and send a message by rocket or mail concerning the result of the renewal.
func (CertManager *CertManager) ForceRenewForSite(siteCertificate fetcher.SiteCertProber) error {
	if err := CertManager.Renew(siteCertificate); err != nil {
		return err
	}
	if err := CertManager.sendToRecipientsByCategories("["+siteCertificate.GetConfig().URL+"] "+"Force Renew;", "RENEW"); err != nil {
		return err
	}
	return nil
}
