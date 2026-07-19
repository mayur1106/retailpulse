FROM golang:1.26-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum* ./
RUN go mod download
COPY apps/api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/retailpulse-api ./cmd/api

FROM alpine:3.22
RUN adduser -D -H retailpulse
WORKDIR /app
COPY --from=build /out/retailpulse-api /app/retailpulse-api
COPY apps/api/migrations /app/migrations
USER retailpulse
EXPOSE 4005
ENTRYPOINT ["/app/retailpulse-api"]
