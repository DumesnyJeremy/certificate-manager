package main

import (
	"errors"
	"flag"
	"github.com/DumesnyJeremy/lets-encrypt"
	"github.com/DumesnyJeremy/lets-encrypt/providers/dns"
	"github.com/DumesnyJeremy/lets-encrypt/providers/dns/gandi"
	"github.com/DumesnyJeremy/lets-encrypt/providers/dns/pdns"
	"github.com/DumesnyJeremy/notification-service"
	"github.com/DumesnyJeremy/notification-service/mail"
	"github.com/DumesnyJeremy/notification-service/rocket"
	legoLog "github.com/go-acme/lego/v4/log"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/DumesnyJeremy/certificate-manager/manager"
	"github.com/DumesnyJeremy/certificate-manager/manager/fetcher"
	updater "github.com/DumesnyJeremy/certificate-manager/manager/updater"
	"github.com/DumesnyJeremy/certificate-manager/manager/updater/local"
	"github.com/DumesnyJeremy/certificate-manager/manager/updater/ssh"
	"github.com/DumesnyJeremy/certificate-manager/viper-fetcher"
)

func main() {
	// Retrieve cmdline args
	confDirPath := flag.String("confdir", "/etc/certificate-manager/", "File configuration path")
	execType := flag.Bool("d", false, "to run the prog as a daemon (loop)")
	flag.Parse()

	// Initialize custom logger
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02/01/2006 15:04:05"
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
	legoLogger := log.StandardLogger()
	legoLogger.SetFormatter(formatter)
	legoLog.Logger = legoLogger

	// Parse config using Viper
	config, err := viper_fetcher.ParseConfig(*confDirPath)
	if err != nil {
		log.Fatal("Fatal error while configuration: ", err)
	}

	// InitMulti Notifiers
	notifiers := initNotifiers(config.Notifiers)

	// InitMulti DNS servers
	dnsServers := initDNSServers(config.DNSServers)

	// InitMulti certificate uploaders
	servers := initCertUpdaters(config.Updaters, config.CertRootPath)

	// InitMulti certificate analyzers
	siteCert := fetcher.InitMulti(config.Sites)

	// InitMulti let's encrypt user/account
	letsEncryptCustomUser, err := lets_encrypt.InitLetsEncryptUser(config.LetsEncryptUser)
	if err != nil {
		log.Error("While InitMulti Let's User: " + err.Error())
	}

	// Index Sites per domain
	indexedSitesPerDomain := fetcher.IndexSitesPerDomains(siteCert)

	// InitMulti let's encrypt
	letsEncrypt, err := lets_encrypt.InitLetsEncrypt(config.CertRootPath, letsEncryptCustomUser.GetLEUser())
	if err != nil {
		log.Error("While InitMulti Let's: " + err.Error())
	}
	// InitMulti alert manager
	CertManager, err := manager.InitCertificateManager(
		config.CertManager,
		servers,
		indexedSitesPerDomain,
		notifiers,
		dnsServers,
		letsEncrypt,
		*confDirPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	//	Main loop
	for {
		CertManager.ParseSites()
		if *execType == true {
			time.Sleep(time.Duration(config.RestartMinutes) * time.Minute)
		} else {
			return
		}
	}
}

func initDNSServers(dnsServersConfig []dns.DNSServerConfig) []dns.DNSServer {
	dnsServers := make([]dns.DNSServer, 0)
	for _, DNSServerConfig := range dnsServersConfig {
		if dnsServer, err := initDNSServer(DNSServerConfig); dnsServer != nil {
			dnsServers = append(dnsServers, dnsServer)
		} else {
			log.Error("While InitDNSServer: ", err.Error())
		}
	}
	return dnsServers
}

func initDNSServer(dnsServerConfig dns.DNSServerConfig) (dns.DNSServer, error) {
	switch dnsServerConfig.Type {
	case dns.ServerDNSTypeGandy:
		return gandi.InitDNSServer(dnsServerConfig)
	case dns.ServerDNSTypePDNS:
		return pdns.InitDNSServer(dnsServerConfig)
	default:
		return nil, errors.New("Didn't found the type")
	}
}

func initNotifiers(notifiersConfig []notification_service.NotifierConfig) []notification_service.Notifier {
	notifiers := make([]notification_service.Notifier, 0)
	for _, notifierConfig := range notifiersConfig {
		if notifier, err := initNotifier(notifierConfig); notifier != nil {
			notifiers = append(notifiers, notifier)
		} else {
			log.Error("While InitNotifier: ", err.Error())
		}
	}
	return notifiers
}

func initNotifier(notifierConfig notification_service.NotifierConfig) (notification_service.Notifier, error) {
	switch notifierConfig.Type {
	case notification_service.NotifierTypeRocket:
		return rocket.InitNotifier(notifierConfig)
	case notification_service.NotifierTypeMail:
		return mail.InitNotifier(notifierConfig)
	default:
		log.Error(errors.New("Didn't found the type"))
		return nil, errors.New("Didn't found the type")
	}
}

func initCertUpdaters(servers []updater.CertificateUpdateConfig, CertificatesRootPath string) []updater.CertificateUpdater {
	ClientsConfigs := make([]updater.CertificateUpdater, 0)
	for _, server := range servers {
		if client, err := initCertifUpdater(server, CertificatesRootPath); err == nil {
			ClientsConfigs = append(ClientsConfigs, client)
		} else {
			log.Error("While InitCertifUpdater: " + err.Error())
		}
	}
	return ClientsConfigs
}

func initCertifUpdater(server updater.CertificateUpdateConfig, CertificatesRootPath string) (updater.CertificateUpdater, error) {
	switch server.Type {
	case updater.RemoteAccessType:
		return ssh.InitCertifUpdater(server, CertificatesRootPath)
	case updater.LocalAccessType:
		return local.InitCertifUpdater(server, CertificatesRootPath)
	default:
		return nil, errors.New("Didn't found the type")
	}
}
