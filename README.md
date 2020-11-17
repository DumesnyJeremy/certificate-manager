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

## Usage

### Linux Installation
```yaml
Create the executable:
$ go build -mod=vendor
```

```yaml
Set up the Config directory and copy the binary: 
$ sudo ./setup-prog.sh -i
```

Go to `$ cd /etc/certificate-manager`, you will found 3 `$ config."extension".sample`,
choose the one you want and rename the file `$ sudo mv config."Extension".sample config."Extension"` 
and start filling it with your information.

You can try it by executing the program. This is what happen when you have a site certificate to renew. 
The communication will be established, and the DNS-01 challenge will be resolved.
```dockerfile
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
```yaml
Set up the timer: 
$ sudo ./setup-prog.sh -t

Enable the timer: 
$ systemctl enable certificate-manager.timer

Restart the timer: 
$ systemctl restart certificate-manager.timer

Check timer status: 
$ systemctl status certificate-manager.timer
```

#### Use the Program as a Daemon
```yaml
Set up the daemon: 
$ sudo ./setup-prog.sh -d

Start the daemon: 
$ systemctl start certificate-manager

Check daemon status: 
$ systemctl status certificate-manager
```
#### Delete configuration files and uninstall Timer/Daemon
```yaml
$ sudo ./setup-prog.sh -p -u
```
