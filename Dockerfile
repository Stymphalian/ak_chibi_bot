## BASE
FROM golang:1.23 AS base
RUN mkdir -p /ak_chibi_assets
RUN mkdir -p /ak_chibi_assets/assets
RUN mkdir -p /ak_chibi_assets/public
RUN mkdir -p /ak_chibi_assets/spine/dist
RUN mkdir -p /ak_chibi_assets/web_app/build
COPY ./static/assets /ak_chibi_assets/assets
COPY ./static/public /ak_chibi_assets/public
COPY ./static/spine/dist /ak_chibi_assets/spine/dist
COPY ./static/web_app/build /ak_chibi_assets/web_app/build

## DEVELOPMENT
## ----------------
FROM base AS development

# Install npm and typescript
# RUN apt-get update
# RUN apt-get -y install zip
# RUN mkdir -p /node
# RUN cd /node
# RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
# RUN export NVM_DIR="$HOME/.nvm" && \. $NVM_DIR/nvm.sh && nvm install node && npm install typescript -g
RUN go install github.com/air-verse/air@6403f4d1e069e4a6eeb49639c8cafb168c28a523

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
CMD ["air"]

## DEBUG
## -----------------
FROM development as debug
WORKDIR /app
RUN CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
CMD ["air", "-c", ".air-debug.toml"]

## BUILDER
# docker build -t ak-chibi-bot --target builder .
## ----------------
FROM base AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY server/ ./server
RUN CGO_ENABLED=0 go build -o /build/ak_chibi_bot server/main.go

## PRODUCTION
# docker build -t ak-chibi-bot:prod --target production .
## ---------------
FROM alpine:3.20 AS production
WORKDIR /prod
COPY --from=builder /build/ak_chibi_bot ./ak_chibi_bot
COPY --from=builder /ak_chibi_assets ./static
COPY secrets/prod_config.json ./config.json
EXPOSE 8080

RUN addgroup -S botgroup && adduser -S botuser -G botgroup
USER botuser
ENTRYPOINT [ \
    "./ak_chibi_bot", \
    "-address=:8080", \
    "-image_assetdir=static/assets", \
    "-static_dir=static", \
    "-bot_config=config.json" \
]