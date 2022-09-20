FROM golang:latest AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o /ci-server


#FROM alpine:latest
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=builder /ci-server /ci-server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/ci-server"]
