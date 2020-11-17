# certificate-manager

This is a Go program who uses the [lets-encrypt](github.com/DumesnyJeremy/lets-encrypt) library and the
[notification-service](github.com/DumesnyJeremy/notification-service) library.

This is a Go program, which aims to check the number of days remaining for sites certificates given and renew them automatically.
Certificates will be renewed during the last month of validity, thanks to the library.
The program will decide, in function of the [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/) .  
If a domain didn't respect one of the limits, it will be discard, and you will receive an alert, thanks to the library,
otherwise the renewal will take place normally.  
If there is a problem, or a normal renew, as said above you will be able to get notified
by mail or by rocket.  
Also, a log file will be filled with all the information that return the program.

* With the Rate Limits, to avoid getting blocked, you can have as many domains you want, with a maximum of 200 sites per domains.
* These 200 sites per domain will be spread over the last month of validity of the certificate to comply with these rate limits.
* From this last month, all renew would be split by 50 renew per week to respect those rate limits.

###Usage
#### Using a configuration file
If you want to create a configuration file, you can use [Viper](https://github.com/spf13/viper#putting-values-into-viper) to read,
and fill this structure by Unmarshalling the config file. The `mapstructure` will allow you to read all the configuration file formats.

```go
type Config struct {
	CertManager    certificate_manager.CertManagerConfig          `mapstructure:"certificate_manager"`
	DNSServers      []dns.DNSServerConfig                         `mapstructure:"dns_servers"`
	Sites           []certificate_prober.Config                   `mapstructure:"sites"`
	Updaters        []certificate_updater.CertificateUpdateConfig `mapstructure:"updaters"`
	Notifiers       []notification_service.NotifierConfig         `mapstructure:"notifiers"`
	LetsEncryptUser lets_encrypt.LetsEncryptUserConfig            `mapstructure:"lets_encrypt_user"`
	CertRootPath    string                                        `mapstructure:"certificates_root_path"`
	RestartMinutes  int64                                         `mapstructure:"loop_restart_min"`
}
```

Here is an example of a configuration file in json format with everything this program needs.

```json
{
  "loop_restart_min": 1440,
  "certificates_root_path": "/etc/ssl-alert-renew/letsencrypt/certificates",
  "certificate_manager": {
    "recipients": [
      {
        "notifier": "rocket-example",
        "categories": [
          "RENEW",
          "ERROR"
        ],
        "dest": [
          "@example.example"
        ]
      },
      {
        "notifier": "gmail-example",
        "categories": [
          "RENEW",
          "ERROR"
        ],
        "dest": [
          "example@example.fr"
        ]
      }
    ]
  },
  "dns_servers": [
    {
      "name": "Example Serv",
      "type": "pdns",
      "url": "http://0.0.0.0:8080",
      "api_key": "api key",
      "server_id": "localhost"
    }
  ],
  "sites": [
    {
      "server": "example",
      "url": "example.re",
      "port": 443,
      "location": {
        "certificate": "/etc/letsencrypt/live/blah.example.re/fullchain.pem",
        "private_key": "/etc/letsencrypt/live/blah.example.re/privkey.pem"
      }
    },
    {
      "server": "example",
      "url": "www.example.fr",
      "port": 443,
      "location": {
       "certificate": "path/to/current/letsencrypt/certif/www.example.fr/fullchain.pem",
       "private_key": "path/to/current/letsencrypt/certif/www.example.fr/privkey.pem"
      }
    }
  ],
  "updaters": [
    {
      "name": "example",
      "type": "remote",
      "certificates_owner": "root",
      "remote_connection": {
        "protocol": "SSH",
        "port": "22",
        "hostname": "0.0.0.0"
      },
      "reload_cmd": "systemctl reload nginx"
    },
    {
      "name": "example",
      "type": "local",
      "certificates_owner": "root",
      "restart_cmd": "systemctl reload nginx"
    }
  ],
  "notifiers": [
    {
      "name": "gmail-example",
      "type": "mail",
      "source": {
        "from": "certif.expire@gmail.com",
        "pwd": "secret pwd"
      },
      "host": "smtp.gmail.com",
      "port": 587,
      "tls": true,
      "debug": false
    },
    {
      "name": "rocket-example",
      "type": "rocket",
      "source": {
        "from": "example@example.com",
        "pwd": "secret_pwd"
      },
      "host": "rocket.example.io",
      "port": 443,
      "tls": true,
      "debug": false
    }
  ],
  "lets_encrypt_user": {
    "mail": "example@gmail.com",
    "account_path": "/etc/ssl-alert-renew/letsencrypt/account"
  }
}
```

