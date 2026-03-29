# ---- Stage 1: Build frontend ----
FROM node:22-alpine AS frontend

WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml* ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# ---- Stage 2: Build Go binary ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Copy frontend build output into the embed directory
COPY --from=frontend /app/frontend/dist ./static/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o /kiln ./cmd/kiln

# ---- Stage 3: Production image ----
FROM alpine:3.19

RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    git \
    unzip

# Install tfenv for managing terraform versions
RUN git clone --depth 1 https://github.com/tfutils/tfenv.git /root/.tfenv
ENV PATH="/root/.tfenv/bin:${PATH}"

# TODO(pulumi): Uncomment when Pulumi support is implemented.
# RUN curl -fsSL https://get.pulumi.com | sh
# ENV PATH="/root/.pulumi/bin:${PATH}"

COPY --from=builder /kiln /usr/local/bin/kiln

EXPOSE 8080

ENTRYPOINT ["kiln"]
