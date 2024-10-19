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
import { ManagedWebGLRenderingContext } from "./WebGL";
import { Restorable, Disposable } from "../core/Utils";


type AssetManagerContext = ManagedWebGLRenderingContext | WebGLRenderingContext;
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

	constructor(context: AssetManagerContextList | AssetManagerContext, pathPrefix: string = "") {
		super((image: HTMLImageElement) => {

			if (context instanceof ManagedWebGLRenderingContext || context instanceof WebGLRenderingContext) {
				return new GLTexture(context as AssetManagerContext, image);
			} else {
				let ctx = context as AssetManagerContextList;
				for (let i = 1; i < ctx.length; i++) {
					this.altTextures.push(new GLTexture(ctx[i], image));
				}
				return new GLTexture(ctx[0], image);
			}

		}, pathPrefix);
	}

	dispose() {
		super.dispose();
		for (let i = 0; i < this.altTextures.length; i++) {
			this.altTextures[i].dispose();
		}
		this.altTextures = [];
	}
}

