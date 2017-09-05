FROM golang:1.9
WORKDIR /root
ADD *.go $GOPATH/src/traveler/
RUN go get traveler && CGO_ENABLED=0 GOOS=linux go build traveler && go install traveler
EXPOSE 80
CMD $GOPATH/bin/traveler
