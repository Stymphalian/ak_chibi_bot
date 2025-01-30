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

import { BlendMode } from "../core/BlendMode";
import { Restorable } from "../core/Utils";

export class ManagedWebGLRenderingContext {
	public canvas: HTMLCanvasElement | OffscreenCanvas;
	public gl: WebGL2RenderingContext;
	private restorables = new Array<Restorable>();

	constructor(canvasOrContext: HTMLCanvasElement | WebGL2RenderingContext | OffscreenCanvas, contextConfig: any = { alpha: "true" }) {
		if (canvasOrContext instanceof HTMLCanvasElement || canvasOrContext instanceof OffscreenCanvas) {
			let canvas = canvasOrContext;
			// this.gl = <WebGLRenderingContext>(canvas.getContext("webgl2", contextConfig) || canvas.getContext("webgl", contextConfig));
			this.gl = <WebGL2RenderingContext>canvas.getContext("webgl2", contextConfig);
			this.canvas = canvas;
			canvas.addEventListener("webglcontextlost", (e: any) => {
				let event = <WebGLContextEvent>e;
				if (e) {
					e.preventDefault();
				}
			});

			canvas.addEventListener("webglcontextrestored", (e: any) => {
				for (let i = 0, n = this.restorables.length; i < n; i++) {
					this.restorables[i].restore();
				}
			});
		} else {
			this.gl = canvasOrContext;
			this.canvas = this.gl.canvas;
		}
	}

	addRestorable(restorable: Restorable) {
		this.restorables.push(restorable);
	}

	removeRestorable(restorable: Restorable) {
		let index = this.restorables.indexOf(restorable);
		if (index > -1) this.restorables.splice(index, 1);
	}

	GetWebGLParameters() {
		let gl = this.gl;
		let parameters: {[k: string]: any} = {
			MAX_TEXTURE_SIZE: gl.getParameter(gl.MAX_TEXTURE_SIZE),
			MAX_CUBE_MAP_TEXTURE_SIZE: gl.getParameter(gl.MAX_CUBE_MAP_TEXTURE_SIZE),
			MAX_RENDERBUFFER_SIZE: gl.getParameter(gl.MAX_RENDERBUFFER_SIZE),
			MAX_VERTEX_ATTRIBS: gl.getParameter(gl.MAX_VERTEX_ATTRIBS),
			MAX_VERTEX_UNIFORM_VECTORS: gl.getParameter(gl.MAX_VERTEX_UNIFORM_VECTORS),
			MAX_VARYING_VECTORS: gl.getParameter(gl.MAX_VARYING_VECTORS),
			MAX_FRAGMENT_UNIFORM_VECTORS: gl.getParameter(gl.MAX_FRAGMENT_UNIFORM_VECTORS),
			MAX_COMBINED_TEXTURE_IMAGE_UNITS: gl.getParameter(gl.MAX_COMBINED_TEXTURE_IMAGE_UNITS),
			MAX_TEXTURE_IMAGE_UNITS: gl.getParameter(gl.MAX_TEXTURE_IMAGE_UNITS),
			MAX_VIEWPORT_DIMS: gl.getParameter(gl.MAX_VIEWPORT_DIMS),
			VENDOR: gl.getParameter(gl.VENDOR),
			RENDERER: gl.getParameter(gl.RENDERER),
			VERSION: gl.getParameter(gl.VERSION),
			SHADING_LANGUAGE_VERSION: gl.getParameter(gl.SHADING_LANGUAGE_VERSION),
		};
	
		if (gl instanceof WebGL2RenderingContext) {
			parameters = {
				...parameters,
				MAX_DRAW_BUFFERS: gl.getParameter(gl.MAX_DRAW_BUFFERS),
				MAX_ELEMENT_INDEX: gl.getParameter(gl.MAX_ELEMENT_INDEX),
				MAX_ELEMENTS_INDICES: gl.getParameter(gl.MAX_ELEMENTS_INDICES),
				MAX_ELEMENTS_VERTICES: gl.getParameter(gl.MAX_ELEMENTS_VERTICES),
				MAX_VARYING_COMPONENTS: gl.getParameter(gl.MAX_VARYING_COMPONENTS),
			}
		}
		console.table(parameters);
		return parameters;
	}
}

export class WebGLBlendModeConverter {
	static ZERO = 0;
	static ONE = 1;
	static SRC_COLOR = 0x0300;
	static ONE_MINUS_SRC_COLOR = 0x0301;
	static SRC_ALPHA = 0x0302;
	static ONE_MINUS_SRC_ALPHA = 0x0303;
	static DST_ALPHA = 0x0304;
	static ONE_MINUS_DST_ALPHA = 0x0305;
	static DST_COLOR = 0x0306

	static getDestGLBlendMode(blendMode: BlendMode) {
		switch (blendMode) {
			case BlendMode.Normal: return WebGLBlendModeConverter.ONE_MINUS_SRC_ALPHA;
			case BlendMode.Additive: return WebGLBlendModeConverter.ONE;
			case BlendMode.Multiply: return WebGLBlendModeConverter.ONE_MINUS_SRC_ALPHA;
			case BlendMode.Screen: return WebGLBlendModeConverter.ONE_MINUS_SRC_ALPHA;
			default: throw new Error("Unknown blend mode: " + blendMode);
		}
	}

	static getSourceGLBlendMode(blendMode: BlendMode, premultipliedAlpha: boolean = false) {
		switch (blendMode) {
			case BlendMode.Normal: return premultipliedAlpha ? WebGLBlendModeConverter.ONE : WebGLBlendModeConverter.SRC_ALPHA;
			case BlendMode.Additive: return premultipliedAlpha ? WebGLBlendModeConverter.ONE : WebGLBlendModeConverter.SRC_ALPHA;
			case BlendMode.Multiply: return WebGLBlendModeConverter.DST_COLOR;
			case BlendMode.Screen: return WebGLBlendModeConverter.ONE;
			default: throw new Error("Unknown blend mode: " + blendMode);
		}
	}
}
