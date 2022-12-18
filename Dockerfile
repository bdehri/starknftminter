FROM golang:1.19-alpine

WORKDIR /usr/src/minter
COPY go.mod go.sum ./
RUN go mod download
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /usr/local/bin/minter main.go
ENTRYPOINT ["minter"]
