FROM golang:1.11.1

WORKDIR /go/src/github.com/Oppodelldog/docker-dns

RUN go get -u github.com/golang/dep/cmd/dep && \
	go get -u golang.org/x/tools/cmd/goimports

# gometalinter
RUN cd $GOPATH && wget -O- https://git.io/vp6lP | sh

RUN wget -O codecov https://codecov.io/bash && chmod ugo+x codecov && cp codecov $GOPATH/bin/.
    
RUN chmod -R a+rw $GOPATH/src
    