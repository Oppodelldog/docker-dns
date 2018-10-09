# docker-dns
### Functional tests
Functional tests currently must me checked manually. (see test/...)

The test script and the docker-compose environment simulate the same behavior:

* **dnstester** is asking for custom domain names that could not be resolved by the docker built-in dns facility
* sice the **pong** is discovered by the **dnsserver**
* **dnsserver** will answer all dns requests coming from **dnstester**

The expected output are some valid DNS lookups for the strange domains. sth like:
``` bash
> test/functional-test-docker.sh

pong. IN A 172.17.0.3
ponge.longe.long.com. IN A 172.17.0.3
www.pong.com. IN A 172.17.0.3
```

In docker-compose you should find the same output when following the output of **dnstester**.
```
cd test
docker-compose up
```