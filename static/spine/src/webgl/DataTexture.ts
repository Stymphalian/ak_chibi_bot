import { Restorable, Disposable, ArrayLike } from "../core/Utils";
import { ManagedWebGLRenderingContext } from "./WebGL";

export class DataTexture implements Disposable, Restorable {
    static MAX_WIDTH = 1024;

    private context: ManagedWebGLRenderingContext;
    private texture: WebGLTexture;
    private numComponents: number;
    private maxVertices: number;
    private expandedData: Float32Array = null;
    private pixelBoundarySize: number = 1;
    public widthPx: number; 
    public heightPx: number;
    
    constructor(context: ManagedWebGLRenderingContext, maxVertices: number, numComponents: number) {
        this.context = context;
        this.texture = null;

        this.numComponents = numComponents
        this.pixelBoundarySize = Math.ceil(numComponents / 4);
        this.widthPx = Math.min(DataTexture.MAX_WIDTH, maxVertices);
        const maxNumPixels = (maxVertices * this.pixelBoundarySize);
        this.heightPx = Math.floor(maxNumPixels / DataTexture.MAX_WIDTH) + 1;
        
        this.maxVertices = maxVertices;
        this.expandedData = null;

        this.context.addRestorable(this);
        this.restore();
    }
    
    public fillData(data: ArrayLike<number>, dataLen: number) {
        // expand the data to 4 values per pixel.
        const numElements = dataLen / this.numComponents;
        if (numElements > this.maxVertices) {
            throw new Error("Too many vertices for the texture: " + numElements);
        }
        const widthPx = this.widthPx;
        const heightPx = Math.floor((numElements*this.pixelBoundarySize) / widthPx) + 1;

        if (this.expandedData == null || this.expandedData.length < widthPx*heightPx*4) {
            this.expandedData = new Float32Array(widthPx*heightPx * 4);
        }
        const expandedData = this.expandedData;
        let dstIndex = 0;
        let pixelBoundaryPaddingIncr = (this.numComponents % 4 == 0) 
            ? 0 
            : 4 - this.numComponents % 4;
        for (let i = 0; i < numElements; ++i) {
            const srcOff = i * this.numComponents;
            for (let j = 0; j < this.numComponents; ++j) {
                expandedData[dstIndex++] = data[srcOff + j];
            }

            // pad to ensure we are on 4-byte/pixel boundary
            dstIndex += pixelBoundaryPaddingIncr;
        }

        let gl = this.context.gl;
        gl.bindTexture(gl.TEXTURE_2D, this.texture);
        gl.texSubImage2D(
            gl.TEXTURE_2D,   // target
            0,               // mimap level (0 == no mipmaps)
            0,               // x offset
            0,               // y offset
            widthPx,
            heightPx,
            gl.RGBA,         // format
            gl.FLOAT,        // element type (float, 4bytes)
            expandedData     // data pointer
        );
    }

    boundUnit: number;
    public bind(unit: number = 0) {
        let gl = this.context.gl;
		this.boundUnit = unit;
		gl.activeTexture(gl.TEXTURE0 + unit);
		gl.bindTexture(gl.TEXTURE_2D, this.texture);  
    }

    unbind () {
		let gl = this.context.gl;
		gl.activeTexture(gl.TEXTURE0 + this.boundUnit);
		gl.bindTexture(gl.TEXTURE_2D, null);
	}

    private makeDataTexture(widthPx: number, heightPx: number) : WebGLTexture{
        let gl = this.context.gl;
        if (this.texture != null) {
            throw new Error("DataTexture.makeDataTexture: texture already exists");
        }

        this.texture = gl.createTexture();
        gl.bindTexture(gl.TEXTURE_2D, this.texture);
        gl.texImage2D(
            gl.TEXTURE_2D,
            0,            // mip level
            gl.RGBA32F,   // format
            widthPx,      // width
            heightPx,     // height
            0,            // border
            gl.RGBA,      // format
            gl.FLOAT,     // type
            null,         // no data, empty
        );
        // make it possible to use a non-power-of-2 texture and
        // we don't need any filtering
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
        return this.texture;
    }

    restore(): void { 
        this.texture = null;
        this.makeDataTexture(this.widthPx, this.heightPx);
    }

    dispose(): void { 
        let gl = this.context.gl;
        gl.deleteTexture(this.texture);
        this.texture = null;
    }

}