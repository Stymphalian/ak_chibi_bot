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

import { Restorable, Disposable } from "../core/Utils";
import { ManagedWebGLRenderingContext } from "./WebGL";

export class Shader implements Disposable, Restorable {
	public static MVP_MATRIX = "u_projTrans";
	public static POSITION = "a_position";
	public static COLOR = "a_color";
	public static COLOR2 = "a_color2";
	public static TEXCOORDS = "a_texCoords";
	public static TEXTURE_INDEX_POS_Z = "a_texture_index_pos_z";
	public static TEXTURE_VERTS = "u_data";
	public static SAMPLER = "u_textures";

	private context: ManagedWebGLRenderingContext;
	private vs: WebGLShader = null;
	private vsSource: string;
	private fs: WebGLShader = null;
	private fsSource: string;
	private program: WebGLProgram = null;
	private tmp2x2: Float32Array = new Float32Array(2 * 2);
	private tmp3x3: Float32Array = new Float32Array(3 * 3);
	private tmp4x4: Float32Array = new Float32Array(4 * 4);

	public getProgram() { return this.program; }
	public getVertexShader() { return this.vertexShader; }
	public getFragmentShader() { return this.fragmentShader; }
	public getVertexShaderSource() { return this.vsSource; }
	public getFragmentSource() { return this.fsSource; }

	constructor(
		context: ManagedWebGLRenderingContext | WebGL2RenderingContext,
		private vertexShader: string,
		private fragmentShader: string) {
		this.vsSource = vertexShader;
		this.fsSource = fragmentShader;
		this.context = context instanceof ManagedWebGLRenderingContext ? context : new ManagedWebGLRenderingContext(context);
		this.context.addRestorable(this);
		this.compile();
	}

	private compile() {
		let gl = this.context.gl;
		try {
			this.vs = this.compileShader(gl.VERTEX_SHADER, this.vertexShader);
			this.fs = this.compileShader(gl.FRAGMENT_SHADER, this.fragmentShader);
			this.program = this.compileProgram(this.vs, this.fs);
		} catch (e) {
			console.log("Shader compilation error", e);
			this.dispose();
			throw e;
		}
	}

	private compileShader(type: number, source: string) {
		let gl = this.context.gl;
		let shader = gl.createShader(type);
		gl.shaderSource(shader, source);
		gl.compileShader(shader);
		if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
			let error = "Couldn't compile shader: " + gl.getShaderInfoLog(shader);
			gl.deleteShader(shader);
			if (!gl.isContextLost()) throw new Error(error);
		}
		return shader;
	}

	private compileProgram(vs: WebGLShader, fs: WebGLShader) {
		let gl = this.context.gl;
		let program = gl.createProgram();
		gl.attachShader(program, vs);
		gl.attachShader(program, fs);
		gl.linkProgram(program);

		if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
			let error = "Couldn't compile shader program: " + gl.getProgramInfoLog(program);
			gl.deleteProgram(program);
			if (!gl.isContextLost()) throw new Error(error);
		}
		return program;
	}

	restore() {
		this.compile();
	}

	public bind() {
		this.context.gl.useProgram(this.program);
	}

	public unbind() {
		this.context.gl.useProgram(null);
	}

	public setUniformi(uniform: string, value: number) {
		this.context.gl.uniform1i(this.getUniformLocation(uniform), value);
	}

	public setUniformf(uniform: string, value: number) {
		this.context.gl.uniform1f(this.getUniformLocation(uniform), value);
	}

	public setUniform2f(uniform: string, value: number, value2: number) {
		this.context.gl.uniform2f(this.getUniformLocation(uniform), value, value2);
	}

	public setUniform3f(uniform: string, value: number, value2: number, value3: number) {
		this.context.gl.uniform3f(this.getUniformLocation(uniform), value, value2, value3);
	}

	public setUniform4f(uniform: string, value: number, value2: number, value3: number, value4: number) {
		this.context.gl.uniform4f(this.getUniformLocation(uniform), value, value2, value3, value4);
	}

	public setUniform2x2f(uniform: string, value: ArrayLike<number>) {
		let gl = this.context.gl;
		this.tmp2x2.set(value);
		gl.uniformMatrix2fv(this.getUniformLocation(uniform), false, this.tmp2x2);
	}

	public setUniform3x3f(uniform: string, value: ArrayLike<number>) {
		let gl = this.context.gl;
		this.tmp3x3.set(value);
		gl.uniformMatrix3fv(this.getUniformLocation(uniform), false, this.tmp3x3);
	}

	public setUniform4x4f(uniform: string, value: ArrayLike<number>) {
		let gl = this.context.gl;
		this.tmp4x4.set(value);
		gl.uniformMatrix4fv(this.getUniformLocation(uniform), false, this.tmp4x4);
	}

	public getUniformLocation(uniform: string): WebGLUniformLocation {
		let gl = this.context.gl;
		let location = gl.getUniformLocation(this.program, uniform);
		if (!location && !gl.isContextLost()) throw new Error(`Couldn't find location for uniform ${uniform}`);
		return location;
	}

	public getAttributeLocation(attribute: string): number {
		let gl = this.context.gl;
		let location = gl.getAttribLocation(this.program, attribute);
		if (location == -1 && !gl.isContextLost()) throw new Error(`Couldn't find location for attribute ${attribute}`);
		return location;
	}

	public dispose() {
		this.context.removeRestorable(this);

		let gl = this.context.gl;
		if (this.vs) {
			gl.deleteShader(this.vs);
			this.vs = null;
		}

		if (this.fs) {
			gl.deleteShader(this.fs);
			this.fs = null;
		}

		if (this.program) {
			gl.deleteProgram(this.program);
			this.program = null;
		}
	}

	public static newColoredTextured(context: ManagedWebGLRenderingContext | WebGL2RenderingContext): Shader {
		let vs = `
		bad
				attribute vec4 ${Shader.POSITION};
				attribute vec4 ${Shader.COLOR};
				attribute vec2 ${Shader.TEXCOORDS};
				uniform mat4 ${Shader.MVP_MATRIX};
				varying vec4 v_color;
				varying vec2 v_texCoords;

				void main () {
					v_color = ${Shader.COLOR};
					v_texCoords = ${Shader.TEXCOORDS};
					gl_Position = ${Shader.MVP_MATRIX} * ${Shader.POSITION};
				}
			`;

		let fs = `
		bad
				#ifdef GL_ES
					#define LOWP lowp
					precision mediump float;
				#else
					#define LOWP
				#endif
				varying LOWP vec4 v_color;
				varying vec2 v_texCoords;
				uniform sampler2D u_texture;

				void main () {
					gl_FragColor = v_color * texture(u_texture, v_texCoords);
				}
			`;

		return new Shader(context, vs, fs);
	}

	public static newTwoColoredTextured(context: ManagedWebGLRenderingContext | WebGL2RenderingContext): Shader {
		let vs = `#version 300 es
				layout(location = 0) in vec2 ${Shader.POSITION};
				layout(location = 1) in vec4 ${Shader.COLOR};
				layout(location = 2) in vec4 ${Shader.COLOR2};
				layout(location = 3) in vec2 ${Shader.TEXCOORDS};
				layout(location = 4) in vec2 ${Shader.TEXTURE_INDEX_POS_Z};
				uniform mat4 ${Shader.MVP_MATRIX};

				out vec4 v_light;
				out vec4 v_dark;
				out vec2 v_texCoords;
				out float v_texIndex;

				void main () {
					v_light = ${Shader.COLOR};
					v_dark = ${Shader.COLOR2};
					v_texCoords = ${Shader.TEXCOORDS};
					v_texIndex = ${Shader.TEXTURE_INDEX_POS_Z}.x;

					vec4 pos = vec4(
						${Shader.POSITION}.x,
						${Shader.POSITION}.y,
						${Shader.TEXTURE_INDEX_POS_Z}.y,
						1.0
					);
					gl_Position = ${Shader.MVP_MATRIX} * pos;
				}
			`;

		let fs = `#version 300 es
				#ifdef GL_ES
					#define LOWP lowp
					precision mediump float;
				#else
					#define LOWP
				#endif
				in LOWP vec4 v_light;
				in LOWP vec4 v_dark;
				in vec2 v_texCoords;
				in float v_texIndex;
				out vec4 FragColor;
				uniform sampler2D u_textures[16];

				vec4 getTextureColor() {
					int index = int(floor(v_texIndex + 0.2));
					if (index < 8) {
						if (index < 4) {
							if (index < 2) {
								return (index == 0) 
									? texture(u_textures[0], v_texCoords) 
									: texture(u_textures[1], v_texCoords);
							} else {
								return (index == 2) 
								? texture(u_textures[2], v_texCoords) 
								: texture(u_textures[3], v_texCoords);
							}
						} else {
							if (index < 6) {
								return (index == 4) 
									? texture(u_textures[4], v_texCoords) 
									: texture(u_textures[5], v_texCoords);
							} else {
								return (index == 6) 
								? texture(u_textures[6], v_texCoords) 
								: texture(u_textures[7], v_texCoords);
							}
						}
					} else {
						if (index < 12) {
							if (index < 10) {
								return (index == 8) 
									? texture(u_textures[8], v_texCoords) 
									: texture(u_textures[9], v_texCoords);
							} else {
								return (index == 10) 
								? texture(u_textures[10], v_texCoords) 
								: texture(u_textures[11], v_texCoords);
							}
						} else {
							if (index < 14) {
								return (index == 12) 
									? texture(u_textures[12], v_texCoords) 
									: texture(u_textures[13], v_texCoords);
							} else {
								return (index == 14) 
								? texture(u_textures[14], v_texCoords) 
								: texture(u_textures[15], v_texCoords);
							}
						}
					}
					
					return vec4(1.0, 0.0, 0.0, 1.0); // Fallback for invalid indices
				}

				void main () {
					vec4 texColor = getTextureColor();
					FragColor.a = texColor.a * v_light.a;
					FragColor.rgb = ((texColor.a - 1.0) * v_dark.a + 1.0 - texColor.rgb) * v_dark.rgb + texColor.rgb * v_light.rgb;
				}
			`;

		return new Shader(context, vs, fs);
	}

	public static newTwoColoredTexturedWithTextureVerts(context: ManagedWebGLRenderingContext | WebGL2RenderingContext): Shader {
		let vs = `#version 300 es
				layout(location = 0) in int u_input_coords;
				uniform mat4 ${Shader.MVP_MATRIX};
				uniform sampler2D ${Shader.TEXTURE_VERTS};

				out vec4 v_light;
				out vec4 v_dark;
				out vec2 v_texCoords;
				out float v_texIndex;

				struct VertexData {
					vec3 position;
					vec4 color1;
					vec4 color2;
					vec2 uv;
					float textureIndex;
				};
				VertexData vd;

				void fillVertexDataByIndex(int index) {
					// x4 because we align to 4 byte/pixel boundaries, and the 
					// vertex has 14 components
					index = index * 4;
					
					int texWidth = textureSize(${Shader.TEXTURE_VERTS}, 0).x;
					int col = index % texWidth;
					int row = index / texWidth;
					vec4 v = texelFetch(${Shader.TEXTURE_VERTS}, ivec2(col, row), 0);
					vd.position = v.xyz;
					vd.textureIndex = v.w;

					int col2 = (col + 1 >= texWidth) ? 0 : col + 1;
					int row2 = (col + 1 >= texWidth) ? row + 1 : row;
					v = texelFetch(${Shader.TEXTURE_VERTS}, ivec2(col2, row2), 0);
					vd.color1 = v;

					col2 = (col + 2 >= texWidth) ? 1 : col + 2;
					row2 = (col + 2 >= texWidth) ? row + 1 : row;
					v = texelFetch(${Shader.TEXTURE_VERTS}, ivec2(col2, row2), 0);
					vd.color2 = v;

					col2 = (col + 3 >= texWidth) ? 2 : col + 3;
					row2 = (col + 3 >= texWidth) ? row + 1 : row;
					v = texelFetch(${Shader.TEXTURE_VERTS}, ivec2(col2, row2), 0);
					vd.uv = v.xy;
				}

				void main () {
					fillVertexDataByIndex(u_input_coords);

					v_light = vd.color1;
					v_dark = vd.color2;
					v_texCoords = vd.uv;
					v_texIndex = vd.textureIndex;
					vec4 pos = vec4(vd.position, 1.0);

					gl_Position = ${Shader.MVP_MATRIX} * pos;
				}
			`;

		let fs = `#version 300 es
				#ifdef GL_ES
					#define LOWP lowp
					precision mediump float;
				#else
					#define LOWP
				#endif
				in LOWP vec4 v_light;
				in LOWP vec4 v_dark;
				in vec2 v_texCoords;
				in float v_texIndex;
				out vec4 FragColor;
				uniform sampler2D u_textures[8];

				vec4 getTextureColor() {
					int index = int(floor(v_texIndex + 0.2));
					if (index < 8) {
						if (index < 4) {
							if (index < 2) {
								return (index == 0) 
									? texture(u_textures[0], v_texCoords) 
									: texture(u_textures[1], v_texCoords);
							} else {
								return (index == 2) 
								? texture(u_textures[2], v_texCoords) 
								: texture(u_textures[3], v_texCoords);
							}
						} else {
							if (index < 6) {
								return (index == 4) 
									? texture(u_textures[4], v_texCoords) 
									: texture(u_textures[5], v_texCoords);
							} else {
								return (index == 6) 
								? texture(u_textures[6], v_texCoords) 
								: texture(u_textures[7], v_texCoords);
							}
						}
					}
					
					return vec4(1.0, 0.0, 0.0, 1.0); // Fallback for invalid indices
				}

				void main () {
					vec4 texColor = getTextureColor();
					FragColor.a = texColor.a * v_light.a;
					FragColor.rgb = ((texColor.a - 1.0) * v_dark.a + 1.0 - texColor.rgb) * v_dark.rgb + texColor.rgb * v_light.rgb;
				}
			`;

		return new Shader(context, vs, fs);
	}

	public static newColored(context: ManagedWebGLRenderingContext | WebGL2RenderingContext): Shader {
		let vs = `#version 300 es
				layout (location = 0) in vec3 ${Shader.POSITION};
				layout (location = 1) in vec4 ${Shader.COLOR};
				uniform mat4 ${Shader.MVP_MATRIX};
				out vec4 v_color;

				void main () {
					v_color = ${Shader.COLOR};
					gl_Position = ${Shader.MVP_MATRIX} * vec4(${Shader.POSITION}, 1.0);
				}
			`;

		let fs = `#version 300 es
				#ifdef GL_ES
					#define LOWP lowp
					precision mediump float;
				#else
					#define LOWP
				#endif
				in LOWP vec4 v_color;
				out vec4 FragColor;

				void main () {
					FragColor = v_color;
				}
			`;

		return new Shader(context, vs, fs);
	}
}
