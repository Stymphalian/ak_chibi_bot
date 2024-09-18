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

## Setup your Development environment
1. Make the `./secrets` folder. It should contain:
    ```
    dev_config.json
    postgres-password.txt
    web-user-password.txt
    ```
3. `dev_config.json` should be a `misc.BotConfig` json file.
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
5. Bring up all the containers. `docker compose -f compose.DEV.yaml up`
6. This should bring up the postgres DB, run any schema migrations and then spin
   up the `web` AKChibiBot server.
7. Login to your Database and run `ALTER USER web_user WITH ENCRYPTED PASSWORD 'REPLACE_ME'` to set your `web_users` password. It should match the contents of `web-user-password.txt`
10. Restart the containers so that the `web` service can pick up the new DB password.

### Schema Updates
1. DB schema migrations are under `db/migrations`. We use [golang-migrate](https://github.com/golang-migrate/migrate) to
manage all our DB schema migrations/updates.
2. Bringing up the services using the `compose.DEV.yaml` file will automatically 
 bring up a service to pick up any new migrations.
3. To add new migrations/schemas. Run the `migrate create` command.
    ```
    cd db
    docker run -it -v .\\migrations:/migrations migrate/migrate create -ext sql -dir /migrations schema_name
    ```
4. To downgrade the migrations. You can explicitly run the `migrate down` command.
    ```
    docker run -it -v .\\migrations:/migrations --network ak-services_default migrate/migrate \
    -database postgres://postgres:db_password@db:5432/akdb  -path ./migrations down
    ```

### Hot Reloading
To hot-reload the `ak-chibi-bot` server:
Running the `compose.DEV.yaml` services will bring up the `ak-chibi-bot` server 
to be hot-reloaded using `air`. It will pick up any `golang` changes automatically
and reload the service.

To hot-reload the spine-runtime typescript library:
```
docker exec -it ak-services-web-1 base
/app
cd /app/static/spine
tsc -w -p tsconfig.json
```

Now open http://localhost:8080/room/?channelName=REPLACE_ME in your web-browser.
Replce the `REPLACE_ME` with your own twitch channel name and you should see 
default operator walking around.