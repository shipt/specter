FROM golang:1.12.4
COPY . /go/src/github.com/shipt/specter/
WORKDIR /go/src/github.com/shipt/specter/

RUN apt update
RUN apt upgrade -y

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN go build -v -o specter-data ./cmd/specter-data/main.go

CMD ["./specter-data"]