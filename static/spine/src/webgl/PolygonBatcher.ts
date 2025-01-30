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

import { GLTexture } from "./GLTexture";
import { Mesh, Position2Attribute, ColorAttribute, TexCoordAttribute, Color2Attribute } from "./Mesh";
import { Shader } from "./Shader";
import { ManagedWebGLRenderingContext } from "./WebGL";
import { Disposable } from "../core/Utils";
import { Camera } from "./Camera";

export class PolygonBatcher implements Disposable {
	// 16 is the minimum supported by openGL. 
	// Might be better to directly query the parameter and then use that value.
	// But that would require us to dynamically generate the fragment shader
	static MAX_LAST_TEXTURES = 16;  

	private context: ManagedWebGLRenderingContext;
	private drawCalls: number;
	private isDrawing = false;
	private mesh: Mesh;
	private shader: Shader = null;
	private verticesLength = 0;
	private indicesLength = 0;
	private texturesLength = 0;
	private srcBlend: number;
	private dstBlend: number;

	private lastTextures: Map<number, GLTexture>;
	private lastTextureIndex: Map<number, number>;
	private lastTextureIndexCount: number;
	private totalNumberTris: number = 0
	private totalNumberVertices: number = 0;

	constructor(
		context: ManagedWebGLRenderingContext | WebGL2RenderingContext,
		twoColorTint: boolean = true,
		maxVertices: number = 10920
	) {
		if (maxVertices > 10920) throw new Error("Can't have more than 10920 triangles per batch: " + maxVertices);
		this.context = context instanceof ManagedWebGLRenderingContext ? context : new ManagedWebGLRenderingContext(context);
		let attributes = twoColorTint ?
			[new Position2Attribute(), new ColorAttribute(), new TexCoordAttribute(), new Color2Attribute()] :
			[new Position2Attribute(), new ColorAttribute(), new TexCoordAttribute()];
		this.mesh = new Mesh(context, attributes, maxVertices, maxVertices * 3);
		this.shader = twoColorTint ? Shader.newTwoColoredTextured(this.context) : Shader.newColoredTextured(this.context);
		this.srcBlend = this.context.gl.SRC_ALPHA;
		this.dstBlend = this.context.gl.ONE_MINUS_SRC_ALPHA;
	}

	prep() {
		this.shader.bind();
		this.mesh.prep(this.shader);
		let gl = this.context.gl;
		gl.enable(gl.BLEND);
		gl.blendFunc(this.srcBlend, this.dstBlend);
	}

	use(camera: Camera) {
		if (this.isDrawing) throw new Error("PolygonBatch is already drawing. Call PolygonBatch.end() before calling PolygonBatch.begin()");
		this.drawCalls = 0;
		this.lastTextures = new Map<number, GLTexture>();
		this.lastTextureIndex = new Map<number, number>();
		this.lastTextureIndexCount = 0;
		this.isDrawing = true;

		this.shader.setUniform4x4f(
			Shader.MVP_MATRIX,
			camera.projectionView.values
		);
	}

	finish() {
		if (!this.isDrawing) throw new Error("PolygonBatch is not drawing. Call PolygonBatch.begin() before calling PolygonBatch.end()");
		if (this.verticesLength > 0 || this.indicesLength > 0) {
			this.flush();
		}
		this.isDrawing = false;
	}

	setBlendMode(srcBlend: number, dstBlend: number) {
		let gl = this.context.gl;
		this.srcBlend = srcBlend;
		this.dstBlend = dstBlend;
		if (this.isDrawing) {
			this.flush();
			gl.blendFunc(this.srcBlend, this.dstBlend);
		}
	}

	draw(texture: GLTexture, vertices: ArrayLike<number>, indices: Array<number>, positionZ: number = 0.0) {
		if (this.verticesLength + vertices.length > this.mesh.getVertices().length ||
			this.indicesLength + indices.length > this.mesh.getIndices().length) {
			this.flush();
		} else if (this.lastTextures.size >= PolygonBatcher.MAX_LAST_TEXTURES) {
			this.flush();
		}

		let indexStart = this.mesh.numVertices();
		this.mesh.getVertices().set(vertices, this.verticesLength);
		this.verticesLength += vertices.length;
		this.mesh.setVerticesLength(this.verticesLength)

		let indicesArray = this.mesh.getIndices();
		for (let i = this.indicesLength, j = 0; j < indices.length; i++, j++)
			indicesArray[i] = indices[j] + indexStart;
		this.indicesLength += indices.length;
		this.mesh.setIndicesLength(this.indicesLength);

		// Attach the textures/positions to the mesh
		if (!this.lastTextures.has(texture.getID())) {
			this.lastTextures.set(texture.getID(), texture);
			this.lastTextureIndex.set(texture.getID(), this.lastTextureIndexCount);
			this.lastTextureIndexCount++;
		}
		let textureIndexToUse = this.lastTextureIndex.get(texture.getID());
		let texturesArray = this.mesh.getTexturesPos();
		let numVerts = vertices.length / this.mesh.getVertexSizeInFloats();
		let textures = [];
		for (let i = 0; i < numVerts; i++) {
			textures.push(textureIndexToUse);
			textures.push(positionZ);
		};
		for (let i = this.texturesLength, j = 0; j < textures.length; i++, j++) {
			texturesArray[i] = textures[j]
		}
		this.texturesLength += textures.length;
		this.mesh.setTexturesPosLength(this.texturesLength);
	}

	public clearTotals() {
		this.totalNumberTris = 0;
		this.totalNumberVertices = 0;
	}
	public printTotals() {
		console.log("Total number tris: ", this.totalNumberTris);
		console.log("Total number vertices: ", this.totalNumberVertices);
	}

	private flush() {
		let gl = this.context.gl;
		if (this.verticesLength == 0) return;

		for (let [key, value] of this.lastTextures) {
			let index = this.lastTextureIndex.get(key);
			value.bind(index);
			this.shader.setUniformi(Shader.SAMPLER + "[" + index + ']', index);
		}
		this.mesh.draw(this.shader, gl.TRIANGLES);
		this.totalNumberTris += this.indicesLength / 3;
		this.totalNumberVertices += this.verticesLength / this.mesh.getVertexSizeInFloats();

		this.verticesLength = 0;
		this.indicesLength = 0;
		this.texturesLength = 0;

		this.lastTextures.clear();
		this.lastTextureIndex.clear();
		this.lastTextureIndexCount = 0;

		this.mesh.setVerticesLength(0);
		this.mesh.setIndicesLength(0);
		this.mesh.setTexturesPosLength(0);
		this.drawCalls++;
	}

	getDrawCalls() { return this.drawCalls; }

	dispose() {
		this.mesh.dispose();
		this.shader.dispose();
	}
}
