import { Color, Vector2 } from "../core/Utils";
import { GLFrameBuffer } from "../webgl/GLFrameBuffer";
import { SceneRenderer } from "../webgl/SceneRenderer";
import { ManagedWebGLRenderingContext } from "../webgl/WebGL";
import { Actor } from "./Actor";
import { BoundingBox, updateCameraSettings } from "./Utils";

export class OffscreenRender {
    public sceneRenderer: SceneRenderer;
    public context: ManagedWebGLRenderingContext;
    public frameBuffer: GLFrameBuffer = null;

    public constructor(sceneRenderer: SceneRenderer) {
        try {
            this.sceneRenderer = sceneRenderer;
            this.context = sceneRenderer.context;
            this.frameBuffer = new GLFrameBuffer(this.context, {
                x: 0,
                y: 0,
                width: 800,
                height: 800
            });
        } catch (e) {
            // this.showError("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
            console.log("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
        }
    }

    public getBoundingBox(actor: Actor, defaultBB: BoundingBox): BoundingBox {
        let ctx = this.context;
        let gl = ctx.gl;

        let newWidth = defaultBB.width + (2 * Math.abs(defaultBB.x));
        let newHeight = defaultBB.height + defaultBB.height + (2 * Math.abs(defaultBB.y));
        if (newHeight > 3000) {
            newHeight = 3000;
        }
        // console.log("newWidth ", newWidth, " newHeight ", newHeight);
        let oldviewport = gl.getParameter(gl.VIEWPORT);
        this.frameBuffer.resize({ x: 0, y: 0, width: newWidth, height: newHeight });
        this.frameBuffer.bind();

        // Clear the viewport
        gl.viewport(0, 0, this.frameBuffer.textureWidth, this.frameBuffer.textureHeight);
        this.sceneRenderer.camera.setViewport(this.frameBuffer.textureWidth, this.frameBuffer.textureHeight);
        let bg = new Color().setFromString("#00000000");
        gl.clearColor(bg.r, bg.g, bg.b, bg.a);
        gl.clear(gl.COLOR_BUFFER_BIT);

        let viewport = {
            x: 0,
            y: 0,
            width: this.frameBuffer.textureWidth,
            height: this.frameBuffer.textureHeight
        };

        // Update the camera
        updateCameraSettings(this.sceneRenderer.camera, actor, viewport);
        // Center the camera so that if the chibi is rendered below the ground
        // we can still detect that in the bounding box.
        this.sceneRenderer.camera.position.y = 0;

        // Draw skeleton 
        this.sceneRenderer.begin();
        actor.Draw(this.sceneRenderer)
        // this.offscreenSceneRenderer.drawSkeleton(actor.skeleton, actor.config.premultipliedAlpha);
        // TODO: Draw with spritesheet if needed
        // this.offscreenSceneRenderer.circle(true, 0, 0, 1, Color.RED);
        this.sceneRenderer.end();

        let w = this.frameBuffer.textureWidth;
        let h = this.frameBuffer.textureHeight;
        let pixels = new Uint8ClampedArray(w * h * 4);
        gl.readPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, pixels);

        // Dump to file (for debugging)
        // let imageData = new ImageData(pixels, w, h);
        // this.offscreenCanvas.width = w;
        // this.offscreenCanvas.height = h;
        // this.offscreenCanvas.getContext("2d").putImageData(imageData, 0, 0);
        // this.offscreenCanvas.convertToBlob({ type: "image/png" })
        //     .then(
        //         (blob: any) => {
        //             var url = window.URL.createObjectURL(blob);
        //             var a = document.createElement('a');
        //             a.href = url;
        //             a.download = "screenshot.png";
        //             a.click();
        //             window.URL.revokeObjectURL(url);
        //         })

        // Reset the GL state. otherwise we can get some weird rendering 
        // artifacts in the next rendered frame.
        this.frameBuffer.unbind();

        gl.viewport(oldviewport[0], oldviewport[1], oldviewport[2], oldviewport[3]);
        this.sceneRenderer.camera.setViewport(oldviewport[2], oldviewport[3]);

        let minX = Number.POSITIVE_INFINITY;
        let minY = Number.POSITIVE_INFINITY;
        let maxX = Number.NEGATIVE_INFINITY;
        let maxY = Number.NEGATIVE_INFINITY;
        for (let col = 0; col < w; col++) {
            for (let row = 0; row < h; row++) {
                let i = (row * w + col) * 4;
                if (pixels[i + 3] == 255) {
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
        offset.x = -(w / 2 - minX);
        offset.y = minY - h / 2;
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