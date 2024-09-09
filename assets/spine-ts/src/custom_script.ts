
module stym {
    export class Runtime {

        public socket: WebSocket;
        public spinePlayer: spine.SpinePlayer;
        public spinePlayerConfig: spine.SpinePlayerConfig;
        public actorConfig: spine.SpineActorConfig;
        public backoffTimeMsec: number;
        public backOffMaxtimeMsec: number;
        public channelName: string;
    
        constructor(channelName: string) {
            let font = new FontFace("lato", "url(static/fonts/Lato/Lato-Black.ttf)");
            font.load().then(() => {document.fonts.add(font);})
    
            this.channelName = channelName;
            this.socket = null;
            this.spinePlayer = null;
            this.backoffTimeMsec = 1000; // 2.5 seconds
            this.backOffMaxtimeMsec = 1 * 60 * 1000; // 5 minutes
            this.openWebSocket(this.channelName);
    
            if (this.spinePlayer == null) {
                this.spinePlayerConfig = {
                    backgroundColor: "#00000000",
                    alpha: true,
                    showControls: true,
                    viewport: {
                        x: 0,
                        y: 0,
                        width: 1920,
                        height: 1080,
                        padLeft: 0,
                        padRight: 0,
                        padTop: 0,
                        padBottom: 0,
                        debugRender: false,
                    },
                    fullScreenBackgroundColor: null,
                    backgroundImage: null,
                    textSize: 14,
                    textFont: "lato",
                };
                
                console.log("Creating a new spine player");
                this.spinePlayer = new spine.SpinePlayer("container", this.spinePlayerConfig);   
            }
        }
    
        openWebSocket(channelName: string) {
            const protocolPrefix = (window.location.protocol === 'https:') ? 'wss:' : 'ws:';
            const websocketPath =  protocolPrefix + '//' + location.host + `/ws/?channelName=${channelName}`;

            console.log("Openning websocket");
            this.socket = new WebSocket(websocketPath);
            this.socket.addEventListener("open", (event) => {
                console.log("Socket opened");
                this.backoffTimeMsec = 1000;
            });
            this.socket.addEventListener("message", 
                this.messageHandler.bind(this)
            );
            this.socket.addEventListener("close", (event) => {
                console.log("Close received: ",event);

                if (event.code >= 1000 && event.code <= 1002) {
                    return
                }
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
            console.log("Message received: ", requestData);
    
            if (requestData["type_name"] == "SET_OPERATOR") {
                this.swapCharacter(requestData)
            } else if (requestData["type_name"] == "REMOVE_OPERATOR") {
                this.removeCharacter(requestData);
            }
        }
        
        swapCharacter(requestData: any) {
            let username = requestData["user_name"];
            let action = requestData["action"]

            let startPosX = null;
            let startPosY = null;
            if (requestData["start_pos"] != null) {
                startPosX = requestData["start_pos"]["x"];
                startPosY = requestData["start_pos"]["y"];
            }

            let targetPosX = null;
            let targetPosY = null;

            let wandering = false;
            let animations = [];
            switch (action) {
                case "PLAY_ANIMATION":
                    animations = requestData["action_data"]["animations"];
                    break;
                case "WANDER":
                    animations = [requestData["action_data"]["wander_animation"]];
                    wandering = true;
                    break;
                case "WALK_TO":
                    animations = [requestData["action_data"]["walk_to_animation"]];
                    targetPosX = requestData["action_data"]["target_pos"]["x"];
                    targetPosY = requestData["action_data"]["target_pos"]["y"];
                    break;
                default:
                    return
            }
    
            this.actorConfig = {
                chibiId: requestData["operator_id"],
                userDisplayName: requestData['user_name_display'],
                skelUrl: requestData["skel_file"],
                atlasUrl: requestData["atlas_file"],

                animations: animations,

                startPosX: startPosX,
                startPosY: startPosY,
                scaleX: 0.45,
                scaleY: 0.45,
                // scaleX: 0.15,
                // scaleY: 0.15,
                maxSizePx: 350,
                premultipliedAlpha: true,
                animationPlaySpeed: requestData["animation_speed"] ?  requestData["animation_speed"] : 1.0,
    
                extraOffsetX: 0,
                extraOffsetY: 0,
    
                desiredPositionX: targetPosX,
                desiredPositionY: targetPosY,
                wandering: wandering,
    
                success: (widget) => {
                    // console.log("Successfully loaded actor");
                },
                error: (widget, error) => {
                    console.log(error);
                }
            };
    
            this.spinePlayer.changeOrAddActor(username, this.actorConfig);
        }
    
        removeCharacter(requestData: any) {
            if (this.spinePlayer) {
                this.spinePlayer.removeActor(requestData["user_name"]);
            }
        }
    }

}



