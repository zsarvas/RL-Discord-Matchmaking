FROM golang:1.19
WORKDIR C:\Users\zachs\rl_github\RL-Discord-Matchmaking\src\
COPY go.mod go.sum ./
COPY . ./
RUN go build -o main main.go
CMD ["./main"]