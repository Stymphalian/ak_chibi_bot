import { Color, Vector2 } from "../core/Utils";
import { PerspectiveCamera, Camera } from "../webgl/Camera";
import { M00 } from "../webgl/Matrix4";
import { SceneRenderer, ResizeMode } from "../webgl/SceneRenderer";
import { Vector3 } from "../webgl/Vector3";
import { ManagedWebGLRenderingContext } from "../webgl/WebGL";
import { Actor } from "./Actor";
import { BoundingBox } from "./Player";

export class OffscreenRender {
    public defaultOffscreenCanvasWidth = 800;
    public defaultOffscreenCanvasHeight = 800;
    public offscreenCanvas: OffscreenCanvas|HTMLCanvasElement;
    public offscreenContext: ManagedWebGLRenderingContext;
    public offscreenSceneRenderer: SceneRenderer;

    public constructor(canvas: HTMLCanvasElement|null = null) {
        try {
            var webglConfig = { alpha: true};
            
            if (canvas !== null) {
                this.offscreenCanvas = canvas;
            } else {
                this.offscreenCanvas = new OffscreenCanvas(
                    this.defaultOffscreenCanvasWidth, 
                    this.defaultOffscreenCanvasHeight
                );
            }
            this.offscreenContext = new ManagedWebGLRenderingContext(this.offscreenCanvas, webglConfig);
            this.offscreenSceneRenderer = new SceneRenderer(this.offscreenCanvas, this.offscreenContext, true);
        } catch (e) {
            // this.showError("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
            console.log("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
        }
    }

    public getBoundingBox(actor: Actor, defaultBB: BoundingBox): BoundingBox {
        let ctx = this.offscreenContext;
        let gl = ctx.gl;

        // Clear the viewport
        let bg = new Color().setFromString("#00000000");
        gl.clearColor(bg.r, bg.g, bg.b, bg.a);
        gl.clear(gl.COLOR_BUFFER_BIT);

        // Resize the canvas. Offcanvas should be square which is twice the size
        // of the maximum sized sprite. This way we can guarantee that the full
        // sprite will be rendered in the offcanvas for us to collect the 
        // width/height information
        this.offscreenCanvas.width = defaultBB.width + (2 * Math.abs(defaultBB.x));
        this.offscreenCanvas.height = defaultBB.height + defaultBB.height + (2*Math.abs(defaultBB.y));
        this.offscreenSceneRenderer.resize(ResizeMode.Expand);

        let viewport = {
            x: 0,
            y: 0,
            width: this.offscreenCanvas.width,
            height: this.offscreenCanvas.height,
        };

        // Update the camera
        updateCameraSettings(this.offscreenSceneRenderer.camera, actor, viewport);
        // Center the camera so that is the chibi is rendered below the ground
        // we can still detect that in the bounding box.
        this.offscreenSceneRenderer.camera.position.y = 0;        

        // Draw skeleton 
        this.offscreenSceneRenderer.begin();
        this.offscreenSceneRenderer.drawSkeleton(actor.skeleton, actor.config.premultipliedAlpha);
        this.offscreenSceneRenderer.end();

        // // Dump to file
        // if (this.offscreenCanvas instanceof OffscreenCanvas) {
        //     this.offscreenCanvas.convertToBlob({ type: "image/png" })
        //     .then(function (blob) {
        //         var url = window.URL.createObjectURL(blob);
        //         var a = document.createElement('a');
        //         a.href = url;
        //         a.download = "screenshot.png";
        //         a.click();
        //         window.URL.revokeObjectURL(url);
        //     })
        // }

        let w = this.offscreenCanvas.width;
        let h = this.offscreenCanvas.height;
        let pixels = new Uint8Array(w * h * 4);
        gl.readPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, pixels);

        let minX = Number.POSITIVE_INFINITY;
        let minY = Number.POSITIVE_INFINITY;
        let maxX = Number.NEGATIVE_INFINITY;
        let maxY = Number.NEGATIVE_INFINITY;
        for (let col = 0; col < w; col++) {
            for (let row = 0; row < h; row++) {
                let i = (row * w + col) * 4;
                if (pixels[i+3] == 255) {
                    // console.log(row, col, pixels[i], pixels[i + 1], pixels[i + 2], pixels[i + 3]);
                    minX = Math.min(minX, col);
                    minY = Math.min(minY, row);
                    maxX = Math.max(maxX, col);
                    maxY = Math.max(maxY, row);
                }
            }
        }

        let offset = new Vector2();
        let size = new Vector2();
        offset.x = -(w/2  - minX);
        offset.y = minY - h/2;
        size.x = maxX - minX;
        size.y = maxY - minY;

        actor.canvasBBCalculated += 1;
        // console.log("actor canvasBBCalculated", actor.config.userDisplayName, actor.canvasBBCalculated);
        // console.log(offset, size, minX, maxX, minY, maxY);
        return {
            x: offset.x,
            y: offset.y,
            width: size.x,
            height: size.y
        }
    }
}

export function getPerspectiveCameraZOffset(viewport: BoundingBox, near: number, far: number, fovY:number): number {
    let width = viewport.width;
    let height = viewport.height;
    let cam = new PerspectiveCamera(width, height);
    cam.near = near;
    cam.far = far;
    cam.fov = fovY;
    cam.zoom = 1;
    cam.position.x = 0.0;
    cam.position.y = 0.0;
    cam.position.z = 0.0;
    cam.update();

    let a = cam.projectionView.values[M00];
    let w = -a * (width/2);
    return w;
}

export function configurePerspectiveCamera(cam: Camera, near: number, far: number, viewport: BoundingBox) {
    cam.near = near;
    cam.far = far;
    cam.zoom = 1;
    cam.position.x = 0;
    cam.position.y = viewport.height/2;
    // TODO: Negative so that the view is not flipped?
    cam.position.z = -getPerspectiveCameraZOffset(
        viewport, cam.near, cam.far, cam.fov
    );
    cam.direction = new Vector3(0, 0, -1);
    cam.update();
}

export function updateCameraSettings(cam :Camera, actor: Actor, viewport: BoundingBox) {
    // let cam = this.sceneRenderer.camera;
    cam.position.x = 0;
    cam.position.y = viewport.height/2;
    // TODO: Negative so that the view is not flipped?
    cam.position.z = -getPerspectiveCameraZOffset(viewport, cam.near, cam.far, cam.fov);
    cam.position.z += actor.getPositionZ();
    cam.direction = new Vector3(0, 0, -1);
    // let origin = new Vector3(0,0,0);
    // let pos = new Vector3(0,0,cam.position.z)
    // let dir = origin.sub(pos).normalize();
    // cam.direction = dir;
    cam.update();
}

export function isContained(dom: HTMLElement, needle: HTMLElement): boolean {
    if (dom === needle) return true;
    let findRecursive = (dom: HTMLElement, needle: HTMLElement) => {
        for(var i = 0; i < dom.children.length; i++) {
            let child = dom.children[i] as HTMLElement;
            if (child === needle) return true;
            if (findRecursive(child, needle)) return true;
        }
        return false;
    };
    return findRecursive(dom, needle);
}

export function findWithId(dom: HTMLElement, id: string): HTMLElement[] {
    let found = new Array<HTMLElement>()
    let findRecursive = (dom: HTMLElement, id: string, found: HTMLElement[]) => {
        for(var i = 0; i < dom.children.length; i++) {
            let child = dom.children[i] as HTMLElement;
            if (child.id === id) found.push(child);
            findRecursive(child, id, found);
        }
    };
    findRecursive(dom, id, found);
    return found;
}

export function findWithClass(dom: HTMLElement, className: string): HTMLElement[] {
    let found = new Array<HTMLElement>()
    let findRecursive = (dom: HTMLElement, className: string, found: HTMLElement[]) => {
        for(var i = 0; i < dom.children.length; i++) {
            let child = dom.children[i] as HTMLElement;
            if (child.classList.contains(className)) found.push(child);
            findRecursive(child, className, found);
        }
    };
    findRecursive(dom, className, found);
    return found;
}

export function createElement(html: string): HTMLElement {
    let dom = document.createElement("div");
    dom.innerHTML = html;
    return dom.children[0] as HTMLElement;
}

export function removeClass(elements: HTMLCollection, clazz: string) {
    for (var i = 0; i < elements.length; i++) {
        elements[i].classList.remove(clazz);
    }
}

export function escapeHtml(str: string) {
    if (!str) return "";
    return str
         .replace(/&/g, "&amp;")
         .replace(/</g, "&lt;")
         .replace(/>/g, "&gt;")
         .replace(/"/g, "&#34;")
         .replace(/'/g, "&#39;");
 }

export function isAlphanumeric(str: string): boolean {
    return /^[a-zA-Z0-9_-]{1,100}$/.test(str)
}