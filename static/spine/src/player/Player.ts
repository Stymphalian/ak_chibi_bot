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

import { AssetManager } from "../webgl/AssetManager";
import { AtlasAttachmentLoader } from "../core/AtlasAttachmentLoader"
import { Skeleton } from "../core/Skeleton"
import { SkeletonBinary } from "../core/SkeletonBinary"
import { SkeletonData } from "../core/SkeletonData"
import { SkeletonJson } from "../core/SkeletonJson"
import { Vector2, Color } from "../core/Utils"
import { Input } from "../webgl/Input"
import { LoadingScreen } from "../webgl/LoadingScreen"
import { ResizeMode, SceneRenderer } from "../webgl/SceneRenderer"
import { Vector3 } from "../webgl/Vector3"
import { ManagedWebGLRenderingContext } from "../webgl/WebGL"
import { Actor, SpineActorConfig } from "./Actor"
import { createElement, findWithClass, escapeHtml, isAlphanumeric, findWithId, configurePerspectiveCamera, updateCameraSettings, BoundingBox, Viewport, isValidTwitchUserDisplayName } from "./Utils";
import { Camera } from "../webgl/Camera";
import { readSpritesheetJsonConfig, SpritesheetActor } from "./Spritesheet";
import { GLFrameBuffer } from "../webgl/GLFrameBuffer";
import { OffscreenRender } from "./OffscreenRender";

export interface SpinePlayerConfig {
	/* Optional: whether to show the player controls. Default: true. */
	showControls: boolean

	/* Optional: the position and size of the viewport in world coordinates of the skeleton. Default: the setup pose bounding box. */
	viewport: Viewport

	/* Optional: whether the canvas should be transparent. Default: false. */
	alpha: boolean

	/* Optional: the background color. Must be given in the format #rrggbbaa. Default: #000000ff. */
	backgroundColor: string

	textSize: number,
	textFont: string,

	/* Optional: the background image. Default: none. */
	backgroundImage: {
		/* The URL of the background image */
		url: string

		/* Optional: the position and size of the background image in world coordinates. Default: viewport. */
		x: number
		y: number
		width: number
		height: number
	} | undefined

	/** Optional: 
	 *  How often to send back runtime debug information to the server.
	 *  This is to help debug performance issues. Default: 60 seconds.
	 */
	runtimeDebugInfoDumpIntervalSec: number

	// Camera Near and Far customization
	cameraPerspectiveNear: number
	cameraPerspectiveFar: number

	// The extra scaling to be applied to the all the chibis
	chibiScale: number
	showChatMessages: boolean

	// Configs to help with crowded chibis on the screen
	// ----------
	excessiveChibiMitigations: boolean
}

class Slider {
	private slider: HTMLElement;
	private value: HTMLElement;
	private knob: HTMLElement;
	public change: (percentage: number) => void;

	constructor(public snaps = 0, public snapPercentage = 0.1, public big = false) { }

	render(): HTMLElement {
		this.slider = createElement(/*html*/`
				<div class="spine-player-slider ${this.big ? "big" : ""}">
					<div class="spine-player-slider-value"></div>
					<!--<div class="spine-player-slider-knob"></div>-->
				</div>
			`);
		this.value = findWithClass(this.slider, "spine-player-slider-value")[0];
		// this.knob = findWithClass(this.slider, "spine-player-slider-knob")[0];
		this.setValue(0);

		let input = new Input(this.slider);
		var dragging = false;
		input.addListener({
			down: (x, y) => {
				dragging = true;
				this.value.classList.add("hovering");
			},
			up: (x, y) => {
				dragging = false;
				let percentage = x / this.slider.clientWidth;
				percentage = percentage = Math.max(0, Math.min(percentage, 1));
				this.setValue(x / this.slider.clientWidth);
				if (this.change) this.change(percentage);
				this.value.classList.remove("hovering");
			},
			moved: (x, y) => {
				if (dragging) {
					let percentage = x / this.slider.clientWidth;
					percentage = Math.max(0, Math.min(percentage, 1));
					percentage = this.setValue(x / this.slider.clientWidth);
					if (this.change) this.change(percentage);
				}
			},
			dragged: (x, y) => {
				let percentage = x / this.slider.clientWidth;
				percentage = Math.max(0, Math.min(percentage, 1));
				percentage = this.setValue(x / this.slider.clientWidth);
				if (this.change) this.change(percentage);
			}
		});


		return this.slider;
	}

	setValue(percentage: number): number {
		percentage = Math.max(0, Math.min(1, percentage));
		if (this.snaps > 0) {
			let modulo = percentage % (1 / this.snaps);
			// floor
			if (modulo < (1 / this.snaps) * this.snapPercentage) {
				percentage = percentage - modulo;
			} else if (modulo > (1 / this.snaps) - (1 / this.snaps) * this.snapPercentage) {
				percentage = percentage - modulo + (1 / this.snaps);
			}
			percentage = Math.max(0, Math.min(1, percentage));
		}
		this.value.style.width = "" + (percentage * 100) + "%";
		// this.knob.style.left = "" + (-8 + percentage * this.slider.clientWidth) + "px";
		return percentage;
	}
}

type RenderCallbackFn = (context: ManagedWebGLRenderingContext, textContext: CanvasRenderingContext2D) => void;
type RemoveCallbackFn = () => void;

export class SpinePlayer {
	static HOVER_COLOR_INNER = new Color(0.478, 0, 0, 0.25);
	static HOVER_COLOR_OUTER = new Color(1, 1, 1, 1);
	static NON_HOVER_COLOR_INNER = new Color(0.478, 0, 0, 0.5);
	static NON_HOVER_COLOR_OUTER = new Color(1, 0, 0, 0.8);

	// private average_width = 0;
	// private average_height = 0;
	// private average_count = 0;

	private sceneRenderer: SceneRenderer;
	private dom: HTMLElement;
	private playerControls: HTMLElement;
	private canvas: HTMLCanvasElement;
	private textCanvas: HTMLCanvasElement;
	private textCanvasContext: CanvasRenderingContext2D;
	private timelineSlider: Slider;
	private playButton: HTMLElement;

	private context: ManagedWebGLRenderingContext;
	private loadingScreen: LoadingScreen;
	private assetManager: AssetManager;
	private paused = false;
	private actors = new Map<string, Actor>();
	private playerConfig: SpinePlayerConfig = null;

	private parent: HTMLElement;

	private stopRequestAnimationFrame = false;
	private lastRequestAnimationFrameId = 0;
	private windowFpsFrameCount: number = 0;
	private webSocket: WebSocket = null;
	private renderCallbacks: RenderCallbackFn[] = [];
	private actorQueue: { actorName: string, config: SpineActorConfig }[] = [];
	private actorQueueIndex: any = null;

	// private offCanvas: HTMLCanvasElement|null;
	private offscreenRender: OffscreenRender = null;
	private actorHeightDirty = true;

	constructor(parent: HTMLElement | string, playerConfig: SpinePlayerConfig) {
		if (typeof parent === "string") {
			this.parent = document.getElementById(parent);
		} else {
			this.parent = parent;
		}
		// this.playerConfig = null;
		this.playerConfig = this.validatePlayerConfig(playerConfig);
		this.parent.appendChild(this.setupDom());
	}

	setWebsocket(webSocket: WebSocket) {
		this.webSocket = webSocket;
		this.setShowChatMessagesInRoom(this.playerConfig.showChatMessages);
	}

	validatePlayerConfig(config: SpinePlayerConfig): SpinePlayerConfig {
		if (!config) throw new Error("Please pass a configuration to new.SpinePlayer().");
		if (!config.alpha) config.alpha = false;
		if (!config.backgroundColor) config.backgroundColor = "#000000";
		if (typeof config.showControls === "undefined")
			config.showControls = true;
		if (!config.runtimeDebugInfoDumpIntervalSec) config.runtimeDebugInfoDumpIntervalSec = 60;
		if (!config.textSize) config.textSize = 14;
		if (!config.textFont) config.textFont = "lato";
		if (!config.cameraPerspectiveNear) config.cameraPerspectiveNear = 1.0;
		if (!config.cameraPerspectiveFar) config.cameraPerspectiveFar = 1000.0;
		if (!config.showChatMessages) config.showChatMessages = false;
		return config;
	}

	validateActorConfig(config: SpineActorConfig): SpineActorConfig {
		if (!config) throw new Error("Please pass a configuration to new.SpinePlayer().");

		if (config.spritesheetDataUrl) {
			if (!config.atlasUrl) throw new Error("Please specify the URL of the atlas file.");
		} else {
			if (!config.jsonUrl && !config.skelUrl) throw new Error("Please specify the URL of the skeleton JSON or .skel file.");
			if (!config.atlasUrl) throw new Error("Please specify the URL of the atlas file.");
		}

		// if (!config.alpha) config.alpha = false;
		// if (!config.backgroundColor) config.backgroundColor = "#000000";
		// if (!config.fullScreenBackgroundColor) config.fullScreenBackgroundColor = config.backgroundColor;
		if (typeof config.premultipliedAlpha === "undefined") config.premultipliedAlpha = true;
		if (!config.scaleX) config.scaleX = 1;
		if (!config.scaleY) config.scaleY = 1;
		if (!config.extraOffsetX) config.extraOffsetX = 0;
		if (!config.extraOffsetY) config.extraOffsetY = 0;
		if (!config.success) config.success = (widget, actor) => { };
		if (!config.error) config.error = (widget, actor, msg) => { };
		if (!config.animation_listener) {
			config.animation_listener = {
				event: function (trackIndex, event) {
					// console.log("Event on track " + trackIndex + ": " + JSON.stringify(event));
				},
				complete: function (trackIndex) {
					// console.log("Animation on track " + trackIndex + " completed");
				},
				start: function (trackIndex) {
					// console.log("Animation on track " + trackIndex + " started");
				},
				end: function (trackIndex) {
					// console.log("Animation on track " + trackIndex + " ended");
				},
				interrupt: function (trackIndex) {
					// console.log("Animation on track " + trackIndex + " ended");
				},
				dispose: function (trackIndex) {
					// console.log("Animation on track " + trackIndex + " ended");
				},
			}
		}

		if (config.skins && config.skin) {
			if (config.skins.indexOf(config.skin) < 0) {
				throw new Error("Default skin '" +
					config.skin +
					"' is not contained in the list of selectable skins " +
					escapeHtml(JSON.stringify(config.skins)) +
					"."
				);
			}
		}
		if (typeof config.defaultMix === "undefined")
			config.defaultMix = 0.25;

		if (config.configScaleX == undefined) { config.configScaleX = 1.0; }
		if (config.configScaleY == undefined) { config.configScaleY = 1.0; }
		if (config.scaleX == undefined) { config.scaleX = 0.45; }
		if (config.scaleY == undefined) { config.scaleY = 0.45; }
		if (config.maxSizePx == undefined) { config.maxSizePx = 350; }
		if (config.startPosX == undefined) { config.startPosX = null; }
		if (config.startPosY == undefined) { config.startPosY = null; }
		if (config.extraOffsetX == undefined) { config.extraOffsetX = 0; }
		if (config.extraOffsetY == undefined) { config.extraOffsetY = 0; }
		if (config.defaultMovementSpeedPxX == undefined) { config.defaultMovementSpeedPxX = 80; }
		if (config.defaultMovementSpeedPxY == undefined) { config.defaultMovementSpeedPxY = 80; }
		if (config.defaultMovementSpeedPxZ == undefined) { config.defaultMovementSpeedPxZ = 80; }
		if (config.movementSpeedPxX == undefined) { config.movementSpeedPxX = null; }
		if (config.movementSpeedPxY == undefined) { config.movementSpeedPxY = null; }
		if (config.movementSpeedPxZ == undefined) { config.movementSpeedPxZ = config.defaultMovementSpeedPxZ; }

		if (config.chibiId == undefined) { config.chibiId = crypto.randomUUID(); }
		if (config.userDisplayName == undefined) { config.userDisplayName = crypto.randomUUID(); }
		if (!isAlphanumeric(config.chibiId)) {
			throw Error("ChibiId is not valid ");
		}
		if (!isValidTwitchUserDisplayName(config.userDisplayName)) {
			throw Error("userDisplayName is not valid ");
		}
		if (config.action == undefined) {
			throw Error("config.action must be set");
		}

		return config;
	}

	showError(actor: Actor, error: string) {
		let errorDom = findWithClass(this.dom, "spine-player-error")[0];
		errorDom.classList.remove("spine-player-hidden");
		errorDom.innerHTML = `<p style="text-align: center; align-self: center;">${error}</p>`;

		// this.playerConfig.error(this, error);
		actor.config.error(this, actor, error);
	}

	hideError() {
		let errorDom = findWithClass(this.dom, "spine-player-error")[0];
		errorDom.classList.add("spine-player-hidden");
	}

	setupDom(): HTMLElement {
		let dom = this.dom = createElement(/*html*/`
				<div class="spine-player">
					<div class="spine-player-error spine-player-hidden"></div>
					<canvas id="spine-canvas" class="spine-player-canvas"></canvas>
					<canvas id="spine-text" class="spine-player-text-canvas"></canvas>
					<div class="spine-player-controls spine-player-popup-parent spine-player-controls-hidden">
						<div class="spine-player-timeline">
						</div>
						<div class="spine-player-buttons">
							<button id="spine-player-button-play-pause" class="spine-player-button spine-player-button-icon-pause"></button>
							<div class="spine-player-button-spacer"></div>
						</div>
					</div>
				</div>
			`)

		try {
			// Setup the scene renderer and OpenGL context
			// this.canvas = findWithClass(dom, "spine-player-canvas")[0] as HTMLCanvasElement;
			this.canvas = findWithId(dom, "spine-canvas")[0] as HTMLCanvasElement;
			this.textCanvas = findWithId(dom, "spine-text")[0] as HTMLCanvasElement;
			this.textCanvasContext = this.textCanvas.getContext("2d");
			this.textCanvasContext.font = this.playerConfig.textSize + "px " + this.playerConfig.textFont;
			this.textCanvasContext.textBaseline = "bottom";

			// var webglConfig = { alpha: config.alpha };
			var webglConfig = { alpha: this.playerConfig.alpha };
			this.context = new ManagedWebGLRenderingContext(this.canvas, webglConfig);
			// Setup the scene renderer and loading screen
			this.sceneRenderer = new SceneRenderer(this.canvas, this.context);
			this.loadingScreen = new LoadingScreen(this.sceneRenderer);
			this.offscreenRender = new OffscreenRender(this.sceneRenderer);
			this.assetManager = new AssetManager(this.context);

			if (this.playerConfig.viewport.debugRender) {
				this.context.GetWebGLParameters();
			}
		} catch (e) {
			// this.showError("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
			console.log("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
			return dom;
		}

		// configure the camera
		configurePerspectiveCamera(
			this.sceneRenderer.camera,
			this.playerConfig.cameraPerspectiveNear,
			this.playerConfig.cameraPerspectiveFar,
			this.playerConfig.viewport
		);

		// Setup rendering loop
		this.lastRequestAnimationFrameId = requestAnimationFrame(() => this.drawFrame());

		// Setup FPS counter
		performance.mark('fps_window_start');
		this.windowFpsFrameCount = 0;
		if (this.playerConfig.runtimeDebugInfoDumpIntervalSec > 0) {
			setInterval(() => this.callbackSendRuntimeUpdateInfo(), 1000 * this.playerConfig.runtimeDebugInfoDumpIntervalSec);
		}

		// Setup the event listeners for UI elements
		this.playerControls = findWithClass(dom, "spine-player-controls")[0];
		let timeline = findWithClass(dom, "spine-player-timeline")[0];
		this.timelineSlider = new Slider();
		timeline.appendChild(this.timelineSlider.render());
		this.playButton = findWithId(dom, "spine-player-button-play-pause")[0];

		this.playButton.onclick = () => {
			if (this.paused) this.play()
			else this.pause();
		}

		// Register a global resize handler to redraw and avoid flicker
		window.onresize = () => {
			this.drawFrame(false);
		}

		// Setup the input processor and controllable bones
		this.setupInput();

		return dom;
	}

	callbackSendRuntimeUpdateInfo() {
		performance.mark('fps_window_end');
		const totalTime = performance.measure('fps_window_time', 'fps_window_start', 'fps_window_end')
		const averageFps = this.windowFpsFrameCount / (totalTime.duration / 1000);

		if (this.webSocket != null) {
			const payload = JSON.stringify(
				{
					type_name: "RUNTIME_DEBUG_UPDATE",
					average_fps: averageFps,
				}
			)
			this.webSocket.send(payload);
		}

		performance.mark('fps_window_start');
		this.windowFpsFrameCount = 0;
	}

	setShowChatMessagesInRoom(showChatMessages: boolean) {
		if (this.webSocket != null) {
			const payload = JSON.stringify({
				type_name: "RUNTIME_ROOM_SETTINGS",
				show_chat_messages: showChatMessages
			})
			this.webSocket.send(payload);
		}
	}

	addActorToUpdateQueue(actorName: string, config: SpineActorConfig) {
		this.actorQueue.push({ actorName: actorName, config: config });
		if (this.actorQueueIndex == null) {
			this.actorQueueIndex = setTimeout(() => this.consumeActorUpdateQueue(), 100);
		}
	}

	private consumeActorUpdateQueue() {
		// console.log("actorUpdateQueue: ", this.actorQueue.length);
		if (this.actorQueue.length == 0) {
			this.actorQueueIndex = null;
			return;
		}
		if (!this.assetManager.isLoadingComplete()) {
			this.actorQueueIndex = setTimeout(() => this.consumeActorUpdateQueue(), 100);
			return;
		}

		let task = this.actorQueue.shift();
		this.changeOrAddActor(task.actorName, task.config);
		this.actorQueueIndex = setTimeout(() => this.consumeActorUpdateQueue(), 100);
	}

	changeOrAddActor(actorName: string, config: SpineActorConfig) {
		this.actorHeightDirty = true;
		if (this.actors.has(actorName)) {
			let actor = this.actors.get(actorName);
			actor.ResetWithConfig(config);
			this.setupActor(actor);
		} else {
			let actor = new Actor(
				config,
				this.playerConfig.viewport,
				this.offscreenRender,
				this.playerConfig.excessiveChibiMitigations);
			this.setupActor(actor);
			this.actors.set(actorName, actor);
		}
	}

	public getActor(actorName: string): Actor {
		return this.actors.get(actorName);
	}

	public getActorNames(): Array<string> {
		return Array.from(this.actors.keys());
	}

	removeActor(actorName: string) {
		this.actors.delete(actorName)
	}

	flashFindCharacter(actorName: string) {
		if (!this.actors.has(actorName)) {
			return;
		}
		this.actors.get(actorName).FlashFindCharacter();
	}

	showChatMessage(actorName: string, message: string) {
		if (!this.actors.has(actorName)) {
			return;
		}
		let actor = this.actors.get(actorName);
		actor.EnqueueChatMessage(message);
	}

	setupActor(actor: Actor) {
		let config = actor.config;

		try {
			// Validate the configuration
			// this.config = this.validatePlayerConfig(config);
			actor.config = this.validateActorConfig(actor.config);
		} catch (e) {
			this.showError(actor, e);
			return;
		}

		// Load the assets
		if (config.rawDataURIs) {
			for (let path in config.rawDataURIs) {
				let data = config.rawDataURIs[path];
				this.assetManager.setRawDataURI(path, data);
			}
		}

		if (actor.isSpritesheetActor()) {
			this.assetManager.loadText(config.spritesheetDataUrl);
			this.assetManager.loadTextureAtlas(config.atlasUrl);
		} else {
			if (config.jsonUrl) {
				this.assetManager.loadText(config.jsonUrl);
			} else {
				this.assetManager.loadBinary(config.skelUrl);
			}
			this.assetManager.loadTextureAtlas(config.atlasUrl);
			if (config.backgroundImage && config.backgroundImage.url) {
				this.assetManager.loadTexture(config.backgroundImage.url);
			}
		}

		// Setup rendering loop
		// this.lastRequestAnimationFrameId = requestAnimationFrame(() => this.drawFrame());
	}

	updateCameraSettings(cam: Camera, actor: Actor, viewport: BoundingBox) {
		updateCameraSettings(cam, actor, viewport)
	}

	drawText(texts: string[], xpx: number, ypx: number, zpx: number = 0) {
		let viewport = this.playerConfig.viewport;
		let tt = this.sceneRenderer.camera.worldToScreen(
			new Vector3(xpx, ypx, zpx)
		);
		let textpos = new Vector2(tt.x, tt.y);
		textpos.y = viewport.height - textpos.y;
		xpx = textpos.x;
		ypx = textpos.y;

		let ctx = this.textCanvasContext;
		ctx.save();
		ctx.translate(xpx, ypx);

		// Measure how much space is required for the speech bubble
		let width = 0;
		let height = 0;
		let heights = [];
		for (let i = 0; i < texts.length; i++) {
			let data = ctx.measureText(texts[i]);
			let h = data.actualBoundingBoxAscent - data.actualBoundingBoxDescent;
			let w = data.width;
			width = Math.max(width, w);
			height += h;
			heights.push(h);
		}

		// Draw a speech bubble box
		ctx.beginPath();
		let pad = 5
		ctx.fillStyle = "black";
		ctx.strokeStyle = "#333";
		ctx.lineWidth = 5;
		ctx.roundRect(-width / 2 - pad, pad, width + 2 * pad, -height - 2 * pad, 3);
		ctx.stroke();
		ctx.fill();

		// Draw a small triangle below the rect
		ctx.beginPath();
		ctx.fillStyle = "black"
		ctx.moveTo(-5, pad);
		ctx.lineTo(0, pad + 5);
		ctx.lineTo(5, pad);
		ctx.closePath();
		ctx.fill();
		// We want the border to blend in with the box and triangle
		ctx.beginPath();
		ctx.strokeStyle = "#333";
		ctx.lineWidth = 2;
		ctx.moveTo(-5, pad);
		ctx.lineTo(0, pad + 5);
		ctx.lineTo(5, pad);
		ctx.stroke();
		ctx.closePath();

		// Draw the text
		ctx.fillStyle = "white";
		let y = -height;
		for (let i = 0; i < texts.length; i++) {
			y += heights[i];
			let text = texts[i];
			ctx.fillText(text, -width / 2, y);
		}

		ctx.restore();
	}

	updateActors(actorsZOrder: Array<string>) {
		for (let key of actorsZOrder) {
			let actor = this.actors.get(key);
			actor.Update(this);
		}
	}

	bindForDraw() {
		let ctx = this.context;
		let gl = ctx.gl;

		// Clear the viewport
		let bg = new Color().setFromString(this.playerConfig.backgroundColor);
		gl.clearColor(bg.r, bg.g, bg.b, bg.a);
		gl.clear(gl.COLOR_BUFFER_BIT);

		// Resize the canvas
		let viewport = this.playerConfig.viewport;
		this.sceneRenderer.resize(ResizeMode.Expand);
		this.updateCameraSettings(this.sceneRenderer.camera, null, viewport);

		// Clear the Text Canvas
		if (this.textCanvas.width != this.textCanvas.clientWidth ||
			this.textCanvas.height != this.textCanvas.clientHeight
		) {
			this.textCanvas.width = this.textCanvas.clientWidth;
			this.textCanvas.height = this.textCanvas.clientHeight;
			this.textCanvasContext.font = this.playerConfig.textSize + "px " + this.playerConfig.textFont;
			this.textCanvasContext.textBaseline = "bottom";
		}
		this.textCanvasContext.clearRect(0, 0, this.textCanvas.width, this.textCanvas.height);
	}

	drawActors(actorsZOrder: Array<string>) {
		this.bindForDraw();

		this.sceneRenderer.begin();
		for (let key of actorsZOrder) {
			let actor = this.actors.get(key);
			actor.Draw(this.sceneRenderer)
		}
		this.sceneRenderer.end();

		// Render the debug output with a fixed camera.
		if (this.playerConfig.viewport.debugRender) {
			this.sceneRenderer.beginShapes();
			for (let key of actorsZOrder) {
				let actor = this.actors.get(key);
				actor.DrawDebug(this.sceneRenderer);
				this.sceneRenderer.circle(true, 0, 0, 0, 10, Color.BLUE);
			}
			this.sceneRenderer.end();
		}

		// Render all the speech bubbles
		for (let key of actorsZOrder) {
			let actor = this.actors.get(key);
			actor.DrawText(
				this.sceneRenderer.camera,
				this.textCanvasContext,
				this.playerConfig.showChatMessages)
		}
	}

	postProcessActors(actorsZOrder: Array<string>) {
		// We want to post-process all the animation state changes AFTER
		// we have fully rendered the frame. This is because the animation stage
		// changes use the same canvas GL context for calculating the accurate
		// bounding boxes and will produce artifacts on the screen if we use
		// it within the same draw frame.
		for (let key of actorsZOrder) {
			let actor = this.actors.get(key);
			actor.ProcessAnimationStateChange();
		}
	}

	lastDrawCall: number = null;
	drawFrame(requestNextFrame = true) {
		let startTime = new Date().getTime();

		// Order the actors to draw based on their z-order
		let all_actors_loaded = true;
		let actorsZOrder = Array
			.from(this.actors.keys()).sort((a: string, b: string) => {
				let a1 = this.actors.get(a);
				let a2 = this.actors.get(b);
				if (a1.ShouldFlashCharacter && !a2.ShouldFlashCharacter) {
					return 1;
				} else if (a2.ShouldFlashCharacter && !a1.ShouldFlashCharacter) {
					return -1;
				}
				let r = a1.getPositionZ() - a2.getPositionZ()
				// To ensure stable sort
				if (r == 0) { return a1.lastUpdatedWhen - a2.lastUpdatedWhen; }
				return r;
			})
			.filter(key => {
				let actor = this.actors.get(key);
				if (actor.load_perma_failed) {
					// Permanant failure trying to load this actor. Just skip it.
					return false;
				}
				if (this.assetManager.isLoadingComplete()) {
					if (!actor.isActorDoneLoading()) {
						this.loadActor(actor);
					}
				}
				if (!actor.loaded) {
					all_actors_loaded = false;
				}
				return actor.loaded;
			});
		this.calculateAverageActorHeights(actorsZOrder);

		this.updateActors(actorsZOrder);
		this.drawActors(actorsZOrder);
		this.postProcessActors(actorsZOrder);
		this.broadcastRenderCallback();

		if (all_actors_loaded) {
			this.assetManager.clearErrors();
			this.hideError();
		}

		let endTime = new Date().getTime();
		this.windowFpsFrameCount += 1;
		if (this.windowFpsFrameCount % 60 == 0) {
			// console.log("Frame time: " + (endTime - startTime) + "ms", "Frame delay: " + (startTime - this.lastDrawCall) + "ms");
		}
		this.lastDrawCall = endTime;

		if (requestNextFrame && !this.stopRequestAnimationFrame) {
			this.lastRequestAnimationFrameId = requestAnimationFrame(() => this.drawFrame());
		}
	}

	public calculateAverageActorHeights(actorsZOrder: string[]) {
		if (!this.playerConfig.excessiveChibiMitigations) {
			return;
		}
		if (actorsZOrder.length < 30) {
			// Only trigger using actor bands if we have lots of actors
			return;
		}
		if (!this.actorHeightDirty) {
			return;
		}
		let averageHeight = 0;
		let actorCount = 0;
		for (let key of actorsZOrder) {
			let actor = this.actors.get(key);
			averageHeight += actor.getRenderingBoundingBox().height
			actorCount += 1;
		}
		Actor.averageActorHeight = averageHeight / actorCount;
		this.actorHeightDirty = false;
	}

	public registerRenderCallback(callback: RenderCallbackFn): RemoveCallbackFn {
		this.renderCallbacks.push(callback);
		return () => {
			this.renderCallbacks = this.renderCallbacks.filter(c => c !== callback);
		}
	}

	broadcastRenderCallback() {
		for (let i = 0; i < this.renderCallbacks.length; i++) {
			this.renderCallbacks[i](this.context, this.textCanvasContext);;
		}
	}

	scale(sourceWidth: number, sourceHeight: number, targetWidth: number, targetHeight: number): Vector2 {
		let targetRatio = targetHeight / targetWidth;
		let sourceRatio = sourceHeight / sourceWidth;
		let scale = targetRatio > sourceRatio ? targetWidth / sourceWidth : targetHeight / sourceHeight;
		let temp = new Vector2();
		temp.x = sourceWidth * scale;
		temp.y = sourceHeight * scale;
		return temp;
	}

	loadSpineSkeleton(actor: Actor) {
		let atlas = this.assetManager.get(actor.config.atlasUrl);
		let skeletonData: SkeletonData;
		if (actor.config.jsonUrl) {
			let jsonText = this.assetManager.get(actor.config.jsonUrl);
			let json = new SkeletonJson(new AtlasAttachmentLoader(atlas));
			try {
				skeletonData = json.readSkeletonData(jsonText);
			} catch (e) {
				this.showError(actor, "Error: could not load skeleton .json.<br><br>" + escapeHtml(JSON.stringify(e)));
				return;
			}
		} else {
			let binaryData = this.assetManager.get(actor.config.skelUrl);
			let binary = new SkeletonBinary(new AtlasAttachmentLoader(atlas));
			try {
				skeletonData = binary.readSkeletonData(binaryData);
			} catch (e) {
				this.showError(actor, "Error: could not load skeleton .skel.<br><br>" + escapeHtml(JSON.stringify(e)));
				return;
			}
		}
		actor.skeleton = new Skeleton(skeletonData);

		// Setup skin
		if (!actor.config.skin) {
			if (skeletonData.skins.length > 0) {
				actor.config.skin = skeletonData.skins[0].name;
			}
		}

		if (actor.config.skins && actor.config.skin.length > 0) {
			actor.config.skins.forEach(skin => {
				if (!actor.skeleton.data.findSkin(skin)) {
					this.showError(actor, `Error: skin '${skin}' in selectable skin list does not exist in skeleton.`);
					return;
				}
			});
		}

		if (actor.config.skin) {
			if (!actor.skeleton.data.findSkin(actor.config.skin)) {
				this.showError(actor, `Error: skin '${actor.config.skin}' does not exist in skeleton.`);
				return;
			}
			actor.skeleton.setSkinByName(actor.config.skin);
			actor.skeleton.setSlotsToSetupPose();
		}
		actor.InitAnimationState();

		actor.config.success(this, actor);
		actor.loaded = true;
	}

	loadSpritesheetActor(actor: Actor) {
		let json = this.assetManager.get(actor.config.spritesheetDataUrl);
		let spritesheetConfig = readSpritesheetJsonConfig(json);

		let parentPrefix = actor.config.spritesheetDataUrl.split("/").slice(0, -1).join("/");
		let textures: Map<string, any> = new Map<string, any>();
		for (let [animationKey, config] of spritesheetConfig.animations) {
			let texture = this.assetManager.get(parentPrefix + "/" + config.filepath);
			textures.set(animationKey, texture);
		}

		let spritesheetActor = new SpritesheetActor(spritesheetConfig, textures);
		spritesheetActor.SetAnimation(actor.GetAnimations()[0]);
		actor.spritesheet = spritesheetActor;

		actor.InitAnimationState();
		actor.config.success(this, actor);
		actor.loaded = true;
	}

	loadActor(actor: Actor) {
		if (actor.loaded || actor.load_perma_failed) return;

		if (this.assetManager.hasErrors()) {
			this.showError(actor, "Error: assets could not be loaded.<br><br>" + escapeHtml(JSON.stringify(this.assetManager.getErrors())));
			return;
		}

		if (actor.isSpritesheetActor()) {
			this.loadSpritesheetActor(actor);
		} else {
			this.loadSpineSkeleton(actor);
		}
	}

	private cancelId: any = 0;
	private setupInput() {
		let canvas = this.canvas;
		let input = new Input(canvas);
		input.addListener({
			down: (x, y) => { },
			dragged: (x, y) => { },
			moved: (x, y) => { },
			up: (x, y) => {
				if (!this.playerConfig.showControls) return;
				if (this.paused) {
					this.play()
				} else {
					this.pause()
				}
			},
		});

		// For the manual hover to work, we need to disable
		// hidding the controls if the mouse/touch entered
		// the clickable area of a child of the controls.
		// For this we need to register a mouse handler on
		// the document and see if we are within the canvas
		// area :/
		var mouseOverControls = true;
		var mouseOverCanvas = false;
		document.addEventListener("mousemove", (ev: UIEvent) => {
			if (ev instanceof MouseEvent) {
				handleHover(ev.clientX, ev.clientY);
			}
		});
		document.addEventListener("touchmove", (ev: UIEvent) => {
			if (ev instanceof TouchEvent) {
				var touches = ev.changedTouches;
				if (touches.length > 0) {
					var touch = touches[0];
					handleHover(touch.clientX, touch.clientY);
				}
			}
		});

		let handleHover = (mouseX: number, mouseY: number) => {
			// if (!this.config.showControls) return;
			if (!this.playerConfig.showControls) return;

			let popup = findWithClass(this.dom, "spine-player-popup");
			mouseOverControls = overlap(mouseX, mouseY, this.playerControls.getBoundingClientRect());
			mouseOverCanvas = overlap(mouseX, mouseY, this.canvas.getBoundingClientRect());
			clearTimeout(this.cancelId);
			let hide = popup.length == 0 && !mouseOverControls && !mouseOverCanvas && !this.paused;
			if (hide) {
				this.playerControls.classList.add("spine-player-controls-hidden");
			} else {
				this.playerControls.classList.remove("spine-player-controls-hidden");
			}
			if (!mouseOverControls && popup.length == 0 && !this.paused) {
				let remove = () => {
					if (!this.paused) this.playerControls.classList.add("spine-player-controls-hidden");
				};
				this.cancelId = setTimeout(remove, 1000);
			}
		}

		let overlap = (mouseX: number, mouseY: number, rect: DOMRect | ClientRect): boolean => {
			let x = mouseX - rect.left;
			let y = mouseY - rect.top;
			return x >= 0 && x <= rect.width && y >= 0 && y <= rect.height;
		}
	}

	private play() {
		this.paused = false;
		let remove = () => {
			if (!this.paused) this.playerControls.classList.add("spine-player-controls-hidden");
		};
		this.cancelId = setTimeout(remove, 1000);
		this.playButton.classList.remove("spine-player-button-icon-play");
		this.playButton.classList.add("spine-player-button-icon-pause");

		for (let [key, actor] of this.actors) {
			actor.paused = false;
			if (actor.animationState) {
				if (!actor.animationState.getCurrent(0)) {
					actor.InitAnimations();
					// this.setAnimations(actor, actor.GetAnimations());
				}
			}
		}
	}

	private pause() {
		this.paused = true;
		this.playerControls.classList.remove("spine-player-controls-hidden");
		clearTimeout(this.cancelId);

		this.playButton.classList.remove("spine-player-button-icon-pause");
		this.playButton.classList.add("spine-player-button-icon-play");

		for (let [key, actor] of this.actors) {
			actor.paused = true;
		}
	}

	// Exposed only for testing.
	public getSceneRenderer(): SceneRenderer {
		return this.sceneRenderer;
	}
}
