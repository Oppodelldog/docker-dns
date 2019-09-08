module github.com/Oppodelldog/docker-dns

go 1.13

require (
	github.com/Oppodelldog/dockertest v0.0.2 // indirect
	github.com/Oppodelldog/filediscovery v0.3.0
	github.com/docker/docker v1.13.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/miekg/dns v1.1.16
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/crypto v0.0.0-20190907121410-71b5226ff739 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190907184412-d223b2b6db03 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/Oppodelldog/dockertest v0.0.1 => ../dockertest
