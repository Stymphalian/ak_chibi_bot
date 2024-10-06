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

module spine.webgl {

	export interface Camera {
		position: Vector3;
		direction: Vector3;
		up: Vector3;
		near: number
		far: number
		fov: number
		zoom: number
		viewportWidth: number
		viewportHeight: number
		projectionView: Matrix4;
		inverseProjectionView: Matrix4;
		projection: Matrix4;
		view: Matrix4;
		is_perspective: boolean;

		update() : void;
		screenToWorld (screenCoords: Vector3, screenWidth: number, screenHeight: number) : Vector3;
		worldToScreen(worldCoords: spine.webgl.Vector3): spine.webgl.Vector3;
		setViewport(viewportWidth: number, viewportHeight: number): void;
	}

	export class OrthoCamera {
		position = new Vector3(0, 0, 0);
		direction = new Vector3(0, 0, -1);
		up = new Vector3(0, 1, 0);
		near = 0;
		far = 100;
		fov = 0.0; // unused
		zoom = 1;
		viewportWidth = 0;
		viewportHeight = 0;
		projectionView = new Matrix4();
		inverseProjectionView = new Matrix4();
		projection = new Matrix4();
		view = new Matrix4();
		is_perspective = false;

		private tmp = new Vector3();

		constructor (viewportWidth: number, viewportHeight: number) {
			this.viewportWidth = viewportWidth;
			this.viewportHeight = viewportHeight;
			this.update();
		}

		update () {
			let projection = this.projection;
			let view = this.view;
			let projectionView = this.projectionView;
			let inverseProjectionView = this.inverseProjectionView;
			let zoom = this.zoom, viewportWidth = this.viewportWidth, viewportHeight = this.viewportHeight;
			projection.ortho(zoom * (-viewportWidth / 2), zoom * (viewportWidth / 2),
							 zoom * (-viewportHeight / 2), zoom * (viewportHeight / 2),
							 this.near, this.far);
			view.lookAt(this.position, this.direction, this.up);
			projectionView.set(projection.values);
			projectionView.multiply(view);
			inverseProjectionView.set(projectionView.values).invert();
		}

		screenToWorld (screenCoords: Vector3, screenWidth: number, screenHeight: number) {
			let x = screenCoords.x;
			let y = screenCoords.y;
			let z = screenCoords.z;
			let tmp = this.tmp;
			tmp.x = (2 * (x / screenWidth)) - 1;
			// tmp.y = 1 - (2 * (y / screenHeight));
			tmp.y = (2 * (y / screenHeight)) - 1;
			tmp.z = z;
			tmp.project(this.inverseProjectionView);
			screenCoords.set(tmp.x, tmp.y, tmp.z);
			return screenCoords;
		}

		worldToScreen(worldCoords: spine.webgl.Vector3): spine.webgl.Vector3 {
			let tmp = new spine.webgl.Vector3(
				worldCoords.x,
				worldCoords.y,
				worldCoords.z,
			)
			tmp.multiply(this.projectionView);
			tmp.x = ((tmp.x + 1)/2) * (this.viewportWidth);
			// tmp.y = (1.0 - (tmp.y + 1)/2) * (this.viewportHeight);
			tmp.y = ((tmp.y + 1)/2) * (this.viewportHeight);
			return new spine.webgl.Vector3(tmp.x, tmp.y, 0);
		}

		// screenToWorld (screenCoords: Vector3, screenWidth: number, screenHeight: number) {
		// 	let x = screenCoords.x;y = screenHeight - screenCoords.y - 1;
		// 	let tmp = this.tmp;
		// 	tmp.x = (2 * x) / screenWidth - 1;
		// 	tmp.y = (2 * y) / screenHeight - 1;
		// 	tmp.z = (2 * screenCoords.z) - 1;
		// 	tmp.project(this.inverseProjectionView);
		// 	screenCoords.set(tmp.x, tmp.y, tmp.z);
		// 	return screenCoords;
		// }

		// worldToScreen(worldCoords: Vector2): spine.Vector2 {
		// 	let tmp = new spine.webgl.Vector3(
		// 		worldCoords.x,
		// 		worldCoords.y,
		// 		0,
		// 	)
		// 	tmp.multiply(this.projectionView);
		// 	tmp.x = ((tmp.x + 1)/2) * (this.viewportWidth);
		// 	tmp.y = ((tmp.y + 1)/2) * (this.viewportHeight);
		// 	return new spine.Vector2(tmp.x, tmp.y);
		// }

		setViewport(viewportWidth: number, viewportHeight: number) {
			this.viewportWidth = viewportWidth;
			this.viewportHeight = viewportHeight;
		}
	}

	export class PerspectiveCamera {
		position = new Vector3(0, 0, 0);
		direction = new Vector3(0, 0, -1);
		up = new Vector3(0, 1, 0);
		near = 1;
		far = 1000;
		zoom = 1;
		fov = 45;
		viewportWidth = 0;
		viewportHeight = 0;
		projectionView = new Matrix4();
		inverseProjectionView = new Matrix4();
		projection = new Matrix4();
		view = new Matrix4();
		is_perspective = true;

		private tmp = new Vector3();

		constructor (viewportWidth: number, viewportHeight: number) {
			this.viewportWidth = viewportWidth;
			this.viewportHeight = viewportHeight;
			this.update();
		}

		update (if_throw = false) {
			let projection = this.projection;
			let view = this.view;
			let projectionView = this.projectionView;
			let inverseProjectionView = this.inverseProjectionView;

			try {
				projection.projection(
					this.near,
					this.far,
					this.fov,
					this.viewportWidth / this.viewportHeight
				)
				view.lookAt(this.position, this.direction, this.up);
				projectionView.set(projection.values);
				projectionView.multiply(view);
				inverseProjectionView.set(projectionView.values).invert();
			} catch(e) {
				if (if_throw) {
					throw e;
				}
			}
		}

		screenToWorld (screenCoords: Vector3, screenWidth: number, screenHeight: number) {
			let x = screenCoords.x;
			let y = screenCoords.y;
			let z = screenCoords.z;
			let tmp = this.tmp;
			tmp.x = (2 * (x / screenWidth)) - 1;
			tmp.y = (2 * (y / screenHeight)) - 1;
			// tmp.y = 1 - (2 * (y / screenHeight));  old
			tmp.z = z;
			tmp.project(this.inverseProjectionView);
			screenCoords.set(tmp.x, tmp.y, tmp.z);
			return screenCoords;
		}

		worldToScreen(worldCoords: spine.webgl.Vector3): spine.webgl.Vector3 {
			let tmp = new spine.webgl.Vector3(
				worldCoords.x,
				worldCoords.y,
				worldCoords.z,
			)

			tmp.project(this.projectionView);
			tmp.x = ((tmp.x + 1)/2) * (this.viewportWidth);
			tmp.y = ((tmp.y + 1)/2) * (this.viewportHeight);
			// tmp.y = (1.0 - ((tmp.y + 1)/2)) * (this.viewportHeight); old
			let ret =  new spine.webgl.Vector3(tmp.x, tmp.y, tmp.z);
			ret.w = tmp.w;
			return ret;
		}

		setViewport(viewportWidth: number, viewportHeight: number) {
			this.viewportWidth = viewportWidth;
			this.viewportHeight = viewportHeight;
		}
	}
}
