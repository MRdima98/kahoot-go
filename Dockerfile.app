FROM golang:1.24

RUN go install github.com/air-verse/air@latest
WORKDIR /kahoot
RUN git config --global --add safe.directory /kahoot
COPY go.mod go.sum ./
COPY go.mod go.sum ./
RUN go mod download
COPY . .

EXPOSE 8001
CMD ["air -c .air.toml"]
