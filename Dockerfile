FROM golang:1.15
WORKDIR /go/src/puzzle_helper
COPY . .
RUN go build
