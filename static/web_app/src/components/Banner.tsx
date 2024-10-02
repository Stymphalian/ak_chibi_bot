import { Code } from "./Code"
export function Banner() {
    return (
        <div className="container-fluid pt-4 pb-2 my-2 bg-light border rounded-3">
            <h1 className="display-5 fw-bold">Arknights Chibi Bot</h1>
            <p className="lead">
                A twitch bot and browser source overlay to show Arknight chibis walking on your stream. <br />
                Viewers can issue <Code>!chibi</Code> chat commands to 
                choose their own operator, change skins and play different animations.
            </p>
        </div>
    )

}