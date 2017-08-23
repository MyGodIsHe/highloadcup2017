#FROM centos:7
#WORKDIR /root
#ADD golang golang
#EXPOSE 80
#CMD ./golang

FROM centos:7
WORKDIR /root
RUN yum install -y wget && \
    wget https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.8.3.linux-amd64.tar.gz && \
    mkdir go && mkdir go/src && mkdir go/bin && mkdir go/pkg && \
    mkdir go/src/dumb
RUN yum install -y git
ENV PATH=${PATH}:/usr/local/go/bin GOROOT=/usr/local/go GOPATH=/root/go
ADD *.go go/src/traveler/
RUN go get traveler && go build traveler && go install traveler
EXPOSE 80
CMD ./go/bin/traveler
