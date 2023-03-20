FROM golang:1.20.2-buster AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o=/bin/api ./cmd/api

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build /bin/api /api

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]