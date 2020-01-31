# docker-dns
[![Build Status](https://travis-ci.com/Oppodelldog/docker-dns.svg?branch=master)](https://travis-ci.com/Oppodelldog/docker-dns)

### Functional tests
The functional test is orchestrated using [dockertest](https://github.com/Oppodelldog/dockertest)
Implementation of the test: [main.go](main.go) 

#### Test Setup

The test is made of three docker services:

* **pong** service is doing nothing, just a stub waiting to receive a ping
* **dnstester** service is trying to connect to pong using custom dns-names.
* **dnsserver** will answer all dns requests coming from **dnstester** and return pongs ip

Execute tests
```bash
make functional-tests
```
exit code will be 0 on success