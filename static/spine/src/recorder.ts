import "./main.css";
import { Runtime } from "./stym/runtime";
import { CanvasRecorder } from "./stym/canvas_recorder";
import { getRuntimeConfigFromQueryParams } from "./stym/utils";

export {}
declare global {
    interface Window {
        SpineRuntime: Runtime;
        CanvasRecorder: CanvasRecorder
    }
}

window.addEventListener("load", () => {
    const queryString = window.location.search;
    const searchParams = new URLSearchParams(queryString);
    const channelName = searchParams.get('channelName');
    const config = getRuntimeConfigFromQueryParams(searchParams);
    window.SpineRuntime = new Runtime(channelName, config);
    window.CanvasRecorder = new CanvasRecorder("custom_container", window.SpineRuntime);
});