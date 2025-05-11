# Setup
FROM golang:1.24
RUN go install github.com/air-verse/air@latest
WORKDIR /kahoot
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go run ./db/db.go


# Execute
EXPOSE 8080
CMD ["air"]
