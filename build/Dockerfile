FROM golang:1.22-alpine as builder

# Download and cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/cache go mod download

# Add source code & build
COPY . .
RUN go build -o /go/bin/infro ./cmd/main.go

FROM golang:1.22-alpine

COPY --from=builder /go/bin/infro /go/bin/infro

ENTRYPOINT [ "bin/infro" ]