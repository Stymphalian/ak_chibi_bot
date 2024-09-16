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
COPY ./static /ak_chibi_assets
RUN rm -rf /ak_chibi_assets/spine/src

# BUILD
RUN go mod download && go mod verify
RUN go build -o ak_chibi_bot server/main.go
EXPOSE 8080
# CMD ["go", "run",  "server/main.go", "-address=:8080", "-image_assetdir=/ak_chibi_assets/assets", "-static_dir=./static", "-bot_config=/work/server/config.json"]