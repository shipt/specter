FROM golang:1.12.4 AS builder 
COPY . /go/src/github.com/shipt/specter/
WORKDIR /go/src/github.com/shipt/specter/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -v -o specter-data -a -installsuffix cgo ./cmd/specter-data/main.go

FROM alpine
COPY --from=builder /go/src/github.com/shipt/specter/specter-data /app/
RUN apk update
RUN apk upgrade
RUN apk add ca-certificates
WORKDIR /app
CMD ["./specter-data"]
