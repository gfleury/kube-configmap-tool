FROM golang:latest 
RUN mkdir -p /go/src/app 
ADD . /go/src/app/ 
WORKDIR /go/src/app
RUN echo $GOPATH
RUN go build .