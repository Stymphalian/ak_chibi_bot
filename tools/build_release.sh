#!/bin/sh
## This script is obsolete. I need to figure out how to just run the bot server
## without using a database.
exit 0
cd /work
mkdir -p release

# Build the server
cd /work/server
env GOOS=windows GOARCH=386 go build -o server_win386.exe
mv /work/server/server_win386.exe /work/release/

# Build the spine runtime
cd /work/static/spine
tsc -p tsconfig.json

# Copy over the files to the release folder
mkdir -p /work/release/static
mkdir -p /work/release/static/spine
cp -r /work/static/admin /work/release/static/admin
cp -r /work/static/fonts /work/release/static/fonts
cp -r /work/static/spine/build /work/release/static/spine/build
cp -r /work/static/spine/css /work/release/static/spine/css
cp /work/static/spine/index.html /work/release/static/spine/index.html
cp /work/static/spine/LICENSE /work/release/static/spine/LICENSE
cp /work/static/favicon.ico /work/release/static/favicon.ico

## Copy over the config
cp /work/tools/default_config.json /work/release/config.json

# zip release folder into release.zip
cd /work
rm release.zip
zip -r release.zip release
rm -rf /work/release/
