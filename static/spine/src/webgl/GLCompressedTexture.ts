/**
 * GLCompressedTexture.ts
 * 
 * WebGL texture that supports compressed formats (DXT/S3TC).
 * Extends GLTexture to handle compressed texture data.
 */

import { GLTexture } from "./GLTexture";
import { ManagedWebGLRenderingContext } from "./WebGL";
import { DDSInfo } from "./DDSLoader";

export class GLCompressedTexture extends GLTexture {
    private ddsInfo: DDSInfo;
    private s3tcExtension: any = null;

    constructor(
        context: ManagedWebGLRenderingContext | WebGL2RenderingContext,
        ddsInfo: DDSInfo
    ) {
        // Call parent constructor with null image
        // Note: parent calls restore() which calls update(), but we guard update() until ddsInfo is set
        super(context, null, false, ddsInfo.width, ddsInfo.height);
        this.ddsInfo = ddsInfo;
        
        // Get S3TC extension (parent already stored context)
        const gl = (this as any).context.gl;
        this.s3tcExtension = gl.getExtension('WEBGL_compressed_texture_s3tc') ||
                            gl.getExtension('WEBKIT_WEBGL_compressed_texture_s3tc') ||
                            gl.getExtension('MOZ_WEBGL_compressed_texture_s3tc');
        
        if (!this.s3tcExtension) {
            console.error('S3TC extension not available');
        }
        
        // Now that ddsInfo is set, call update to upload the texture
        this.update(false);
    }

    update(useMipMaps: boolean) {
        // Guard against being called before ddsInfo is set (happens in parent constructor)
        if (!this.ddsInfo) {
            return;
        }
        
        const gl = (this as any).context.gl;
        
        if (!this.texture) {
            this.texture = gl.createTexture();
        }
        
        this.bind();
        
        // Upload each mipmap level
        for (let i = 0; i < this.ddsInfo.mipmaps.length; i++) {
            gl.compressedTexImage2D(
                gl.TEXTURE_2D,
                i,                              // Mipmap level
                this.ddsInfo.format,            // Internal format (DXT1/3/5)
                this.ddsInfo.width >> i,        // Width at this level
                this.ddsInfo.height >> i,       // Height at this level
                0,                              // Border (must be 0)
                this.ddsInfo.mipmaps[i]         // Compressed data
            );
        }
        
        // Set texture parameters
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
        
        // Use mipmaps if available in DDS file
        if (this.ddsInfo.mipmapCount > 1) {
            gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
        } else {
            gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        }
        
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    }

    restore() {
        this.texture = null;
        this.update(false);
    }
}
