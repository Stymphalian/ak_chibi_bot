FROM golang:1.22.3

# Install npm and typescript
RUN apt-get update
RUN apt-get -y install zip
RUN mkdir -p /node
RUN cd /node
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
RUN export NVM_DIR="$HOME/.nvm" && \. $NVM_DIR/nvm.sh && nvm install node && npm install typescript -g

# Copy over the source and program
WORKDIR /work
RUN mkdir /ak_chibi_assets
COPY go.mod go.sum ./
COPY server/ ./server
COPY ./assets/assets /ak_chibi_assets/assets
COPY ./assets/spine-ts/build /ak_chibi_assets/spine-ts/build
COPY ./assets/spine-ts/css /ak_chibi_assets/spine-ts/css
COPY ./assets/spine-ts/static /ak_chibi_assets/spine-ts/static
COPY ./assets/spine-ts/favicon.ico /ak_chibi_assets/spine-ts/favicon.ico
COPY ./assets/spine-ts/index.html /ak_chibi_assets/spine-ts/index.html
COPY ./assets/admin /ak_chibi_assets/admin

# BUILD
RUN go mod download && go mod verify
RUN go build -o ak_chibi_bot server/main.go
EXPOSE 8080