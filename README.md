# Overview 
This is a twitch bot to allow viewers to play with an Arknights chibi on the screen by 
issuing `!chibi` commands in chat. The bot runs locally on your machine meaning 
all network requests are local to your computer except to/from the twitch chat.

![Demo Image](readme_assets/demo1.png)

# Example Commands
Command Name | Description
-----------|-----------------------
!chibi help | Displays a help message in chat for how to use the bot
!chibi Amiya | Changes the User's current chibi to Amiya. The bot does fuzzy matching on the operator's names.
!chibi skins | Lists the available skins for the chibi. `default` is the operators normal e0 skin. Other available skins are identified by the fashion brand (i.e epoque, winter, etc)
!chibi anims | Lists the available animations for the chibi. Normal idle animations are `Relax`, `Idle`, `Sit` and `Move`. Operators with skins also have a `Special` animation that can be played.
!chibi face <front,back> | Only works in battle stance. Make the chibi face forward or backwards.
!chibi stance <base,battle> | Set the chibi to their base model (ie. the one which walk around the factory/dorms) or the battle map model
!chibi skin epoque | Change the skin to `epoque` or any other skin name.
!chibi play Attack | Change the animation the chibi currently plays. The animation loops forever.
!chibi walk | Have the chibi walk back and forth across the screen.
!chibi enemy The Last Steam Knight | Change the chibi into one of the enemies in the game.

# Quick Start
1. Download this repository and open up a window to it
2. Open up a command-prompt/terminal window in that directory (i.e C:/Users/Stymphalian/Downloads/ak_chibi_bot/)
by right clicking and selecting (open in command prompt/terminal).
3. Go to the `release/` folder
4. Download the chibi assets files from [here](https://f002.backblazeb2.com/file/ak-gamedata/assets_20240610.zip) and unzip into the `releases` folder with the name `assets`.
5. You now need register the bot to your channel. Follow the instructions from [Authentication](#Authentication)
5. Open up the `config.json` file in a text-editor. Update the `broadcaster` name, `channelName` and `twitch_access_token` fields.
4. Run the server binary (`server_win386.exe -assetdir=assets -address=localhost:7001 -twitch_config=config.json`)
6. Open up OBS and add a `Browser` source to your stream.
7. Set the URL to http://localhost:7001/player/example
8. Set the width and height to 1920x1080
6. A chibi should now be walking around.
7. Open up your twitch chat and start typing in commands


# For Developers
I develop using Docker, but as long as you have golang installed on your computer
you can run this locally as well. 
For some commentary on the architecture and layout of the program see [README_DEV.md](README_DEV.md)

### Downloading
```
git clone https://github.com/Stymphalian/ak_chibi_bot
```

### Authentication
To connect to your twitch you need to register the bot through twitch dev 
[console](https://dev.twitch.tv/console).
Please read the twitch [documentation](https://dev.twitch.tv/docs/authentication/register-app/).
Set the Oauth Redirect URL to `http://localhost:7001`

You will now need an OAUTH access token in order to connect to the Twitch IRC.
Again follow the twitch [documentation](https://dev.twitch.tv/docs/irc/authenticate-bot/)
You can do also quickly get one by going to this URL:

`https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=<your client id>&redirect_uri=http://localhost:7001&scope=chat%3Aread+chat%3Aedit`

Replace the `<your client id>` with the client ID that you got when registering your bot.

Once you get that token enter it into `twitch_access_token` field in the `config.json` file.
Also change the `broadcaster` and `channel_name` fields to match your own Name and Channel.
(i.e stymphalian__ is my username, and my channel is called Stymphalian__)
If you want to run your Bot under a special Bot twitch account just make sure
to change the `twitch_bot` field in the config to your bot's name and make sure 
to get the correct access token for your Bot and give them a moderator role in your own channel.

### Building
To build the image and start the docker container:
```
docker build -t ak-chibi-bot .
docker run -it --rm -v .:/work -p 7001:7001 ak-chibi-bot
```

Once within the container:
```
/work
cd /work/server
go run main.go -twitch_config=config.json
# This will start the HTTP server and connect to the twitch IRC
```

To hot-reload the spine-runtime typescript library:
```
/work
cd /work/server/spine-ts
tsc -w -p tsconfig.core.json
tsc -w -p tsconfig.webgl.json
tsc -w -p tsconfig.player.json
```

Now open http://localhost:7001/player/example in your web-browser and you can now 
see a chibi walking around. 

### Releasing
To build a release binary. Run the following within the container:
```
cd /work/server
env GOOS=windows GOARCH=386 go build -o server_win386.exe
cd /work
mkdir -p release/
mkdir -p release/spine-ts
mv /work/server/server_win386.exe release/
cp -r /work/server/spine-ts/build release/spine-ts
cp -r /work/server/spine-ts/player release/spine-ts
```

# Disclaimer
1. All the art assets/chibis are owned by Arknights/Hypergryph. I claim no ownership and
the use of their assets are purely for personal enjoyment and entertainment.
2. The software used for rendering the chibis (i.e Spine models) use the Esoteric 
runtime libraries which is under [License](http://esotericsoftware.com/spine-editor-license). 
Strictly speaking use of this software requires each individual to have purchased
their own software license.