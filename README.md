# docker-dns
> dns container for custom aliases

**Problem**
* You maintain a lot of domains, stedily growing.
* Your docker-compose.yaml already has way too many aliases for your LB.
* You want to update the dns list automatically wihtout patching the compose file.

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

### Go simply rocks
**It was fun to implement using go and those nice libraries:**

DNS: htts://github.com/miekg/dns

Docker: https://github.com/docker/docker



