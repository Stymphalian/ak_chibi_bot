
class CharacterSwapper {
    constructor() {
        this.socket = null;
        this.spinePlayer = null;

        this.openWebSocket();
    }

    openWebSocket() {
        console.log("Openning websocket");
        this.socket = new WebSocket("ws://localhost:7001/spine");
        this.socket.addEventListener("open", (event) => {
            console.log("Socket opened");
        });
        this.socket.addEventListener("message", this.messageHandler.bind(this));
        this.socket.addEventListener("close", (event) => {
            console.log("Close received: " + event.data);
        });
        this.socket.addEventListener("error", (event) => {
            console.log("Error received: " + event.data);
        });
    }

    messageHandler(event) {
        console.log("Message received: " + event.data);
        let requestData = JSON.parse(event.data)
        console.log(requestData);

        if (requestData["type_name"] == "SET_OPERATOR") {
            this.swapCharacter(requestData)
        } else if (requestData["type_name"] == "REMOVE_OPERATOR") {
            this.removeCharacter(requestData);
        } else if (requestData["type_name"] == "UPDATE_OPERATOR") {
            this.updateCharacter(requestData);
        }
    }
    
    swapCharacter(requestData) {
        // let map = new Map();
        // map[requestData["skel_file"]] = requestData["skel_file_base64"];
        // map[requestData["atlas_file"]] = requestData["atlas_file_base64"];
        // map[requestData["png_file"]] = requestData["png_file_base64"];
        let username = requestData["user_name"];
        this.spinePlayerConfig = {
            backgroundColor: "#00000000",
            alpha: true,
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
            }
        };
        this.actorConfig = {
            userDisplayName: requestData['user_name_display'],
            skelUrl: requestData["skel_file"],
            atlasUrl: requestData["atlas_file"],
            animation: requestData["animation"],
            scaleX: 0.45,
            scaleY: 0.45,
            maxSizePx: 350,
            // rawDataURIs: map,
            premultipliedAlpha: true,
            backgroundColor: "#00000000",
            success: (widget) => {
                console.log("Successfully loaded actor");
            },
            error: (widget, error) => {
                console.log(error);
            }
        };

        if (this.spinePlayer == null) {
            console.log("Creating a new spine player");
            this.spinePlayer = new spine.SpinePlayer("container", this.spinePlayerConfig);   
        }
        this.spinePlayer.changeOrAddActor(username, this.actorConfig);
    }

    removeCharacter(requestData) {
        if (this.spinePlayer) {
            this.spinePlayer.removeActor(requestData["user_name"]);
        }
    }

    // updateOperator(requestData) {
    //     if (!this.spinePlayer) {
    //         return;
    //     }
    //     this.spinePlayer.updateActor(
    //         requestData["user_name"],
    //         requestData['config']
    //     );
    // }
}

window.addEventListener("load", () => {
    window.CharacterSwapper = new CharacterSwapper();
});