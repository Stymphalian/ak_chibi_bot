export {}
declare global {
    interface Window {
        SpineRuntime: stym.Runtime;
        ControlCam: stym.ControlCam;
    }
}
