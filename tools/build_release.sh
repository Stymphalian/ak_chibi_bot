#!/bin/sh
cd /work
mkdir -p release
mkdir -p release/spine-ts

# Build the server
cd /work/server
env GOOS=windows GOARCH=386 go build -o server_win386.exe
mv /work/server/server_win386.exe /work/release/

# Copy over the spine data
cd /work/server/spine-ts
tsc -p tsconfig.json
cp -r /work/server/spine-ts/build /work/release/spine-ts
cp -r /work/server/spine-ts/css /work/release/spine-ts
cp -r /work/server/spine-ts/static /work/release/spine-ts
cp /work/server/favicon.ico /work/release/spine-ts/favicon.ico
cp /work/server/spine-ts/index.html /work/release/spine-ts/index.html

## Copy over the config
cp /work/tools/default_config.json /work/release/config.json

# zip release folder into release.zip
cd /work
zip -r release.zip release
rm -rf /work/release/
