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

import { Pool } from "../core/Utils";

export class Input {
	element: HTMLElement;
	lastX = 0;
	lastY = 0;
	buttonDown = false;
	currTouch: Touch = null;
	touchesPool = new Pool<Touch>(() => {
		return new Touch(0, 0, 0);
	});

	private listeners = new Array<InputListener>();
	constructor(element: HTMLElement) {
		this.element = element;
		this.setupCallbacks(element);
	}

	private setupCallbacks(element: HTMLElement) {
		let mouseDown = (ev: UIEvent) => {
			if (ev instanceof MouseEvent) {
				let rect = element.getBoundingClientRect();
				let x = ev.clientX - rect.left;
				let y = ev.clientY - rect.top;

				let listeners = this.listeners;
				for (let i = 0; i < listeners.length; i++) {
					if (listeners[i].down) listeners[i].down(x, y);
				}

				this.lastX = x;
				this.lastY = y;
				this.buttonDown = true;

				document.addEventListener("mousemove", mouseMove);
				document.addEventListener("mouseup", mouseUp);
			}
		}

		let mouseMove = (ev: UIEvent) => {
			if (ev instanceof MouseEvent) {
				let rect = element.getBoundingClientRect();
				let x = ev.clientX - rect.left;
				let y = ev.clientY - rect.top;

				let listeners = this.listeners;
				for (let i = 0; i < listeners.length; i++) {
					if (this.buttonDown) {
						if (listeners[i].dragged) listeners[i].dragged(x, y);
					} else {
						if (listeners[i].moved) listeners[i].moved(x, y);
					}
				}

				this.lastX = x;
				this.lastY = y;
			}
		};

		let mouseUp = (ev: UIEvent) => {
			if (ev instanceof MouseEvent) {
				let rect = element.getBoundingClientRect();
				let x = ev.clientX - rect.left;
				let y = ev.clientY - rect.top;

				let listeners = this.listeners;
				for (let i = 0; i < listeners.length; i++) {
					if (listeners[i].up) listeners[i].up(x, y);
				}

				this.lastX = x;
				this.lastY = y;
				this.buttonDown = false;
				document.removeEventListener("mousemove", mouseMove);
				document.removeEventListener("mouseup", mouseUp);
			}
		}



		element.addEventListener("mousedown", mouseDown, true);
		element.addEventListener("mousemove", mouseMove, true);
		element.addEventListener("mouseup", mouseUp, true);
		element.addEventListener("touchstart", (ev: TouchEvent) => {
			if (this.currTouch != null) return;

			var touches = ev.changedTouches;
			for (var i = 0; i < touches.length; i++) {
				var touch = touches[i];
				let rect = element.getBoundingClientRect();
				let x = touch.clientX - rect.left;
				let y = touch.clientY - rect.top;
				this.currTouch = this.touchesPool.obtain();
				this.currTouch.identifier = touch.identifier;
				this.currTouch.x = x;
				this.currTouch.y = y;
				break;
			}

			let listeners = this.listeners;
			for (let i = 0; i < listeners.length; i++) {
				if (listeners[i].down) listeners[i].down(this.currTouch.x, this.currTouch.y);
			}

			this.lastX = this.currTouch.x;
			this.lastY = this.currTouch.y;
			this.buttonDown = true;
			ev.preventDefault();
		}, false);
		element.addEventListener("touchend", (ev: TouchEvent) => {
			var touches = ev.changedTouches;
			for (var i = 0; i < touches.length; i++) {
				var touch = touches[i];
				if (this.currTouch.identifier === touch.identifier) {
					let rect = element.getBoundingClientRect();
					let x = this.currTouch.x = touch.clientX - rect.left;
					let y = this.currTouch.y = touch.clientY - rect.top;
					this.touchesPool.free(this.currTouch);
					let listeners = this.listeners;
					for (let i = 0; i < listeners.length; i++) {
						if (listeners[i].up) listeners[i].up(x, y);
					}

					this.lastX = x;
					this.lastY = y;
					this.buttonDown = false;
					this.currTouch = null;
					break;
				}
			}
			ev.preventDefault();
		}, false);
		element.addEventListener("touchcancel", (ev: TouchEvent) => {
			var touches = ev.changedTouches;
			for (var i = 0; i < touches.length; i++) {
				var touch = touches[i];
				if (this.currTouch.identifier === touch.identifier) {
					let rect = element.getBoundingClientRect();
					let x = this.currTouch.x = touch.clientX - rect.left;
					let y = this.currTouch.y = touch.clientY - rect.top;
					this.touchesPool.free(this.currTouch);
					let listeners = this.listeners;
					for (let i = 0; i < listeners.length; i++) {
						if (listeners[i].up) listeners[i].up(x, y);
					}

					this.lastX = x;
					this.lastY = y;
					this.buttonDown = false;
					this.currTouch = null;
					break;
				}
			}
			ev.preventDefault();
		}, false);
		element.addEventListener("touchmove", (ev: TouchEvent) => {
			if (this.currTouch == null) return;

			var touches = ev.changedTouches;
			for (var i = 0; i < touches.length; i++) {
				var touch = touches[i];
				if (this.currTouch.identifier === touch.identifier) {
					let rect = element.getBoundingClientRect();
					let x = touch.clientX - rect.left;
					let y = touch.clientY - rect.top;

					let listeners = this.listeners;
					for (let i = 0; i < listeners.length; i++) {
						if (listeners[i].dragged) listeners[i].dragged(x, y);
					}

					this.lastX = this.currTouch.x = x;
					this.lastY = this.currTouch.y = y;
					break;
				}
			}
			ev.preventDefault();
		}, false);
	}

	addListener(listener: InputListener) {
		this.listeners.push(listener);
	}

	removeListener(listener: InputListener) {
		let idx = this.listeners.indexOf(listener);
		if (idx > -1) {
			this.listeners.splice(idx, 1);
		}
	}
}

export class Touch {
	constructor(public identifier: number, public x: number, public y: number) {
	}
}

export interface InputListener {
	down(x: number, y: number): void;
	up(x: number, y: number): void;
	moved(x: number, y: number): void;
	dragged(x: number, y: number): void;
}
