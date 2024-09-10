# Overview 
This is a twitch bot to allow viewers to play with an Arknights chibi on the screen by 
issuing `!chibi` commands in chat. You can use a hosted bot `StymTwitchBot` or you
can run the bot locally on your machine so that all network requests only happen
on your own machine.

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
!chibi walk <?number> | Have the chibi walk back and forth across the screen. Optionally specify a number between 0 and 1.0 for the chibi to walk to that part of the screen.
!chibi enemy The Last Steam Knight | Change the chibi into one of the enemies in the game. You can also use the enemies "code". `!chibi enemy B4` is equivalent to `!chibi enemy Originium Slug`. You can find all the enemy codes from [Arknights Wiki](https://arknights.wiki.gg/wiki/Enemy/Normal)
!chibi speed 2.0 | change the speed in which an animation is played (min 0.1, max 3.0)
!chibi who Rock | Ask the bot for operators/enemies which match the given name.

# Quick Start - Hosted
Have the twich bot directly connect to your chat by using a hosted bot.
1. Open up OBS and add a `Browser` source to your stream.
2. Set the URL to `https://jellyfish-app-gseou.ondigitalocean.app/room/?channelName=REPLACE_ME`
3. Change the `REPLACE_ME` part of the channelName (lowercase) to your own channel (ie. `?channelName=stymphalian2__`). 
4. Set the width and height to 1920x1080
5. A chibi should now be walking around.
6. Open up your twitch chat and start typing in commands

# Quick Start - Local
Run the twitch bot locally on your own machine and attached to your own twitch account.

1. Download this repository and open up a window to it
2. Open up a command-prompt/terminal window in that directory (i.e C:/Users/Stymphalian/Downloads/ak_chibi_bot/)
by right clicking and selecting (open in command prompt/terminal).
3. Unzip the `release.zip` folder into a `release` folder.
4. Download the chibi assets files from [here](https://f002.backblazeb2.com/file/ak-gamedata/assets_20240805.zip) and unzip into the `releases` folder with the name `assets`.
5. You now need register the bot to your channel. Follow the instructions from [Authentication](#Authentication)
6. Open up the `config.json` file in a text-editor. Update the `twitch_bot` field to your channel name, `twitch_client_id` to the clientId you registered your application with, and set the `twitch_access_token` field to what you created in the previous step.
7. Run the server binary (`.\server_win386.exe -image_assetdir assets -spine_assetdir spine-ts -admin_assetdir admin -address localhost:8080 -bot_config config.json`)
8. Open up OBS and add a `Browser` source to your stream.
9. Set the URL to http://localhost:8080/
10. Set the width and height to 1920x1080
11. A chibi should now be walking around.
12. Open up your twitch chat and start typing in commands


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
Set the Oauth Redirect URL to `http://localhost:8080`

You will now need an OAUTH access token in order to connect to the Twitch IRC.
Again follow the twitch [documentation](https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/#implicit-grant-flow )
You can do also quickly get one by going to this URL:

`https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=<your client id>&redirect_uri=http://localhost:8080&scope=chat%3Aread+chat%3Aedit`

Replace the `<your client id>` with the client ID that you got when registering your bot.
Also make sure you have a `localhost:8080` server running (`python -m http.server 8080`)
so that the redirect of the OAUTH has somewhere to go. You can also run step 7 in the [Quickstart](#Quick-Start-Local)

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
docker run -it --rm -v .:/work -p 8080:8080 ak-chibi-bot
```

Once within the container:
```
/work
cd /work/server
go run main.go -spine_assetdir=../assets/spine-ts -address=:8080 -bot_config=config.json
# This will start the HTTP server and connect to the twitch IRC
```

To hot-reload the spine-runtime typescript library:
```
/work
cd /work/server/spine-ts
tsc -w -p tsconfig.json
```

Now open http://localhost:8080/room/?channelName=REPLACE_ME in your web-browser.
Replce the `REPLACE_ME` with your own twitch channel name and you should see 
default operator walking around.

### Releasing
To build a release binary run the following within the container `./tools/build_release.sh`

# Disclaimer
1. All the art assets/chibis are owned by Arknights/Hypergryph. I claim no ownership and
the use of their assets are purely for personal enjoyment and entertainment.
2. The software used for rendering the chibis (i.e Spine models) use the Esoteric 
runtime libraries which is under [License](http://esotericsoftware.com/spine-editor-license). 
Strictly speaking use of this software requires each individual to have purchased
their own software license.