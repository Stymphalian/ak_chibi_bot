
import { Actor, SpineActorConfig } from "../player/Actor";
import { SpinePlayer, SpinePlayerConfig } from "../player/Player";

export interface RuntimeConfig {
    width: number,
    height: number
    debugMode: boolean
    chibiScale: number
    accurateBoundingBoxFlag: boolean
    showChatMessagesFlag: boolean
    usernameBlacklist: string[]
}

export class Runtime {
    public socket: WebSocket;
    public spinePlayer: SpinePlayer;
    public spinePlayerConfig: SpinePlayerConfig;
    public actorConfig: SpineActorConfig;
    public defaultBackoffTimeMsec: number;
    public backoffTimeMsec: number;
    public backOffMaxtimeMsec: number;
    public channelName: string;
    public runtimeConfig: RuntimeConfig;

    constructor(channelName: string, config: RuntimeConfig) {
        this.runtimeConfig = config;
        const { width, height, debugMode, chibiScale } = config;

        // TODO: Make the fonts configurable
        let font = new FontFace("lato", "url(/public/fonts/Lato/Lato-Black.ttf)");
        font.load().then(() => { document.fonts.add(font); })

        this.channelName = channelName;
        this.socket = null;
        this.spinePlayer = null;
        this.defaultBackoffTimeMsec = 15 * 1000; // 15 seconds
        this.backoffTimeMsec = 15 * 1000; // 15 seconds
        this.backOffMaxtimeMsec = 5 * 60 * 1000; // 5 minutes
        this.openWebSocket(this.channelName);

        if (this.spinePlayer == null) {
            this.spinePlayerConfig = {
                backgroundColor: "#00000000",
                alpha: true,
                showControls: true,
                viewport: {
                    x: 0,
                    y: 0,
                    width: width,
                    height: height,
                    padLeft: 0,
                    padRight: 0,
                    padTop: 0,
                    padBottom: 0,
                    debugRender: debugMode,
                },
                fullScreenBackgroundColor: null,
                backgroundImage: null,
                textSize: 14,
                textFont: "lato",
                runtimeDebugInfoDumpIntervalSec: 60,
                chibiScale: this.runtimeConfig.chibiScale,
                cameraPerspectiveNear: 1,
                cameraPerspectiveFar: 2000,
                useAccurateBoundingBox: this.runtimeConfig.accurateBoundingBoxFlag,
                showChatMessages: this.runtimeConfig.showChatMessagesFlag,
            };

            console.log("Creating a new spine player");
            this.spinePlayer = new SpinePlayer("container", this.spinePlayerConfig);
        }
    }

    openWebSocket(channelName: string) {
        const protocolPrefix = (window.location.protocol === 'https:') ? 'wss:' : 'ws:';
        const websocketPath = protocolPrefix + '//' + location.host + `/ws/?channelName=${channelName}`;

        console.log("Openning websocket");
        this.socket = new WebSocket(websocketPath);
        this.socket.addEventListener("open", (event) => {
            console.log("Socket opened");
            this.backoffTimeMsec = this.defaultBackoffTimeMsec;
            this.spinePlayer.setWebsocket(this.socket);
        });
        this.socket.addEventListener("message",
            this.messageHandler.bind(this)
        );
        this.socket.addEventListener("close", (event) => {
            console.log("Close received: ", event);
            this.spinePlayer.setWebsocket(null);

            // if (event.code >= 1000 && event.code <= 1002) {
            //     return
            // }
            this.backoffTimeMsec *= 2;
            if (this.backoffTimeMsec < this.backOffMaxtimeMsec) {
                console.log("Retrying in " + this.backoffTimeMsec + "ms");
                setTimeout(() => this.openWebSocket(this.channelName), this.backoffTimeMsec);
            }
        });
        this.socket.addEventListener("error", (event) => {
            console.log("Error received: ", event);
        });
    }

    messageHandler(event: MessageEvent) {
        let requestData = JSON.parse(event.data)
        if (requestData["type_name"] == "SET_OPERATOR") {
            console.log("Message received: ", requestData);
            this.swapCharacter(requestData)
        } else if (requestData["type_name"] == "REMOVE_OPERATOR") {
            console.log("Message received: ", requestData);
            this.removeCharacter(requestData);
        } else if (requestData["type_name"] == "SHOW_CHAT_MESSAGE")  {
            this.showChatMessage(requestData);
        }
    }

    swapCharacter(requestData: any) {
        let username = requestData["user_name"];
        if (this.runtimeConfig.usernameBlacklist.includes(username)) {
            console.log("User " + username + " is blacklisted");
            return;
        }

        let startPosX = null;
        let startPosY = null;
        if (requestData["start_pos"] != null) {
            startPosX = requestData["start_pos"]["x"];
            startPosY = requestData["start_pos"]["y"];
        }
        let configScaleX = 1;
        let configScaleY = 1;
        if (requestData["sprite_scale"] != null) {
            configScaleX = requestData["sprite_scale"]["x"];
            configScaleY = requestData["sprite_scale"]["y"];
        }
        let configMaxPixelSize = 350;
        if (requestData["max_sprite_pixel_size"] != null) {
            configMaxPixelSize = requestData["max_sprite_pixel_size"];
        }

        let referenceMovementSpeedPx = 80;
        let referenceMovementSpeedPy = 80;
        let referenceMovementSpeedPz = 80;
        if (requestData["movement_speed_px"] != null) {
            referenceMovementSpeedPx = requestData["movement_speed_px"];
        }
        if (requestData["movement_speed_py"] != null) {
            referenceMovementSpeedPx = requestData["movement_speed_py"];
        }
        if (requestData["movement_speed_pz"] != null) {
            referenceMovementSpeedPz = requestData["movement_speed_pz"];
        }
        let movementSpeedPxX = null;
        let movementSpeedPxY = null;
        let movementSpeedPxZ = referenceMovementSpeedPz;
        if (requestData["movement_speed"] != null) {
            movementSpeedPxX = Math.floor(requestData["movement_speed"]["x"] * referenceMovementSpeedPx);
            movementSpeedPxY = Math.floor(requestData["movement_speed"]["y"] * referenceMovementSpeedPy);
            // movementSpeedPxZ = Math.floor(requestData["movement_speed"]["z"] * referenceMovementSpeedPz);
        }
        let defaultMovementSpeedPxX = referenceMovementSpeedPx;
        let defaultMovementSpeedPxY = referenceMovementSpeedPy;
        let defaultMovementSpeedPxZ = referenceMovementSpeedPz;

        this.actorConfig = {
            chibiId: requestData["operator_id"],
            userDisplayName: requestData['user_name_display'],
            skelUrl: requestData["skel_file"],
            atlasUrl: requestData["atlas_file"],

            startPosX: startPosX,
            startPosY: startPosY,
            defaultMovementSpeedPxX: defaultMovementSpeedPxX,
            defaultMovementSpeedPxY: defaultMovementSpeedPxY,
            defaultMovementSpeedPxZ: defaultMovementSpeedPxZ,
            movementSpeedPxX: movementSpeedPxX,
            movementSpeedPxY: movementSpeedPxY,
            movementSpeedPxZ: movementSpeedPxZ,

            configScaleX: configScaleX,
            configScaleY: configScaleY,
            scaleX: 0.45 * this.runtimeConfig.chibiScale * configScaleX,
            scaleY: 0.45 * this.runtimeConfig.chibiScale * configScaleY,
            maxSizePx: configMaxPixelSize,
            premultipliedAlpha: true,
            animationPlaySpeed: requestData["animation_speed"] ? requestData["animation_speed"] : 1.0,
            extraOffsetX: 0,
            extraOffsetY: 0,

            action: requestData["action"],
            action_data: requestData["action_data"],

            success: (widget, actor) => {
                // console.log("Successfully loaded actor");
            },
            error: (widget, actor: Actor, error) => {
                actor.load_attempts += 1;
                if (actor.load_attempts > actor.max_load_attempts) {
                    actor.load_perma_failed = true;
                }
                console.log(this);
                console.log(error);
            }
        };

        this.spinePlayer.addActorToUpdateQueue(username, this.actorConfig);
    }

    removeCharacter(requestData: any) {
        if (this.spinePlayer) {
            this.spinePlayer.removeActor(requestData["user_name"]);
        }
    }

    showChatMessage(requestData: any) {
        if (this.spinePlayer) {
            this.spinePlayer.showChatMessage(
                requestData["user_name"],
                requestData["message"],
            );
        }
    }
}
