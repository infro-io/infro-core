FROM golang:1.22-alpine as builder

# Download and cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/cache go mod download

# install argocd cli
RUN --mount=type=cache,target=/go/cache wget -O /go/bin/argocd https://github.com/argoproj/argo-cd/releases/download/v2.10.4/argocd-linux-amd64
RUN chmod +x /go/bin/argocd

# Add source code & build
COPY . .
RUN go build -o /go/bin/infro ./cmd/main.go

FROM golang:1.22-alpine

COPY --from=builder /go/bin/infro /go/bin/infro
COPY --from=builder /go/bin/argocd /go/bin/argocd

ENTRYPOINT [ "infro" ]