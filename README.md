[![Build Status](https://travis-ci.com/Oppodelldog/docker-dns.svg?branch=master)](https://travis-ci.com/Oppodelldog/docker-dns)
# docker-dns
> dns container for custom aliases

**Problem**
* You maintain a lot of domains, steadily growing.
* Your docker-compose.yaml already has way too many aliases for your LB.
* You want to update the dns list automatically without patching the compose file.

**Solution**
* Write a DNS Server that sits in your compose environment, and enables you to define mappings
from ```domain alias``` to ```container name```.

This is what this experiment was about. And it worked.

### Configuration
Since it's an experiment there is not much config options.
>Define the aliases in **dnsserver/data/alias**

The rest should be obvious from docker-compose.yaml or the go code.


### Restrictions

**Restrictions for the docker-compose setup:**
* your need to give the DNS service a static ip
* so you need a network as well


### Tests
Since this project was a just a quick try, there are no unit tests yet.
But to hold the code stable there is two functional test beds, for docker and docker-compose.
* [functional tests](test/README.md)

### Go simply rocks
**It was fun to implement using go and those nice libraries:**

DNS: htts://github.com/miekg/dns

Docker: https://github.com/docker/docker



