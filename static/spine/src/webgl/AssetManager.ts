/******************************************************************************
 * Spine Runtimes License Agreement
 * Last updated January 1, 2020. Replaces all prior versions.
 *
 * Copyright (c) 2013-2020, Esoteric Software LLC
 *
 * Integration of the Spine Runtimes into software or otherwise creating
 * derivative works of the Spine Runtimes is permitted under the terms and
 * conditions of Section 2 of the Spine Editor License Agreement:
 * http://esotericsoftware.com/spine-editor-license
 *
 * Otherwise, it is permitted to integrate the Spine Runtimes into software
 * or otherwise create derivative works of the Spine Runtimes (collectively,
 * "Products"), provided that each user of the Products must obtain their own
 * Spine Editor license and redistribution of the Products in any form must
 * include this license and copyright notice.
 *
 * THE SPINE RUNTIMES ARE PROVIDED BY ESOTERIC SOFTWARE LLC "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL ESOTERIC SOFTWARE LLC BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES,
 * BUSINESS INTERRUPTION, OR LOSS OF USE, DATA, OR PROFITS) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 * THE SPINE RUNTIMES, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *****************************************************************************/

import { AssetManager as spineAssetManager } from "../core/AssetManager";
import { GLTexture } from "./GLTexture";
import { GLCompressedTexture } from "./GLCompressedTexture";
import { ManagedWebGLRenderingContext } from "./WebGL";
import { loadDDS, DDSInfo } from "./DDSLoader";
import { CompressionCapabilities } from "./CompressionCapabilities";


type AssetManagerContext = ManagedWebGLRenderingContext | WebGL2RenderingContext;
type AssetManagerContextList = AssetManagerContext[];

export class AssetManager extends spineAssetManager {

	// HACK!!!!
	// This is a hack to allow loading textures for multiple webGl Context
	// without needing to keep multiple assetManagers around.
	// We load the GLTextures for all the contexts but only the first
	// context is treated as 'primary' and returned from the textureLoader
	// callback. This allows the existing code for TextureAtlas to work
	// normally.
	private altTextures: GLTexture[] = [];
	// TODO: This saved seenImages should happen within the spineAssetManager 
	// instead of here.
	private seenImages = new Map<string, GLTexture>();
	private compressionCapabilities: CompressionCapabilities | null = null;
	private useCompressedTextures: boolean = false;
	private context: AssetManagerContextList | AssetManagerContext;

	constructor(context: AssetManagerContextList | AssetManagerContext, pathPrefix: string = "", compressionCapabilities: CompressionCapabilities | null = null) {
		super((image: HTMLImageElement) => {
			if (this.seenImages.has(image.src)) {
				return this.seenImages.get(image.src);
			}

			if (context instanceof ManagedWebGLRenderingContext || context instanceof WebGL2RenderingContext) {
				this.seenImages.set(image.src, new GLTexture(context, image));
				return this.seenImages.get(image.src);
				// return new GLTexture(context as AssetManagerContext, image);
			} else {
				let ctx = context as AssetManagerContextList;
				for (let i = 1; i < ctx.length; i++) {
					this.altTextures.push(new GLTexture(ctx[i], image));
				}
				return new GLTexture(ctx[0], image);
			}

		}, pathPrefix);
		
		// Store context for later use
		this.context = context;
		
		// Store compression capabilities and check if we can use compressed textures
		this.compressionCapabilities = compressionCapabilities;
		this.useCompressedTextures = compressionCapabilities !== null && compressionCapabilities.formats.s3tc === true;
		
		if (this.useCompressedTextures) {
			console.log('🎨 Compressed textures enabled (DXT/S3TC)');
		} else {
			console.log('📦 Using uncompressed textures');
		}
	}

	/**
	 * Try to load a compressed version of a texture path.
	 * Replaces the extension with .dds and tries to load it.
	 * 
	 * @param path - Original texture path (e.g., "/assets/characters/char_001.png")
	 * @returns Promise that resolves with DDSInfo or null if not available
	 */
	private async tryLoadCompressedTexture(path: string): Promise<DDSInfo | null> {
		if (!this.useCompressedTextures) {
			return null;
		}
		
		try {
			// Convert path: /assets/characters/char_001.png -> /assets/characters/char_001.dds
			const compressedPath = path.replace(/\.(png|jpg|jpeg)$/i, '.dds');
			
			const ddsInfo = await loadDDS(compressedPath);
			return ddsInfo;
		} catch (error) {
			// Compressed texture not available, will fall back to uncompressed
			// Only log actual errors (not 404s which are expected)
			if (error.message) {
				console.warn(`Error loading compressed texture for ${path}:`, error.message);
			}
			return null;
		}
	}

	/**
	 * Get the WebGL contexts (helper for texture creation)
	 */
	private getContexts(): AssetManagerContext[] {
		// This is a bit of a hack to access the context from parent
		// We stored it during construction
		if (this.context) {
			if (Array.isArray(this.context)) {
				return this.context;
			}
			return [this.context];
		}
		return [];
	}
	
	/**
	 * Helper to set asset in parent class asset map
	 */
	private setAsset(path: string, asset: any) {
		// Access parent's assets map (it's protected, we need to cast)
		(this as any).assets[path] = asset;
	}
	
	/**
	 * Override loadTexture to support compressed textures.
	 * Tries to load compressed DDS first, falls back to uncompressed.
	 */
	loadTexture(
		path: string,
		success: (path: string, image: HTMLImageElement) => void = null,
		error: (path: string, error: string) => void = null
	) {
		// If compressed textures are not supported, use parent implementation
		if (!this.useCompressedTextures) {
			super.loadTexture(path, success, error);
			return;
		}
		
		const originalPath = path;
		const fullPath = (this as any).pathPrefix + path;
		
		// Try loading compressed texture first
		this.tryLoadCompressedTexture(fullPath).then(ddsInfo => {
			if (ddsInfo) {
				// Successfully loaded compressed texture
				// console.log(`✅ Loaded compressed: ${path}`);
				
				// Create compressed texture for each context
				const contexts = this.getContexts();
				
				if (contexts.length > 1) {
					for (let i = 1; i < contexts.length; i++) {
						this.altTextures.push(new GLCompressedTexture(contexts[i], ddsInfo));
					}
				}
				
				const primaryTexture = new GLCompressedTexture(contexts[0], ddsInfo);
				
				// Store under original path so atlas lookups work
				this.seenImages.set(fullPath, primaryTexture);
				this.setAsset(fullPath, primaryTexture);
				
				// Call success with null image (we don't have an HTMLImageElement)
				if (success) success(fullPath, null);
			} else {
				// Fall back to uncompressed texture loading
				super.loadTexture(path, (loadedPath, image) => {
					console.log(`📦 Loaded uncompressed: ${path}`);
					if (success) success(loadedPath, image);
				}, error);
			}
		}).catch(err => {
			// Error loading compressed, fall back to uncompressed
			console.log(`⚠️  Compressed load failed for ${path}, using uncompressed with error:`, err);
			super.loadTexture(path, success, error);
		});
	}
	
	/**
	 * Calculate estimated VRAM usage from loaded textures
	 * @returns VRAM usage in megabytes
	 */
	getEstimatedVRAMUsage(): number {
		let totalBytes = 0;
		
		// Calculate bytes for all loaded textures
		this.seenImages.forEach((texture) => {
			const img = texture.getImage();
			if (img) {
				const width = img.width;
				const height = img.height;
				
				// Check if this is a compressed texture
				if (texture instanceof GLCompressedTexture) {
					// DXT1 = 0.5 bytes per pixel, DXT3/5 = 1 byte per pixel
					// We'll assume DXT5 (worst case) = 1 byte per pixel
					totalBytes += width * height * 1;
				} else {
					// Uncompressed RGBA = 4 bytes per pixel
					totalBytes += width * height * 4;
				}
			}
		});
		
		// Convert to megabytes
		return totalBytes / (1024 * 1024);
	}

	dispose() {
		super.dispose();
		for (let i = 0; i < this.altTextures.length; i++) {
			this.altTextures[i].dispose();
		}
		this.altTextures = [];
	}
}

