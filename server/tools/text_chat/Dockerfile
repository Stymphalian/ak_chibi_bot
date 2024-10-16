## BASE
FROM golang:1.23 AS base

## DEVELOPMENT
## ----------------
FROM base AS development

RUN go install github.com/air-verse/air@6403f4d1e069e4a6eeb49639c8cafb168c28a523
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
CMD ["air", "-c", "server/tools/text_chat/.air-cli.toml"]