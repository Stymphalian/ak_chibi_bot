# Developer Documentation

There are 4 major components to the twitch bot.
1. The Twitch IRC client which connects to the Twitch chat channel.
2. The Chibi module which provides an abstraction for dealing with rendering the
 Chibis.
3. The Spine Runtime (web/typescript) which renders and plays the chibi/spine 
animations in a web browser.
4. The Spine Bridge which parses the `!chibi` commands from chat IRC messages and
forwards instructions to the Runtime via websockets. It acts as a client between
`Chibi` component and the `Runtime`.
5. The HTTP [Server](server/main.go) which glues everything together. It is the
entry point of the program. It connects the Twitch IRC client, and also serves
the web files required to show the Browser spine runtime webpage. It also contains
the spine Bridge to facilitate communication between the `Chibi` and 
the `Runtime` components.


## Arknights Assets
The AK chibi models use a program called `Spine`.
The runtime requires 3 files in order to render them (`.skel`, `.atlas` and the spritesheet `.png`).
The files need to be put under an `assets/` directory which is accessible to the 
HTTP server for serving. The structure is very specific:
```
assets/
  characters/
    char_002_amiya/default/base/Front/
        char_002_amiya.animations.json
        char_002_amiya.atlas
        char_002_amiya.png
        char_002_amiya.skel
    char_002_amiya/epoque#4/battle/Front/...
    char_002_amiya/winter#1/battle/Back/...
    ...
  enemies/
    enemy_1000_gopro/default/battle/Front/
    ...
saved_enemy_names.json
saved_names.json
```

The arknights assets are organized such that we can easily identify
key parts of the chibi we want.
1. The Operator's Name/ID (char_002_amiya)
2. The Skin which the opeators have (default, epoque#4, winter#1, etc)
3. The chibi "stance" which can be either "base" or "battle". Base is the model which 
shows up when your OPs are working in the factories or wandering the dorms. 
"battle" stance are the models which appears in the playable maps of the game. 
4. Front and Back is the facing direction (i.e when deploying an operator facing up
they would be showing you their back instead of their face). Only "battle" chibis
have a Back facing. Everything else should be put into the
"Front" directory.

The `saved_names.json` files are used for fuzzy matching on the operators
names when issuing `!chibi` commands when selecting operators.
For example `Młynar` would be very hard to type due to the non-ascii 
character, so we have some alternative spellings like `Mlynar` 
which can be matched against to resolve the name.

These asset files are not widely distributed. You can find many github repositories
which have done this work for you, but not all of them are up to date.
I got these assets myself through other means, but in order to not bloat
the repo I have not included them here.
Please find a zip [here](https://f002.backblazeb2.com/file/ak-gamedata/assets_20240610.zip) which I will try to keep up to date.