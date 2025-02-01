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

import { Restorable, Disposable, Utils } from "../core/Utils";
import { DataTexture } from "./DataTexture";
import { VertexAttribute } from "./Mesh";
import { Shader } from "./Shader";
import { ManagedWebGLRenderingContext } from "./WebGL";

export class MeshTextureVerts implements Disposable, Restorable {
	private context: ManagedWebGLRenderingContext;

	private dataTexture: DataTexture = null;
	public dataTextureBuffer: Float32Array  = null;
	public dateTexturebufferCurrentLength: number;

	private textureVertsBuffer: WebGLBuffer;


	private vertices: Float32Array;
	// private verticesBuffer: WebGLBuffer;
	private verticesLength = 0;
	private dirtyVertices = false;

	private indices: Uint16Array;
	private indicesBuffer: WebGLBuffer;
	private indicesLength = 0;
	private dirtyIndices = false;

	private texturePos: Float32Array;
	// private texturesPosBuffer: WebGLBuffer;
	private texturesPosLength = 0;
	private dirtyTexturesPos = false;

	private vao: WebGLVertexArrayObject;
	private elementsPerVertex = 0;

	getAttributes(): VertexAttribute[] { return this.attributes; }

	maxVertices(): number { return this.vertices.length / this.elementsPerVertex; }
	numVertices(): number { return this.verticesLength / this.elementsPerVertex; }
	setVerticesLength(length: number) {
		this.dirtyVertices = true;
		this.verticesLength = length;
	}
	getVertices(): Float32Array { return this.vertices; }

	maxIndices(): number { return this.indices.length; }
	numIndices(): number { return this.indicesLength; }
	setIndicesLength(length: number) {
		this.dirtyIndices = true;
		this.indicesLength = length;
	}
	getIndices(): Uint16Array { return this.indices };

	public enabletexturesPos: boolean = true;
	maxTexturesPos(): number { return this.texturePos.length; }
	numTexturesPos(): number { return this.texturesPosLength; }
	setTexturesPosLength(length: number) {
		this.dirtyTexturesPos = true;
		this.texturesPosLength = length;
	}
	getTexturesPos(): Float32Array { return this.texturePos };

	getVertexSizeInFloats(): number {
		let size = 0;
		for (var i = 0; i < this.attributes.length; i++) {
			let attribute = this.attributes[i];
			size += attribute.numElements;
		}
		return size;
	}

	constructor(
		context: ManagedWebGLRenderingContext | WebGL2RenderingContext,
		private attributes: VertexAttribute[],
		maxVertices: number,
		maxIndices: number) {
		this.context = context instanceof ManagedWebGLRenderingContext ? context : new ManagedWebGLRenderingContext(context);
		this.elementsPerVertex = 0;
		for (let i = 0; i < attributes.length; i++) {
			this.elementsPerVertex += attributes[i].numElements;
		}
		this.vertices = new Float32Array(maxVertices * this.elementsPerVertex);
		this.indices = new Uint16Array(maxIndices);
		this.texturePos = new Float32Array(maxVertices * 2)

		this.dataTexture = new DataTexture(this.context, maxVertices, 3 + 4 + 4 + 3);
		this.dataTextureBuffer = new Float32Array(this.dataTexture.widthPx * this.dataTexture.heightPx * (3+4+4+3));
		this.dateTexturebufferCurrentLength = 0;

		// this.positions = new DataTexture(this.context, maxVertices, 3);
		// this.color1 = new DataTexture(this.context, maxVertices, 4);
		// this.color2 = new DataTexture(this.context, maxVertices, 4);
		// this.texCoordsAndIndex = new DataTexture(this.context, maxVertices, 3);
		this.context.addRestorable(this);
	}

	setVertices(vertices: Array<number>) {
		this.dirtyVertices = true;
		if (vertices.length > this.vertices.length) throw Error("Mesh can't store more than " + this.maxVertices() + " vertices");
		this.vertices.set(vertices, 0);
		this.verticesLength = vertices.length;
	}

	setIndices(indices: Array<number>) {
		this.dirtyIndices = true;
		if (indices.length > this.indices.length) throw Error("Mesh can't store more than " + this.maxIndices() + " indices");
		this.indices.set(indices, 0);
		this.indicesLength = indices.length;
	}

	setTexturesPos(texturesPos: Array<number>) {
		this.dirtyTexturesPos = true;
		if (texturesPos.length > this.texturePos.length) throw Error("Mesh can't store more than " + this.maxTexturesPos() + " texturesPos");
		this.texturePos.set(texturesPos, 0);
		this.texturesPosLength = texturesPos.length;
	}

	draw(shader: Shader, primitiveType: number) {
		this.drawWithOffset(shader, primitiveType, 0,
			this.indicesLength > 0 ? this.indicesLength : this.verticesLength / this.elementsPerVertex);
	}

	drawWithOffset(shader: Shader, primitiveType: number, offset: number, count: number) {
		let gl = this.context.gl;
		this.use(shader);
		if (this.dirtyVertices || this.dirtyIndices || this.dirtyTexturesPos) this.update();
		if (this.indicesLength > 0) {
			gl.drawElements(primitiveType, count, gl.UNSIGNED_SHORT, offset * 2);
		} else {
			gl.drawArrays(primitiveType, offset, count);
		}
		this.finish();
	}

	public prep(shader: Shader) {
		let gl = this.context.gl;
		if (this.vao == null) {
			this.vao = gl.createVertexArray();
			gl.bindVertexArray(this.vao);

			this.textureVertsBuffer = gl.createBuffer();
			this.indicesBuffer = gl.createBuffer();

			gl.bindBuffer(gl.ARRAY_BUFFER, this.textureVertsBuffer);
			gl.bufferData(gl.ARRAY_BUFFER, this.maxVertices()*4, gl.DYNAMIC_DRAW)

			gl.enableVertexAttribArray(0);
			gl.vertexAttribIPointer(0, 1, gl.INT, 0, 0);

			gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.indicesBuffer);
			gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, this.indices.length*2, gl.DYNAMIC_DRAW)
		}
		// gl.bindVertexArray(this.vao);
	}

	use(shader: Shader) {
		let gl = this.context.gl;
		gl.bindVertexArray(this.vao);

		this.dataTexture.bind(10);
		shader.setUniformi(Shader.TEXTURE_VERTS, 10);
	}

	finish() {
		let gl = this.context.gl;
		gl.bindVertexArray(null);
	}

	interleaveData() {
		// vertices data format:
		// 0, 1,      2,3,4,5,   6,7,  8,9,10,11
		// x,y        color1     uv,   color2
		// textrePos data format:
		// 0, 1
		// z, textureIndex
		let output = this.dataTextureBuffer;

		const numVertices = this.verticesLength / this.elementsPerVertex;
		let positionStride = this.elementsPerVertex;
		let texturePosStride = 2;
		const neededSizeNumFloats = numVertices * (positionStride + texturePosStride);
		if (output.length < neededSizeNumFloats) {
			throw new Error("Output array is too small to hold interleaved data. Need: " + neededSizeNumFloats + ", got: " + output.length);
		}
		

		let oi = 0;
		for (let i = 0; i < numVertices; i++) {
			const posOffset = i * positionStride;
			const tOffset = i * texturePosStride;
			output[oi++] = this.vertices[posOffset + 0]; // x
			output[oi++] = this.vertices[posOffset + 1]; // y
			output[oi++] = this.texturePos[tOffset + 1]; // z
			output[oi++] = this.texturePos[tOffset + 0]; // texture_index

			output[oi++] = this.vertices[posOffset + 2]; // r1
			output[oi++] = this.vertices[posOffset + 3]; // g1
			output[oi++] = this.vertices[posOffset + 4]; // b1
			output[oi++] = this.vertices[posOffset + 5]; // a1

			output[oi++] = this.vertices[posOffset + 8]; // r2
			output[oi++] = this.vertices[posOffset + 9]; // g2
			output[oi++] = this.vertices[posOffset + 10]; // b2
			output[oi++] = this.vertices[posOffset + 11]; // a2

			output[oi++] = this.vertices[posOffset + 6]; // u1  (texture)
			output[oi++] = this.vertices[posOffset + 7]; // v1
		}
		return neededSizeNumFloats;
	}


	private update() {
		let gl = this.context.gl;

		const numVertices = this.verticesLength / this.elementsPerVertex;
		if (numVertices != this.texturesPosLength / 2) throw Error("vertices and texturePos must have the same length");

		const bufferLength = this.interleaveData();
		this.dataTexture.fillData(this.dataTextureBuffer, bufferLength);

		let attributeBuffer = new Int32Array(numVertices);
		for (let i = 0; i < attributeBuffer.length; i++) {
			attributeBuffer[i] = i;
		}
		gl.bindBuffer(gl.ARRAY_BUFFER, this.textureVertsBuffer);
		gl.bufferSubData(gl.ARRAY_BUFFER, 0, attributeBuffer);

		gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.indicesBuffer);
		gl.bufferSubData( gl.ELEMENT_ARRAY_BUFFER, 0, this.indices, 0, this.indicesLength);

		this.dirtyIndices = false;
		this.dirtyVertices = false;
	}

	restore() {
		this.indicesBuffer = null;
		this.textureVertsBuffer = null;
		this.vao = null;
		this.dataTexture.restore();

		// TODO: restore for DataTextures
		this.update();
	}

	dispose() {
		this.context.removeRestorable(this);
		let gl = this.context.gl;

		this.dataTexture.dispose();
		gl.deleteBuffer(this.indicesBuffer);
		gl.deleteBuffer(this.textureVertsBuffer);
		gl.deleteVertexArray(this.vao);

		this.indicesBuffer = null;
		this.textureVertsBuffer = null;
		this.vao = null;
	}
}
