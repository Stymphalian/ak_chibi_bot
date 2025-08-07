# Developer Documentation

The AK Chibi Bot system components
Component | Description
----------| ------------
Spine Runtime | The Web/Typescript component which renders and plays the chibi spine animations in a browser window/OBS browser source.
Spine Bridge | A server-side component connects to the Spine Runtime via websockets in order to forward instructions on what chibis to display and what animations should be played.
Twitch Chat | Twitch IRC client which connects to the User's Twitch chat. This component parses the text `!chibi` commands and converts them into `Commands` to be processed by the `ChibiActor`.
Chibi Actor | The chibi component provides an abstraction for dealing with rendering Chibis. It provides methods for setting the current operator, choosing which animations are being played, etc. This component also holds the list of viewers/chatters and keeps the mapping/state between a viewer and their current chosen chibi. This component uses the spine-bridge to communicate to the User's browser (ie. Spine Runtime) for displaying the Chibi.
Rooms/Rooms Manager | A Room encapsulates a twitch chat, spine bridge/runtime and a chibi actor into a single object. Each time a Streamer creates a new Browser Source by hitting `HTTP GET /room` a new Room entry is added to the Rooms Manager. This abstraction keeps all the logic isolated between different streams/sessions.
Web App | A frontend web application avaialable at the top-level domain. It provides a front facing website for the Bot as well as provides a `/settings/` page for Streamers to customize the Bots settings (i.e min/max values for `!chibi` commands). We login via Twitch OAUTH in order to authorize changes to a channel's settings.
Database (postgres) | A database to hold persistent state for the the active rooms/chatters, as well as to store the twitch oauth tokens/session cookies. Mainly needed to allow for persistent user preferences saving, and to allow for seamless server restarts.
Bot Server | The HTTP [server](server/main.go) which glues everthhing together. It serves all the HTML/JS for the web_app/spine_runtime, as well as provides the HTTP endpoints for connecting to the Rooms (i.e `/room`) and the Spine Bridge websocket connections (`/ws`). It holds all the state within the Rooms Manager, and provides the connection to the DB.


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

## Setup Local Development Environment
1. Make the `./secrets` folder. It should contain:
    ```
    dev_config.json
    postgres-password.txt
    web-user-password.txt
    ```
3. `dev_config.json` should be a [misc.BotConfig](server/internal/misc/config.go) json file.
4. `*-password.txt` files should contain the passwords for the `postgres` and `web_user` database users.
   The `postgres` user will have the `postgres-password.txt` set upon first DB creation.
   But `web_user` will need to be updated with the new password manually.
5. Make `db/certs` folder and create a self-signed certificate.
    ```
    # Create CA private key
    openssl genrsa -des3 -out root.key 4096
    openssl rsa -in root.key -out root.key
    openssl  req -new -x509  -days 365  -subj "/CN=jordanyu.com"  -key root.key  -out root.crt
    # Create Server Key and Certificate
    openssl genrsa -des3 -out privkey.pem 4096
    openssl rsa -in privkey.pem -out privkey.pem
    openssl req -new -key privkey.pem -subj "/CN=localhost" -text -out server.csr
    openssl x509 -req -in server.csr -text -days 365 -CA root.crt -CAkey root.key -CAcreateserial -out fullchain.pem
    ```
8. Make `./env` file. The contents should be something below. Update the 
  `MIGRATE_DB_PASSWORD` the same as what is in `postgres-password.txt`
    ```
    MIGRATE_DB_USER=postgres
    MIGRATE_DB_PASSWORD=password
    ``` 
1. Download the chibi assets files from [here](https://f002.backblazeb2.com/file/ak-gamedata/assets.zip) and extract into the `static/assets` folder.
5. Bring up all the containers. `docker compose -f compose.yaml up --build`
6. This should bring up the postgres DB, run any schema migrations and then spin
   up the `bot` AKChibiBot server.
7. Login to your Database and run `ALTER USER web_user WITH ENCRYPTED PASSWORD 'REPLACE_ME'` to set your `web_users` password. It should match the contents of `web-user-password.txt`
10. Restart the containers so that the `bot` service can pick up the new DB password.

### Schema Updates
1. DB schema migrations are under `db/migrations`. We use [golang-migrate](https://github.com/golang-migrate/migrate) to
manage all our DB schema migrations/updates.
2. Bringing up the services using the `compose.yaml` file will automatically 
 bring up a service to pick up any new migrations.
3. To add new migrations/schemas. Run the `migrate create` command.
    ```
    cd db
    docker run -it -v .\\migrations:/migrations migrate/migrate create -ext sql -dir /migrations schema_name
    ```
4. To downgrade the migrations. You can explicitly run the `migrate down` command.
    ```
    docker run -it -v .\\migrations:/migrations --network ak-services_default migrate/migrate \
    -database postgres://postgres:<REPLACE_ME_WITH_DB_PASSWORD>@db:5432/akdb  -path ./migrations down
    ```

### Hot Reloading

#### Reload the HTTP Server
To hot-reload the `ak-chibi-bot` server:
Running the `compose.DEV.yaml` services will bring up the `ak-chibi-bot` server 
to be hot-reloaded using [air-verse](https://github.com/air-verse/air). 
It will pick up any `golang` changes automatically and reload the service.

#### Reload the Spine Runtime Typescript Library
To hot-reload the spine-runtime typescript library:
```
$ docker exec -it ak-services-bot-1 bash
/app
cd /app/static/spine
tsc -w -p tsconfig.json
```

#### Reloading the Web App (React)
Several pages in the React App required OAUTH to access the pages and view data
so in order to correctly edit those pages we need to serve the HTML/JS directly
from the server. This means we need to run a full build for each edit to the files.
```
cd static/web_app
npm run watch
```

If you are _only_ editting the HTML/JS you might be able to get away with
a local development server and use React's normal dev tools
```
cd static/web_app
npm run start
```