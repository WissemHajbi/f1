FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/probe ./cmd/probe

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/api /out/probe /app/
COPY data /app/data
RUN mkdir -p /app/storage && chown -R app:app /app
USER app
ENV HTTP_ADDRESS=:8080 DB_PATH=/app/storage/oidysts.db UPSTREAM_USER_AGENT=oidysts/0.1
EXPOSE 8080
CMD ["/app/api"]
