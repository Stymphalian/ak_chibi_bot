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

import { ClippingAttachment } from "../core/attachments/ClippingAttachment";
import { MeshAttachment } from "../core/attachments/MeshAttachment";
import { RegionAttachment } from "../core/attachments/RegionAttachment";
import { BlendMode } from "../core/BlendMode";
import { Skeleton } from "../core/Skeleton";
import { SkeletonClipping } from "../core/SkeletonClipping";
import { TextureAtlasRegion } from "../core/TextureAtlas";
import { Color, Vector2, Utils, ArrayLike } from "../core/Utils";
import { VertexEffect } from "../core/VertexEffect";
import { GLTexture } from "./GLTexture";
import { PolygonBatcher } from "./PolygonBatcher";
import { ManagedWebGLRenderingContext, WebGLBlendModeConverter } from "./WebGL";

class Renderable {
	constructor(public vertices: ArrayLike<number>, public numVertices: number, public numFloats: number) { }
};

export class SkeletonRenderer {
	static QUAD_TRIANGLES = [0, 1, 2, 2, 3, 0];

	premultipliedAlpha = false;
	vertexEffect: VertexEffect = null;
	private tempColor = new Color();
	private tempColor2 = new Color();
	private vertices: ArrayLike<number>;
	private vertexSize = 2 + 2 + 4;
	private twoColorTint = false;
	private renderable: Renderable = new Renderable(null, 0, 0);
	private clipper: SkeletonClipping = new SkeletonClipping();
	private temp = new Vector2();
	private temp2 = new Vector2();
	private temp3 = new Color();
	private temp4 = new Color();

	constructor(context: ManagedWebGLRenderingContext, twoColorTint: boolean = true) {
		this.twoColorTint = twoColorTint;
		if (twoColorTint)
			this.vertexSize += 4;
		this.vertices = Utils.newFloatArray(this.vertexSize * 1024);
	}

/**
  * Draws the skeleton using the provided PolygonBatcher and Skeleton instances.
  * This method handles rendering of different attachment types, applying vertex effects,
  * clipping, and blending.
  *
  * @param batcher The PolygonBatcher instance to use for drawing.
  * @param skeleton The Skeleton instance to draw.
  * @param slotRangeStart (Optional) The index of the first slot to draw. If not provided, all slots will be drawn.
  * @param slotRangeEnd (Optional) The index of the last slot to draw. If not provided, all slots will be drawn.
  */
 draw(batcher: PolygonBatcher, skeleton: Skeleton, slotRangeStart: number = -1, slotRangeEnd: number = -1) {
		let clipper = this.clipper;
		let premultipliedAlpha = this.premultipliedAlpha;
		let twoColorTint = this.twoColorTint;
		let blendMode: BlendMode = null;

		let tempPos = this.temp;
		let tempUv = this.temp2;
		let tempLight = this.temp3;
		let tempDark = this.temp4;

		let renderable: Renderable = this.renderable;
		let uvs: ArrayLike<number> = null;
		let triangles: Array<number> = null;
		let drawOrder = skeleton.drawOrder;
		let attachmentColor: Color = null;
		let skeletonColor = skeleton.color;
		let vertexSize = twoColorTint ? 12 : 8;
		let inRange = false;
		if (slotRangeStart == -1) inRange = true;
		for (let i = 0, n = drawOrder.length; i < n; i++) {
			let clippedVertexSize = clipper.isClipping() ? 2 : vertexSize;
			let slot = drawOrder[i];
			if (!slot.bone.active) {
				clipper.clipEndWithSlot(slot);
				continue;
			}

			if (slotRangeStart >= 0 && slotRangeStart == slot.data.index) {
				inRange = true;
			}

			if (!inRange) {
				clipper.clipEndWithSlot(slot);
				continue;
			}

			if (slotRangeEnd >= 0 && slotRangeEnd == slot.data.index) {
				inRange = false;
			}

			let attachment = slot.getAttachment();
			let texture: GLTexture = null;
			if (attachment instanceof RegionAttachment) {
				let region = <RegionAttachment>attachment;
				renderable.vertices = this.vertices;
				renderable.numVertices = 4;
				renderable.numFloats = clippedVertexSize << 2; // x4 (4 bytes per float)
				region.computeWorldVertices(slot.bone, renderable.vertices, 0, clippedVertexSize);
				triangles = SkeletonRenderer.QUAD_TRIANGLES;
				uvs = region.uvs;
				texture = <GLTexture>(<TextureAtlasRegion>region.region.renderObject).texture;
				attachmentColor = region.color;
			} else if (attachment instanceof MeshAttachment) {
				let mesh = <MeshAttachment>attachment;
				renderable.vertices = this.vertices;
				renderable.numVertices = (mesh.worldVerticesLength >> 1);
				renderable.numFloats = renderable.numVertices * clippedVertexSize;
				if (renderable.numFloats > renderable.vertices.length) {
					renderable.vertices = this.vertices = Utils.newFloatArray(renderable.numFloats);
				}
				mesh.computeWorldVertices(slot, 0, mesh.worldVerticesLength, renderable.vertices, 0, clippedVertexSize);
				triangles = mesh.triangles;
				texture = <GLTexture>(<TextureAtlasRegion>mesh.region.renderObject).texture;
				uvs = mesh.uvs;
				attachmentColor = mesh.color;
			} else if (attachment instanceof ClippingAttachment) {
				let clip = <ClippingAttachment>(attachment);
				clipper.clipStart(slot, clip);
				continue;
			} else {
				clipper.clipEndWithSlot(slot);
				continue;
			}

			if (texture != null) {
				let slotColor = slot.color;
				let finalColor = this.tempColor;
				finalColor.r = skeletonColor.r * slotColor.r * attachmentColor.r;
				finalColor.g = skeletonColor.g * slotColor.g * attachmentColor.g;
				finalColor.b = skeletonColor.b * slotColor.b * attachmentColor.b;
				finalColor.a = skeletonColor.a * slotColor.a * attachmentColor.a;
				if (premultipliedAlpha) {
					finalColor.r *= finalColor.a;
					finalColor.g *= finalColor.a;
					finalColor.b *= finalColor.a;
				}
				let darkColor = this.tempColor2;
				if (slot.darkColor == null)
					darkColor.set(0, 0, 0, 1.0);
				else {
					if (premultipliedAlpha) {
						darkColor.r = slot.darkColor.r * finalColor.a;
						darkColor.g = slot.darkColor.g * finalColor.a;
						darkColor.b = slot.darkColor.b * finalColor.a;
					} else {
						darkColor.setFromColor(slot.darkColor);
					}
					darkColor.a = premultipliedAlpha ? 1.0 : 0.0;
				}

				let slotBlendMode = slot.data.blendMode;
				if (slotBlendMode != blendMode) {
					blendMode = slotBlendMode;
					batcher.setBlendMode(
						WebGLBlendModeConverter.getSourceGLBlendMode(blendMode, premultipliedAlpha),
						WebGLBlendModeConverter.getDestGLBlendMode(blendMode));
				}

				if (clipper.isClipping()) {
					clipper.clipTriangles(renderable.vertices, renderable.numFloats, triangles, triangles.length, uvs, finalColor, darkColor, twoColorTint);
					let clippedVertices = new Float32Array(clipper.clippedVertices);
					let clippedTriangles = clipper.clippedTriangles;
					if (this.vertexEffect != null) {
						let vertexEffect = this.vertexEffect;
						let verts = clippedVertices;
						if (!twoColorTint) {
							for (let v = 0, n = clippedVertices.length; v < n; v += vertexSize) {
								tempPos.x = verts[v];
								tempPos.y = verts[v + 1];
								tempLight.set(verts[v + 2], verts[v + 3], verts[v + 4], verts[v + 5]);
								tempUv.x = verts[v + 6];
								tempUv.y = verts[v + 7];
								tempDark.set(0, 0, 0, 0);
								vertexEffect.transform(tempPos, tempUv, tempLight, tempDark);
								verts[v] = tempPos.x;
								verts[v + 1] = tempPos.y;
								verts[v + 2] = tempLight.r;
								verts[v + 3] = tempLight.g;
								verts[v + 4] = tempLight.b;
								verts[v + 5] = tempLight.a;
								verts[v + 6] = tempUv.x;
								verts[v + 7] = tempUv.y
							}
						} else {
							for (let v = 0, n = clippedVertices.length; v < n; v += vertexSize) {
								tempPos.x = verts[v];
								tempPos.y = verts[v + 1];
								tempLight.set(verts[v + 2], verts[v + 3], verts[v + 4], verts[v + 5]);
								tempUv.x = verts[v + 6];
								tempUv.y = verts[v + 7];
								tempDark.set(verts[v + 8], verts[v + 9], verts[v + 10], verts[v + 11]);
								vertexEffect.transform(tempPos, tempUv, tempLight, tempDark);
								verts[v] = tempPos.x;
								verts[v + 1] = tempPos.y;
								verts[v + 2] = tempLight.r;
								verts[v + 3] = tempLight.g;
								verts[v + 4] = tempLight.b;
								verts[v + 5] = tempLight.a;
								verts[v + 6] = tempUv.x;
								verts[v + 7] = tempUv.y
								verts[v + 8] = tempDark.r;
								verts[v + 9] = tempDark.g;
								verts[v + 10] = tempDark.b;
								verts[v + 11] = tempDark.a;
							}
						}
					}
					batcher.draw(texture, clippedVertices, clippedTriangles, skeleton.z);
				} else {
					let verts = renderable.vertices;
					if (this.vertexEffect != null) {
						let vertexEffect = this.vertexEffect;
						if (!twoColorTint) {
							for (let v = 0, u = 0, n = renderable.numFloats; v < n; v += vertexSize, u += 2) {
								tempPos.x = verts[v];
								tempPos.y = verts[v + 1];
								tempUv.x = uvs[u];
								tempUv.y = uvs[u + 1]
								tempLight.setFromColor(finalColor);
								tempDark.set(0, 0, 0, 0);
								vertexEffect.transform(tempPos, tempUv, tempLight, tempDark);
								verts[v] = tempPos.x;
								verts[v + 1] = tempPos.y;
								verts[v + 2] = tempLight.r;
								verts[v + 3] = tempLight.g;
								verts[v + 4] = tempLight.b;
								verts[v + 5] = tempLight.a;
								verts[v + 6] = tempUv.x;
								verts[v + 7] = tempUv.y
							}
						} else {
							for (let v = 0, u = 0, n = renderable.numFloats; v < n; v += vertexSize, u += 2) {
								tempPos.x = verts[v];
								tempPos.y = verts[v + 1];
								tempUv.x = uvs[u];
								tempUv.y = uvs[u + 1]
								tempLight.setFromColor(finalColor);
								tempDark.setFromColor(darkColor);
								vertexEffect.transform(tempPos, tempUv, tempLight, tempDark);
								verts[v] = tempPos.x;
								verts[v + 1] = tempPos.y;
								verts[v + 2] = tempLight.r;
								verts[v + 3] = tempLight.g;
								verts[v + 4] = tempLight.b;
								verts[v + 5] = tempLight.a;
								verts[v + 6] = tempUv.x;
								verts[v + 7] = tempUv.y
								verts[v + 8] = tempDark.r;
								verts[v + 9] = tempDark.g;
								verts[v + 10] = tempDark.b;
								verts[v + 11] = tempDark.a;
							}
						}
					} else {
						if (!twoColorTint) {
							for (let v = 2, u = 0, n = renderable.numFloats; v < n; v += vertexSize, u += 2) {
								verts[v] = finalColor.r;
								verts[v + 1] = finalColor.g;
								verts[v + 2] = finalColor.b;
								verts[v + 3] = finalColor.a;
								verts[v + 4] = uvs[u];
								verts[v + 5] = uvs[u + 1];
							}
						} else {
							for (let v = 2, u = 0, n = renderable.numFloats; v < n; v += vertexSize, u += 2) {
								verts[v] = finalColor.r;
								verts[v + 1] = finalColor.g;
								verts[v + 2] = finalColor.b;
								verts[v + 3] = finalColor.a;
								verts[v + 4] = uvs[u];
								verts[v + 5] = uvs[u + 1];
								verts[v + 6] = darkColor.r;
								verts[v + 7] = darkColor.g;
								verts[v + 8] = darkColor.b;
								verts[v + 9] = darkColor.a;
							}
						}
					}
					let view = (renderable.vertices as Float32Array).subarray(0, renderable.numFloats);
					batcher.draw(texture, view, triangles, skeleton.z);
				}
			}

			clipper.clipEndWithSlot(slot);
		}
		clipper.clipEnd();
	}
}
