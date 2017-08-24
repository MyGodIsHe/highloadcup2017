#FROM centos:7
#WORKDIR /root
#ADD golang golang
#EXPOSE 80
#CMD ./golang

FROM golang:1.9
WORKDIR /root
ADD *.go $GOPATH/src/traveler/
RUN go get traveler && go build traveler && go install traveler
EXPOSE 80
CMD $GOPATH/bin/traveler
