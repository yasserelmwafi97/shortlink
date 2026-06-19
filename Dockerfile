FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/shortlink ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/shortlink /app/shortlink
ENV PORT=8080 DB_PATH=/data/shortlink.db
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/shortlink"]
