# docker-dns
### Functional tests
Functional tests currently must me checked manually. (see test/...)

The test script and the docker-compose environment simulate the same behavior:

* **dnstester** is asking for custom domain names that could not be resolved by the docker built-in dns facility
* sice the **pong** is discovered by the **dnsserver**
* **dnsserver** will answer all dns requests coming from **dnstester**

Execute tests
```bash
make functional-tests
```
exit code will be 0 on success