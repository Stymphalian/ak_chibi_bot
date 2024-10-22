import { RuntimeConfig } from "./runtime";

export function setContainerSizeFromQuery(searchParams: URLSearchParams) {
    let containerWidth = searchParams.has('width')
        ? Number.parseInt(searchParams.get('width'))
        : 1920;
    let containerHeight = searchParams.has('height')
        ? Number.parseInt(searchParams.get('height'))
        : 1080;
    let container = document.getElementById("container")
    if (!Number.isInteger(containerWidth)) {
        containerWidth = 1920;
    }
    if (!Number.isInteger(containerHeight)) {
        containerHeight = 1080;
    }
    containerWidth = Math.max(containerWidth, 0)
    containerHeight = Math.max(containerHeight, 0)

    // const windowWidth = window.screen.availWidth;
    // const windowHeight = window.screen.availHeight;
    // containerWidth = Math.min(containerWidth, windowWidth);
    // containerHeight = Math.min(containerHeight, windowHeight);

    console.log(containerWidth, containerHeight);
    container.style.width = `${containerWidth}px`
    container.style.height = `${containerHeight}px`
    return [containerWidth, containerHeight];
}

export function getRuntimeConfigFromQueryParams(searchParams: URLSearchParams): RuntimeConfig {
    const debugMode = searchParams.get('debug') === 'true';
    const useAccurateBoundingBox = searchParams.get('accurate_bb') === 'true';
    const showChatMessages = searchParams.get('show_chat') === 'true';
    const scale = Math.max(Math.min((parseFloat(searchParams.get('scale')) || 1), 3), 0.1);
    const [containerWidth, containerHeight] = setContainerSizeFromQuery(searchParams);
    return {
        width: containerWidth,
        height: containerHeight,
        debugMode: debugMode,
        chibiScale: scale,
        accurateBoundingBoxFlag: useAccurateBoundingBox,
        showChatMessagesFlag: showChatMessages,
    }
}