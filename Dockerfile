FROM golang:1.22.3

# Install npm and typescript
RUN apt-get update
RUN apt-get -y install zip
RUN mkdir -p /node
RUN cd /node
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
RUN export NVM_DIR="$HOME/.nvm" && \. $NVM_DIR/nvm.sh && nvm install node && npm install typescript -g

WORKDIR /work
COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN mkdir /ak_chibi_assets
COPY ./server/assets /ak_chibi_assets/assets

EXPOSE 7001 