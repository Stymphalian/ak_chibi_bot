import { GLTexture } from "../webgl/GLTexture";
import { ManagedWebGLRenderingContext } from "../webgl/WebGL";
import { Restorable, Disposable } from "../core/Utils";
import { BoundingBox } from "../player/Utils";

export class GLFrameBuffer implements Disposable, Restorable {
    public context: ManagedWebGLRenderingContext;
    public textureWidth: number = -1;
    public textureHeight: number = -1;
    public offscreenTexture: GLTexture = null;
    public frameBuffer: WebGLFramebuffer = null;

    public constructor(context: ManagedWebGLRenderingContext, viewport: BoundingBox) {
        try {
            this.context = context;
            this.textureWidth =  Math.ceil(viewport.width);
            this.textureHeight = Math.ceil(viewport.height);

            this.restore();
            this.context.addRestorable(this);
        } catch (e) {
            // this.showError("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
            console.log("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
        }
    }

    public bind() {
        this.offscreenTexture.bind();
        this.context.gl.bindFramebuffer(this.context.gl.FRAMEBUFFER, this.frameBuffer);
    }

    public unbind() {
        this.offscreenTexture.unbind();
        this.context.gl.bindFramebuffer(this.context.gl.FRAMEBUFFER, null);
    }

    public resize(viewport: BoundingBox) {
        if (this.offscreenTexture) {
            this.offscreenTexture.dispose();
            this.offscreenTexture = null;
        }

        this.textureHeight = Math.ceil(viewport.height);
        this.textureWidth = Math.ceil(viewport.width);
        this.setupFramebuffer();
    }

    private setupFramebuffer() {
        let context = this.context;
        let gl = context.gl;

        if (this.offscreenTexture == null) {
            this.offscreenTexture = new GLTexture(
                context,
                null,
                false,
                this.textureWidth,
                this.textureHeight
            );
        }
        if (this.frameBuffer == null) {
            this.frameBuffer = gl.createFramebuffer();
        }

        this.offscreenTexture.bind();
        gl.bindFramebuffer(gl.FRAMEBUFFER, this.frameBuffer);
        gl.framebufferTexture2D(
            gl.FRAMEBUFFER,
            gl.COLOR_ATTACHMENT0,
            gl.TEXTURE_2D,
            this.offscreenTexture.getTextureId(),
            0
        );

        var status = gl.checkFramebufferStatus(gl.FRAMEBUFFER);
        if (status != gl.FRAMEBUFFER_COMPLETE) {
            console.log("Framebuffer is not complete: ", status);
        }

        this.offscreenTexture.unbind();
        gl.bindFramebuffer(gl.FRAMEBUFFER, null);
    }

    restore() {
        this.offscreenTexture = null;
        this.frameBuffer = null;
        this.setupFramebuffer();
    }

    dispose () {
        this.context.removeRestorable(this);
        let gl = this.context.gl;

        if (this.offscreenTexture) {
            this.offscreenTexture.dispose();
            this.offscreenTexture = null;
        }
        if (this.frameBuffer) {
            gl.deleteFramebuffer(this.frameBuffer);
            this.frameBuffer = null;
        }
    }
}