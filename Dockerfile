FROM golang:1.24 as build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kahoot

FROM alpine
WORKDIR /app
COPY --from=build /app .
EXPOSE 8001
CMD ["/app/kahoot"]
