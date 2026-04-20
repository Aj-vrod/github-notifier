FROM golang:1.26.2-alpine3.23

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY / /app/

RUN CGO_ENABLED=0 GOOS=linux go build -o /out

EXPOSE 8001

CMD [ "/out", "http" ]
