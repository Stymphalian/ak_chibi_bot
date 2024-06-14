#!/bin/sh

cd /work/server
env GOOS=windows GOARCH=386 go build -o server_win386.exe
cd /work
mkdir -p release/
mkdir -p release/spine-ts
mv /work/server/server_win386.exe release/
cp -r /work/server/spine-ts/build release/spine-ts
cp -r /work/server/spine-ts/player release/spine-ts

# zip release folder into release.zip
zip -r release.zip release
rm -rf /work/release/
