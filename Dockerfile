# syntax=docker/dockerfile:1

FROM golang:1.19

WORKDIR C:\Users\zachs\rl github\RL-Discord-Matchmaking\src

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

CMD ["/docker-gs-ping"]