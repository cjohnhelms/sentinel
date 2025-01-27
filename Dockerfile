FROM golang:1.24rc2-alpine3.21 AS build

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o sentinel cmd/main.go

FROM golang:1.22.11-alpine3.21 AS final

WORKDIR /app

COPY --from=BUILD /app/sentinel .

EXPOSE 2112

CMD [ "/app/sentinel" ]

