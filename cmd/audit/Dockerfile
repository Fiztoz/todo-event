# syntax=docker/dockerfile:1.6
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/audit ./cmd/audit

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/audit /audit
USER nonroot:nonroot
ENTRYPOINT ["/audit"]
