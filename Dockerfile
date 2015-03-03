FROM quay.io/geonet/golang-godep:latest

RUN apt-get update 

RUN go get github.com/GeoNet/geonet-rest

WORKDIR /go/src/github.com/GeoNet/geonet-rest

RUN godep go install

EXPOSE 8080

CMD ["/go/bin/geonet-rest"]
