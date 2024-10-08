import "./main.css";
import { Runtime } from "./stym/runtime";
import { getRuntimeConfigFromQueryParams } from "./stym/utils";

export {}
declare global {
    interface Window {
        SpineRuntime: Runtime;
        // controlCam: ControlCam
    }
}


window.addEventListener("load", () => {
    const queryString = window.location.search;
    const searchParams = new URLSearchParams(queryString);
    const channelName = searchParams.get('channelName');
    const config = getRuntimeConfigFromQueryParams(searchParams);
    window.SpineRuntime = new Runtime(channelName, config);
    // window.ControlCam = new ControlCam(
    // 	"custom_container", 
    // 	config.width, 
    // 	config.height,
    // 	window.SpineRuntime
    // );
});