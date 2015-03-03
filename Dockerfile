FROM quay.io/geonet/golang-godep:latest

COPY . /go/src/github.com/GeoNet/geonet-rest

WORKDIR /go/src/github.com/GeoNet/geonet-rest

RUN godep go install -a

EXPOSE 8080

CMD ["/go/bin/geonet-rest"]
