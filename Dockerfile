# ---- Build Stage ----
FROM golang:1.26.8-bookworm AS builder

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod & go.sum dulu biar layer cache kepakai kalau dependency gak berubah
    COPY go.mod go.sum ./
    RUN go mod download

    # cp source code
    COPY . .

# Build binary statis (CGO_ENABLED=0 biar gak depend ke libc)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server .

# ---- Runtime Stage ----
FROM ubuntu:24.04

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user (Debian/Ubuntu syntax)
RUN groupadd --system app \
    && useradd --system --gid app --no-create-home app

WORKDIR /app

COPY --from=builder --chown=app:app /app/server .

USER app

EXPOSE 8005

CMD ["./server"]