# certificate-manager

This is a Go program that uses the [lets-encrypt](github.com/DumesnyJeremy/lets-encrypt) library and the
[notification-service](github.com/DumesnyJeremy/notification-service) library.

This is a Go program, which aims to check the number of remaining days for sites certificate 'given' and, renew them automatically.
Certificates will be renewed during the last month of validity, by accomplishing the DNS-01 challenge. 
It will prove to Let's Encrypt that you control the DNS of this domain.
The program will decide if we can renew or not the site certificate, in function of the [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/) .  
If this domain doesn't respect one of the limits, it will be discard, and you will receive an alert by **mail** or **rocket**,
who will explains to you the reason why it has been deleted. Otherwise, the renewal will take place normally, and you will be notified
of the smooth running of this renewal.   
Also, a log file will be filled with all the information that returns the program.

* With the Rate Limits, to avoid getting blocked, you can have as many domains as you want, with a maximum of 200 sites per domain.
* Those 200 sites will be spread over their last month of validity to comply with the rate limits.
* During this last month, all renewals would be split by 50 renewals per week to respect the limits.

## Usage

### Dependencies
Goland [v1.14](https://golang.org/doc/go1.14) or above is required. To install it in your system, follow the [Official doc](https://golang.org/doc/install).
The project uses [go vendoring mod](https://golang.org/ref/mod#go-mod-vendor) (aka .vgo) for dependencies management. 

### Linux Installation
```shell script
git clone https://github.com/DumesnyJeremy/certificate-manager.git
cd certificate-manager
go mod vendor
go build -mod vendor
./install.sh -h
```

```yaml
#Set up the Config directory and copy the binary: 
sudo ./install.sh -i
```

Go to `$ cd /etc/certificate-manager`, you will found 3 `$ config."extension".sample`,
choose the one you want and rename the file `$ sudo mv config."Extension".sample config."Extension"` 
and start filling it with your information.

You can try it by executing the program. This is what happen when you have a site certificate to renew. 
The communication will be established, and the DNS-01 challenge will be resolved.
```shell script
INFO[17/11/2020 17:44:22] [INFO] acme: Registering account for example@gmail.com 
INFO[17/11/2020 17:44:24] [INFO] [www.example.com] acme: Obtaining bundled SAN certificate 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] AuthURL: https://acme-staging-v02.api.letsencrypt.org/acme/authz-v3/1234567 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] acme: Could not find solver for: tls-alpn-01 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] acme: Could not find solver for: http-01 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] acme: use dns-01 solver 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] acme: Preparing to solve DNS-01 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] acme: Trying to solve DNS-01 
INFO[17/11/2020 17:44:25] [INFO] [www.example.com] acme: Checking DNS record propagation using [192.0.0.0:53 0.0.0.0:53 0.0.0.0:53 0.0.0.0:53 0.0.0:53] 
INFO[17/11/2020 17:44:27] [INFO] Wait for propagation [timeout: 1m0s, interval: 2s] 
INFO[17/11/2020 17:44:33] [INFO] [www.example.com] The server validated our request 
INFO[17/11/2020 17:44:33] [INFO] [www.example.com] acme: Cleaning DNS-01 challenge 
INFO[17/11/2020 17:44:33] [INFO] [www.example.com] acme: Validations succeeded; requesting certificates 
INFO[17/11/2020 17:44:34] [INFO] [www.example.com] Server responded with a certificate. 
INFO[17/11/2020 17:44:34] [www.example.com] New certificate upload; Rocket to @example 
INFO[17/11/2020 17:44:36] [www.example.com] New certificate upload; Mail to example@gmail.fr 
INFO[17/11/2020 17:44:36] [www.example.com] Force Renew; Rocket to @example 
INFO[17/11/2020 17:44:37] [www.example.com] Force Renew; Mail to example@gmail.fr 
```

#### Use the Program as a Timer
The **timer** is used, to run the program every day at a given time. As it happens, here the launch will take place at 01:00 AM.
If you want to change it, go to: `/etc/systemd/system/certificate-manager.timer`

```yaml
#Set up the timer:
sudo ./setup-prog.sh -t

#Enable the timer: 
systemctl enable certificate-manager.timer

#Restart the timer: 
systemctl restart certificate-manager.timer

#Check timer status: 
systemctl status certificate-manager.timer

#Stop the timer:
systemctl stop certificate-manager.timer
```

#### Use the Program as a Daemon
```yaml
#Set up the daemon: 
sudo ./install.sh -d

#Start the daemon: 
systemctl start certificate-manager

#Check daemon status: 
systemctl status certificate-manager

#Stop the daemon:
systemctl stop certificate-manager

```
#### Delete configuration files and Uninstall the Timer or Daemon
```yaml
#Disable and stop the Daemon or the Timer:
sudo ./install.sh -u

#Delete the certificate-manager file in /etc/:
sudo ./install.sh -p
```
