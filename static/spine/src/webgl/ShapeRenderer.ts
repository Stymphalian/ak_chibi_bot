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

import { Color, Vector2, MathUtils, Disposable } from "../core/Utils";
import { Camera } from "./Camera";
import { Mesh, Position2Attribute, ColorAttribute, Position3Attribute } from "./Mesh";
import { Shader } from "./Shader";
import { ManagedWebGLRenderingContext } from "./WebGL";

// TODO: Make the 'z' parameters for all the draw functions "better". It shouldn't
// be put at the end of the argument list, and it should be be specifed for every 
// coordinate value instead of just the a single z value for all points.

export class ShapeRenderer implements Disposable {
	private context: ManagedWebGLRenderingContext;
	private isDrawing = false;
	private mesh: Mesh;
	private shapeType = ShapeType.Filled;
	private color = new Color(1, 1, 1, 1);
	private shader: Shader;
	private vertexIndex = 0;
	private tmp = new Vector2();
	private srcBlend: number;
	private dstBlend: number;

	constructor(context: ManagedWebGLRenderingContext | WebGL2RenderingContext, maxVertices: number = 10920) {
		if (maxVertices > 10920) throw new Error("Can't have more than 10920 triangles per batch: " + maxVertices);
		this.context = context instanceof ManagedWebGLRenderingContext ? context : new ManagedWebGLRenderingContext(context);
		this.mesh = new Mesh(context, [new Position3Attribute(), new ColorAttribute()], maxVertices, 0);
		this.mesh.enabletexturesPos = false;
		this.srcBlend = this.context.gl.SRC_ALPHA;
		this.dstBlend = this.context.gl.ONE_MINUS_SRC_ALPHA;
		this.shader = Shader.newColored(this.context);
	}

	prep() {
		this.shader.bind();
		this.mesh.prep(this.shader);
		let gl = this.context.gl;
		gl.enable(this.context.gl.BLEND);
		gl.blendFunc(this.srcBlend, this.dstBlend);
	}

	// begin(shader: Shader) {
	use(camera: Camera) {
		if (this.isDrawing) throw new Error("ShapeRenderer.begin() has already been called");
		this.vertexIndex = 0;
		this.isDrawing = true;

		this.shader.setUniform4x4f(
			Shader.MVP_MATRIX,
			camera.projectionView.values
		);
	}

	finish() {
		if (!this.isDrawing) throw new Error("ShapeRenderer.begin() has not been called");
		this.flush();
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

	setColor(color: Color) {
		this.color.setFromColor(color);
	}

	setColorWith(r: number, g: number, b: number, a: number) {
		this.color.set(r, g, b, a);
	}

	point(x: number, y: number, color: Color = null, z: number) {
		this.check(ShapeType.Point, 1);
		if (color === null) color = this.color;
		this.vertex(x, y, z, color);
	}

	line(x: number, y: number, x2: number, y2: number, color: Color = null, z: number=0) {
		this.check(ShapeType.Line, 2);
		let vertices = this.mesh.getVertices();
		let idx = this.vertexIndex;
		if (color === null) color = this.color;
		this.vertex(x, y, z, color);
		this.vertex(x2, y2, z, color);
	}

	triangle(filled: boolean, x: number, y: number, x2: number, y2: number, x3: number, y3: number, color: Color = null, color2: Color = null, color3: Color = null, z: number = 0) {
		this.check(filled ? ShapeType.Filled : ShapeType.Line, 3);
		let vertices = this.mesh.getVertices();
		let idx = this.vertexIndex;
		if (color === null) color = this.color;
		if (color2 === null) color2 = this.color;
		if (color3 === null) color3 = this.color;
		if (filled) {
			this.vertex(x, y, z, color);
			this.vertex(x2, y2, z, color2);
			this.vertex(x3, y3, z, color3);
		} else {
			this.vertex(x, y, z, color);
			this.vertex(x2, y2, z, color2);

			this.vertex(x2, y2, z, color);
			this.vertex(x3, y3, z, color2);

			this.vertex(x3, y3, z, color);
			this.vertex(x, y, z, color2);
		}
	}

	quad(filled: boolean, x: number, y: number, x2: number, y2: number, x3: number, y3: number, x4: number, y4: number, color: Color = null, color2: Color = null, color3: Color = null, color4: Color = null, z: number = 0) {
		this.check(filled ? ShapeType.Filled : ShapeType.Line, 3);
		let vertices = this.mesh.getVertices();
		let idx = this.vertexIndex;
		if (color === null) color = this.color;
		if (color2 === null) color2 = this.color;
		if (color3 === null) color3 = this.color;
		if (color4 === null) color4 = this.color;
		if (filled) {
			this.vertex(x, y, z, color); this.vertex(x2, y2, z, color2); this.vertex(x3, y3, z, color3);
			this.vertex(x3, y3, z, color3); this.vertex(x4, y4, z, color4); this.vertex(x, y, z, color);
		} else {
			this.vertex(x, y, z, color); this.vertex(x2, y2, z, color2);
			this.vertex(x2, y2, z, color2); this.vertex(x3, y3, z, color3);
			this.vertex(x3, y3, z, color3); this.vertex(x4, y4, z, color4);
			this.vertex(x4, y4, z, color4); this.vertex(x, y, z, color);
		}
	}

	rect(filled: boolean, x: number, y: number, width: number, height: number, color: Color = null, z:number = 0) {
		this.quad(filled, x, y, x + width, y, x + width, y + height, x, y + height, color, color, color, color, z);
	}

	rectLine(filled: boolean, x1: number, y1: number, x2: number, y2: number, width: number, color: Color = null, z: number = 0) {
		this.check(filled ? ShapeType.Filled : ShapeType.Line, 8);
		if (color === null) color = this.color;
		let t = this.tmp.set(y2 - y1, x1 - x2);
		t.normalize();
		width *= 0.5;
		let tx = t.x * width;
		let ty = t.y * width;
		if (!filled) {
			this.vertex(x1 + tx, y1 + ty, z, color);
			this.vertex(x1 - tx, y1 - ty, z, color);
			this.vertex(x2 + tx, y2 + ty, z, color);
			this.vertex(x2 - tx, y2 - ty, z, color);

			this.vertex(x2 + tx, y2 + ty, z, color);
			this.vertex(x1 + tx, y1 + ty, z, color);

			this.vertex(x2 - tx, y2 - ty, z, color);
			this.vertex(x1 - tx, y1 - ty, z, color);
		} else {
			this.vertex(x1 + tx, y1 + ty, z, color);
			this.vertex(x1 - tx, y1 - ty, z, color);
			this.vertex(x2 + tx, y2 + ty, z, color);

			this.vertex(x2 - tx, y2 - ty, z, color);
			this.vertex(x2 + tx, y2 + ty, z, color);
			this.vertex(x1 - tx, y1 - ty, z, color);
		}
	}

	x(x: number, y: number, size: number) {
		this.line(x - size, y - size, x + size, y + size, null, 0);
		this.line(x - size, y + size, x + size, y - size, null, 0);
	}

	polygon(polygonVertices: ArrayLike<number>, offset: number, count: number, color: Color = null, z:number = 0) {
		if (count < 3) throw new Error("Polygon must contain at least 3 vertices");
		this.check(ShapeType.Line, count * 2);
		if (color === null) color = this.color;
		let vertices = this.mesh.getVertices();
		let idx = this.vertexIndex;

		offset <<= 1;
		count <<= 1;

		let firstX = polygonVertices[offset];
		let firstY = polygonVertices[offset + 1];
		let last = offset + count;

		for (let i = offset, n = offset + count - 2; i < n; i += 2) {
			let x1 = polygonVertices[i];
			let y1 = polygonVertices[i + 1];

			let x2 = 0;
			let y2 = 0;

			if (i + 2 >= last) {
				x2 = firstX;
				y2 = firstY;
			} else {
				x2 = polygonVertices[i + 2];
				y2 = polygonVertices[i + 3];
			}

			this.vertex(x1, y1, z, color);
			this.vertex(x2, y2, z, color);
		}
	}

	circle(filled: boolean, x: number, y: number, radius: number, color: Color = null, segments: number = 0, z: number = 0) {
		if (segments === 0) segments = Math.max(1, (6 * MathUtils.cbrt(radius)) | 0);
		if (segments <= 0) throw new Error("segments must be > 0.");
		if (color === null) color = this.color;
		let angle = 2 * MathUtils.PI / segments;
		let cos = Math.cos(angle);
		let sin = Math.sin(angle);
		let cx = radius, cy = 0;
		if (!filled) {
			this.check(ShapeType.Line, segments * 2 + 2);
			for (let i = 0; i < segments; i++) {
				this.vertex(x + cx, y + cy, z, color);
				let temp = cx;
				cx = cos * cx - sin * cy;
				cy = sin * temp + cos * cy;
				this.vertex(x + cx, y + cy, z, color);
			}
			// Ensure the last segment is identical to the first.
			this.vertex(x + cx, y + cy, z, color);
		} else {
			this.check(ShapeType.Filled, segments * 3 + 3);
			segments--;
			for (let i = 0; i < segments; i++) {
				this.vertex(x, y, z, color);
				this.vertex(x + cx, y + cy, z, color);
				let temp = cx;
				cx = cos * cx - sin * cy;
				cy = sin * temp + cos * cy;
				this.vertex(x + cx, y + cy, z, color);
			}
			// Ensure the last segment is identical to the first.
			this.vertex(x, y, z, color);
			this.vertex(x + cx, y + cy, z, color);
		}

		let temp = cx;
		cx = radius;
		cy = 0;
		this.vertex(x + cx, y + cy, z, color);
	}

	curve(x1: number, y1: number, cx1: number, cy1: number, cx2: number, cy2: number, x2: number, y2: number, segments: number, color: Color = null, z: number = 0) {
		this.check(ShapeType.Line, segments * 2 + 2);
		if (color === null) color = this.color;

		// Algorithm from: http://www.antigrain.com/research/bezier_interpolation/index.html#PAGE_BEZIER_INTERPOLATION
		let subdiv_step = 1 / segments;
		let subdiv_step2 = subdiv_step * subdiv_step;
		let subdiv_step3 = subdiv_step * subdiv_step * subdiv_step;

		let pre1 = 3 * subdiv_step;
		let pre2 = 3 * subdiv_step2;
		let pre4 = 6 * subdiv_step2;
		let pre5 = 6 * subdiv_step3;

		let tmp1x = x1 - cx1 * 2 + cx2;
		let tmp1y = y1 - cy1 * 2 + cy2;

		let tmp2x = (cx1 - cx2) * 3 - x1 + x2;
		let tmp2y = (cy1 - cy2) * 3 - y1 + y2;

		let fx = x1;
		let fy = y1;

		let dfx = (cx1 - x1) * pre1 + tmp1x * pre2 + tmp2x * subdiv_step3;
		let dfy = (cy1 - y1) * pre1 + tmp1y * pre2 + tmp2y * subdiv_step3;

		let ddfx = tmp1x * pre4 + tmp2x * pre5;
		let ddfy = tmp1y * pre4 + tmp2y * pre5;

		let dddfx = tmp2x * pre5;
		let dddfy = tmp2y * pre5;

		while (segments-- > 0) {
			this.vertex(fx, fy, z, color);
			fx += dfx;
			fy += dfy;
			dfx += ddfx;
			dfy += ddfy;
			ddfx += dddfx;
			ddfy += dddfy;
			this.vertex(fx, fy, z, color);
		}
		this.vertex(fx, fy, z, color);
		this.vertex(x2, y2, z, color);
	}

	private vertex(x: number, y: number, z: number, color: Color) {
		let idx = this.vertexIndex;
		let vertices = this.mesh.getVertices();
		vertices[idx++] = x;
		vertices[idx++] = y;
		vertices[idx++] = z;
		vertices[idx++] = color.r;
		vertices[idx++] = color.g;
		vertices[idx++] = color.b;
		vertices[idx++] = color.a;
		this.vertexIndex = idx;
	}

	private flush() {
		if (this.vertexIndex == 0) return;
		this.mesh.setVerticesLength(this.vertexIndex);
		this.mesh.draw(this.shader, this.shapeType);
		this.vertexIndex = 0;
	}

	private check(shapeType: ShapeType, numVertices: number) {
		if (!this.isDrawing) throw new Error("ShapeRenderer.begin() has not been called");
		if (this.shapeType == shapeType) {
			if (this.mesh.maxVertices() - this.mesh.numVertices() < numVertices) this.flush();
			else return;
		} else {
			this.flush();
			this.shapeType = shapeType;
		}
	}

	dispose() {
		this.mesh.dispose();
		this.shader.dispose();
	}
}

export enum ShapeType {
	Point = 0x0000,
	Line = 0x0001,
	Filled = 0x0004
}
