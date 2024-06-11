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

module spine {
	export interface Viewport {
		x: number,
		y: number,
		width: number,
		height: number,
		padLeft: string | number
		padRight: string | number
		padTop: string | number
		padBottom: string | number
	}

	export interface BoundingBox {
		x: number,
		y: number,
		width: number,
		height: number,
	}

	export interface SpinePlayerConfig {
		/* Optional: whether to show the player controls. Default: true. */
		showControls: boolean

		/* Optional: the position and size of the viewport in world coordinates of the skeleton. Default: the setup pose bounding box. */
		viewport: {
			x: number
			y: number
			width: number
			height: number
			padLeft: string | number
			padRight: string | number
			padTop: string | number
			padBottom: string | number
			animations: Map<Viewport>
			debugRender: boolean,
			transitionTime: number
		}

		/* Optional: whether the canvas should be transparent. Default: false. */
		alpha: boolean

		/* Optional: the background color. Must be given in the format #rrggbbaa. Default: #000000ff. */
		backgroundColor: string

		/* Optional: the background image. Default: none. */
		backgroundImage: {
			/* The URL of the background image */
			url: string

			/* Optional: the position and size of the background image in world coordinates. Default: viewport. */
			x: number
			y: number
			width: number
			height: number
		}

		/* Optional: the background color used in fullscreen mode. Must be given in the format #rrggbbaa. Default: backgroundColor. */
		fullScreenBackgroundColor: string
	}

	class Slider {
		private slider: HTMLElement;
		private value: HTMLElement;
		private knob: HTMLElement;
		public change: (percentage: number) => void;

		constructor(public snaps = 0, public snapPercentage = 0.1, public big = false) { }

		render(): HTMLElement {
			this.slider = createElement(/*html*/`
				<div class="spine-player-slider ${this.big ? "big": ""}">
					<div class="spine-player-slider-value"></div>
					<!--<div class="spine-player-slider-knob"></div>-->
				</div>
			`);
			this.value = findWithClass(this.slider, "spine-player-slider-value")[0];
			// this.knob = findWithClass(this.slider, "spine-player-slider-knob")[0];
			this.setValue(0);

			let input = new spine.webgl.Input(this.slider);
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

	export class SpinePlayer {
		static HOVER_COLOR_INNER = new spine.Color(0.478, 0, 0, 0.25);
		static HOVER_COLOR_OUTER = new spine.Color(1, 1, 1, 1);
		static NON_HOVER_COLOR_INNER = new spine.Color(0.478, 0, 0, 0.5);
		static NON_HOVER_COLOR_OUTER = new spine.Color(1, 0, 0, 0.8);

		private sceneRenderer: spine.webgl.SceneRenderer;
		private dom: HTMLElement;
		private playerControls: HTMLElement;
		private canvas: HTMLCanvasElement;
		private textCanvas: HTMLCanvasElement;
		private textCanvasContext: CanvasRenderingContext2D;
		private timelineSlider: Slider;
		private playButton: HTMLElement;

		private context: spine.webgl.ManagedWebGLRenderingContext;
		private loadingScreen: spine.webgl.LoadingScreen;
		private assetManager: spine.webgl.AssetManager;

		private paused = true;
		private actors = new Map<string, Actor>();
		private playerConfig: SpinePlayerConfig = null;
		
		private parent: HTMLElement;

		private stopRequestAnimationFrame = false;
		private lastRequestAnimationFrameId = 0;

		constructor(parent: HTMLElement | string, playerConfig: SpinePlayerConfig) {
			if (typeof parent === "string") {
				this.parent = document.getElementById(parent);
			} else  {
				this.parent = parent;
			}
			// this.playerConfig = null;
			this.playerConfig = this.validatePlayerConfig(playerConfig);
			this.parent.appendChild(this.setupDom());
		}

		validatePlayerConfig(config: SpinePlayerConfig): SpinePlayerConfig {
			if (!config) throw new Error("Please pass a configuration to new.spine.SpinePlayer().");
			if (!config.alpha) config.alpha = false;
			if (!config.backgroundColor) config.backgroundColor = "#000000";
			if (!config.fullScreenBackgroundColor) config.fullScreenBackgroundColor = config.backgroundColor;
			if (typeof config.showControls === "undefined")
				config.showControls = true;
			return config;
		}

		validateActorConfig(config: SpineActorConfig): SpineActorConfig {
			if (!config) throw new Error("Please pass a configuration to new.spine.SpinePlayer().");
			if (!config.jsonUrl && !config.skelUrl) throw new Error("Please specify the URL of the skeleton JSON or .skel file.");
			if (!config.atlasUrl) throw new Error("Please specify the URL of the atlas file.");
			// if (!config.alpha) config.alpha = false;
			// if (!config.backgroundColor) config.backgroundColor = "#000000";
			// if (!config.fullScreenBackgroundColor) config.fullScreenBackgroundColor = config.backgroundColor;
			if (typeof config.premultipliedAlpha === "undefined") config.premultipliedAlpha = true;
			if (!config.scaleX) config.scaleX = 1;
			if (!config.scaleY) config.scaleY = 1;
			if (!config.success) config.success = (widget) => {};
			if (!config.error) config.error = (widget, msg) => {};
			if (!config.animation_listener) {
				config.animation_listener = {
					event: function(trackIndex, event) {
						// console.log("Event on track " + trackIndex + ": " + JSON.stringify(event));
					},
					complete: function(trackIndex) {
						// console.log("Animation on track " + trackIndex + " completed, loop count: " + loopCount);
					},
					start: function(trackIndex) {
						// console.log("Animation on track " + trackIndex + " started");
					},
					end: function(trackIndex) {
						// console.log("Animation on track " + trackIndex + " ended");
					},
					interrupt: function(trackIndex) {
						// console.log("Animation on track " + trackIndex + " ended");
					},
					dispose: function(trackIndex) {
						// console.log("Animation on track " + trackIndex + " ended");
					},
				}
			}
			if (config.animations && config.animation) {
				if (config.animations.indexOf(config.animation) < 0) throw new Error("Default animation '" + config.animation + "' is not contained in the list of selectable animations " + escapeHtml(JSON.stringify(config.animations)) + ".");
			}

			if (config.skins && config.skin) {
				if (config.skins.indexOf(config.skin) < 0) throw new Error("Default skin '" + config.skin + "' is not contained in the list of selectable skins " + escapeHtml(JSON.stringify(config.skins)) + ".");
			}
			if (typeof config.defaultMix === "undefined")
				config.defaultMix = 0.25;

			return config;
		}

		showError(actor: Actor, error: string) {
			let errorDom = findWithClass(this.dom, "spine-player-error")[0];
			errorDom.classList.remove("spine-player-hidden");
			errorDom.innerHTML = `<p style="text-align: center; align-self: center;">${error}</p>`;

			// this.playerConfig.error(this, error);
			actor.config.error(this, error);
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
				// var webglConfig = { alpha: config.alpha };
				var webglConfig = { alpha: this.playerConfig.alpha};
				this.context = new spine.webgl.ManagedWebGLRenderingContext(this.canvas, webglConfig);
				// Setup the scene renderer and loading screen
				this.sceneRenderer = new spine.webgl.SceneRenderer(this.canvas, this.context, true);
				this.loadingScreen = new spine.webgl.LoadingScreen(this.sceneRenderer);
			} catch (e) {
				// this.showError("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
				console.log("Sorry, your browser does not support WebGL.<br><br>Please use the latest version of Firefox, Chrome, Edge, or Safari.");
				return dom;
			}

			// Load the assets
			this.assetManager = new spine.webgl.AssetManager(this.context);

			// Setup rendering loop
			this.lastRequestAnimationFrameId = requestAnimationFrame(() => this.drawFrame());

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

		changeOrAddActor(actorName: string, config: SpineActorConfig) {
			if (this.actors.has(actorName)) {
				let actor = this.actors.get(actorName);
				actor.ResetWithConfig(config);
				this.setupActor(actor);
			} else {
				let actor = new Actor(config);
				this.setupActor(actor);
				this.actors.set(actorName, actor);
			}
		}

		removeActor(actorName: string) {
			this.actors.delete(actorName)
		}

		updateActor(actorName: string, config) {
			if (!this.actors.has(actorName)) {
				return;
			}
			let actor = this.actors.get(actorName);
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
			if (config.jsonUrl) this.assetManager.loadText(config.jsonUrl);
			else this.assetManager.loadBinary(config.skelUrl);
			this.assetManager.loadTextureAtlas(config.atlasUrl);
			if (config.backgroundImage && config.backgroundImage.url)
				this.assetManager.loadTexture(config.backgroundImage.url);

			// Setup rendering loop
			// this.lastRequestAnimationFrameId = requestAnimationFrame(() => this.drawFrame());
		}

		drawText(text: string, xpx: number, ypx: number) {
			let viewport = this.playerConfig.viewport;
			let textpos = this.sceneRenderer.camera.worldToScreen(
				new spine.Vector2(xpx, ypx)
			);
			textpos.y = viewport.height - textpos.y;
			xpx = textpos.x;
			ypx = textpos.y;

			let ctx = this.textCanvasContext;
			ctx.save();
			ctx.translate(xpx, ypx);

			ctx.font = "16px lato";
			ctx.textBaseline = "bottom";
			let data = ctx.measureText(text);
			let height = data.actualBoundingBoxAscent - data.actualBoundingBoxDescent;
			let width = data.width;

			ctx.beginPath();
			let pad = 5;
			ctx.fillStyle = "black";
			ctx.fillRect(-width/2 - pad, pad, width + 2*pad, -height - 2*pad);
			ctx.fill();

			ctx.fillStyle = "white";
			ctx.fillText(text, -width/2, 0);
			ctx.restore();
		}

		drawFrame (requestNextFrame = true) {
			if (requestNextFrame && !this.stopRequestAnimationFrame) {
				this.lastRequestAnimationFrameId = requestAnimationFrame(() => this.drawFrame());
			}
			let ctx = this.context;
			let gl = ctx.gl;

			// Clear the viewport
			var doc = document as any;
			var isFullscreen = doc.fullscreenElement || doc.webkitFullscreenElement || doc.mozFullScreenElement || doc.msFullscreenElement;
			// let bg = new Color().setFromString(isFullscreen ? this.config.fullScreenBackgroundColor : this.config.backgroundColor);
			let bg = new Color().setFromString(
				isFullscreen ? 
				this.playerConfig.fullScreenBackgroundColor :
				this.playerConfig.backgroundColor
			);
			gl.clearColor(bg.r, bg.g, bg.b, bg.a);
			gl.clear(gl.COLOR_BUFFER_BIT);

			// Clear the Text Canvas
			this.textCanvas.width = this.textCanvas.clientWidth;
			this.textCanvas.height = this.textCanvas.clientHeight;
			this.textCanvasContext.clearRect(0,0, this.textCanvas.width, this.textCanvas.height);

			// Display loading screen
			this.loadingScreen.backgroundColor.setFromColor(bg);
			this.loadingScreen.draw(this.assetManager.isLoadingComplete());

			let all_actors_loaded = true;
			for (let [key,actor] of this.actors) {
				// Have we finished loading the asset? Then set things up
				// if (this.assetManager.isLoadingComplete() && this.skeleton == null) this.loadSkeleton();
				if (this.assetManager.isLoadingComplete() && actor.skeleton == null) {
					this.loadSkeleton(actor);
				}
					
				// Resize the canvas
				this.sceneRenderer.resize(webgl.ResizeMode.Expand);

				// Update and draw the skeleton
				if (!actor.loaded) {
					all_actors_loaded = false;
					continue;
				}

				let viewport = this.playerConfig.viewport;

				// Update animation and skeleton based on user selections
				if (!actor.paused && actor.config.animation) {
					actor.time.update();
					let delta = actor.time.delta * actor.speed;

					let animationDuration = actor.animationState.getCurrent(0).animation.duration;
					actor.playTime += delta;
					while (actor.playTime >= animationDuration && animationDuration != 0) {
						actor.playTime -= animationDuration;
					}
					actor.playTime = Math.max(0, Math.min(actor.playTime, animationDuration));
					this.timelineSlider.setValue(actor.playTime / animationDuration);

					actor.UpdatePhysics(delta, viewport);
					actor.skeleton.setToSetupPose();					
					actor.animationState.update(delta);
					actor.animationState.apply(actor.skeleton);
				}
				actor.skeleton.updateWorldTransform();

				// Update the camera
				let viewportSize = this.scale(viewport.width, viewport.height, this.canvas.width, this.canvas.height);
				this.sceneRenderer.camera.zoom = viewport.width / viewportSize.x;
				this.sceneRenderer.camera.position.x = viewport.x;
				this.sceneRenderer.camera.position.y = viewport.y + viewport.height / 2;
				this.sceneRenderer.camera.position.z = 0;
				
				// HACK: Code for using projection perspetive for the camera	
				// this.sceneRenderer.camera.near = 1;
				// this.sceneRenderer.camera.near = 10000;
				// this.sceneRenderer.camera.position.z = 1320;
				// let origin = new spine.webgl.Vector3(0,0,0);
				// let pos = new spine.webgl.Vector3(
				// 	this.sceneRenderer.camera.position.x,
				// 	0,
				// 	this.sceneRenderer.camera.position.z,
				// );
				// let dir = origin.sub(pos).normalize();
				// this.sceneRenderer.camera.direction = dir;
				

				this.sceneRenderer.begin();
				// // Draw background image if given
				if (actor.config.backgroundImage && actor.config.backgroundImage.url) {
					let bgImage = this.assetManager.get(actor.config.backgroundImage.url);
					if (!(actor.config.backgroundImage.hasOwnProperty("x") && actor.config.backgroundImage.hasOwnProperty("y") && actor.config.backgroundImage.hasOwnProperty("width") && actor.config.backgroundImage.hasOwnProperty("height"))) {
						this.sceneRenderer.drawTexture(bgImage, viewport.x, viewport.y, viewport.width, viewport.height);
					} else {
						this.sceneRenderer.drawTexture(bgImage, actor.config.backgroundImage.x, actor.config.backgroundImage.y, actor.config.backgroundImage.width, actor.config.backgroundImage.height);
					}
				}
				// Draw skeleton 
				this.sceneRenderer.drawSkeleton(actor.skeleton, actor.config.premultipliedAlpha);

				// Render the user's name above the chibi
				this.drawText(
					actor.config.userDisplayName,
					actor.position.x,
					actor.defaultBB.height - 20,
				)
				this.sceneRenderer.end();

				// Render the debug output with a fixed camera.
				if (this.playerConfig.viewport.debugRender) {
					this.sceneRenderer.begin();

					if (actor.scale.x > 0) {
						this.sceneRenderer.rect(
							false, 
							actor.position.x + actor.animViewport.x,
							actor.position.y + actor.animViewport.y,
							actor.animViewport.width,
							actor.animViewport.height,
							Color.BLUE
						);					
					} else {
						this.sceneRenderer.rect(
							false, 
							actor.position.x - (actor.animViewport.width + actor.animViewport.x),
							actor.position.y + actor.animViewport.y,
							actor.animViewport.width,
							actor.animViewport.height,
							Color.GREEN
						);
					}

					this.sceneRenderer.rect(
						false,
						actor.position.x + actor.defaultBB.x,
						actor.position.y + actor.defaultBB.y,
						actor.defaultBB.width,
						actor.defaultBB.height,
						Color.ORANGE
					)
					
					if (actor.endPosition) {
						this.sceneRenderer.line(actor.endPosition.x, 0, actor.endPosition.x, 2000, Color.WHITE);
					}
					this.sceneRenderer.circle(true, actor.position.x, actor.position.y, 10, Color.RED);

					this.sceneRenderer.line(0, 0, 0, 2000, Color.RED);
					this.sceneRenderer.line(0, 540, 1920, 540, Color.RED);
					this.sceneRenderer.line(0, 270, 1920, 270, Color.RED);
					this.sceneRenderer.line(0, actor.animViewport.height, 1920, actor.animViewport.height, Color.RED);
					this.sceneRenderer.end();
				}

				
				this.sceneRenderer.camera.zoom = 0;
			}

			if (all_actors_loaded) {
				this.assetManager.clearErrors();
				this.hideError();
			}
		}

		scale(sourceWidth: number, sourceHeight: number, targetWidth: number, targetHeight: number): Vector2 {
			let targetRatio = targetHeight / targetWidth;
			let sourceRatio = sourceHeight / sourceWidth;
			let scale = targetRatio > sourceRatio ? targetWidth / sourceWidth : targetHeight / sourceHeight;
			let temp = new spine.Vector2();
			temp.x = sourceWidth * scale;
			temp.y = sourceHeight * scale;
			return temp;
		}

		loadSkeleton (actor: Actor) {
			if (actor.loaded) return;

			if (this.assetManager.hasErrors()) {
				this.showError(actor, "Error: assets could not be loaded.<br><br>" + escapeHtml(JSON.stringify(this.assetManager.getErrors())));
				return;
			}

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
			let stateData = new AnimationStateData(skeletonData);
			stateData.defaultMix = actor.config.defaultMix;
			actor.animationState = new AnimationState(stateData);
			if (actor.config.animation_listener) {
				actor.animationState.addListener(actor.config.animation_listener);
			}
			
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

			// Setup empty viewport if none is given and check
			// if all animations for which viewports where given
			// exist.
			if (!actor.config.viewport) {
				(actor.config.viewport as any) = {
					animations: {},
					debugRender: false,
					transitionTime: 0.2
				}
			}
			if (typeof actor.config.viewport.debugRender === "undefined") actor.config.viewport.debugRender = false;
			if (typeof actor.config.viewport.transitionTime === "undefined") actor.config.viewport.transitionTime = 0.2;
			if (!actor.config.viewport.animations) {
				actor.config.viewport.animations = {};
			} else {
				Object.getOwnPropertyNames(actor.config.viewport.animations).forEach((animation: string) => {
					if (!skeletonData.findAnimation(animation)) {
						this.showError(actor, `Error: animation '${animation}' for which a viewport was specified does not exist in skeleton.`);
						return;
					}
				});
			}

			// Setup the animations after viewport, so default bounds don't get messed up.
			if (actor.config.animations && actor.config.animations.length > 0) {
				actor.config.animations.forEach(animation => {
					if (!actor.skeleton.data.findAnimation(animation)) {
						this.showError(actor, `Error: animation '${animation}' in selectable animation list does not exist in skeleton.`);
						return;
					}
				});

				if (!actor.config.animation) {
					actor.config.animation = actor.config.animations[0];
				}
			}

			if (!actor.config.animation) {
				if (skeletonData.animations.length > 0) {
					actor.config.animation = skeletonData.animations[0].name;
				}
			}

			if(actor.config.animation) {
				if (!skeletonData.findAnimation(actor.config.animation)) {
					this.showError(actor, `Error: animation '${actor.config.animation}' does not exist in skeleton.`);
					return;
				}
				this.play()
				this.timelineSlider.change = (percentage) => {
					this.pause();
					var animationDuration = actor.animationState.getCurrent(0).animation.duration;
					var time = animationDuration * percentage;
					actor.animationState.update(time - actor.playTime);
					actor.animationState.apply(actor.skeleton);
					actor.skeleton.updateWorldTransform();
					actor.playTime = time;
				}
			}

			actor.config.success(this);
			actor.loaded = true;
		}

		private cancelId = 0;
		setupInput () {
			let canvas = this.canvas;
			let input = new spine.webgl.Input(canvas);
			input.addListener({
				down: (x, y) => {},
				dragged: (x, y) => {},
				moved: (x, y) => {},
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

		private play () {
			this.paused = false;
			let remove = () => {
				if (!this.paused) this.playerControls.classList.add("spine-player-controls-hidden");
			};
			this.cancelId = setTimeout(remove, 1000);
			this.playButton.classList.remove("spine-player-button-icon-play");
			this.playButton.classList.add("spine-player-button-icon-pause");

			for(let [key,actor] of this.actors) {
				actor.paused = false;
				if (actor.config.animation) {
					if (!actor.animationState.getCurrent(0)) {
						this.setAnimation(actor, actor.config.animation);
					}
				}
			}
		}

		private pause () {
			this.paused = true;
			this.playerControls.classList.remove("spine-player-controls-hidden");
			clearTimeout(this.cancelId);

			this.playButton.classList.remove("spine-player-button-icon-pause");
			this.playButton.classList.add("spine-player-button-icon-play");

			for(let [key, actor] of this.actors) {
				actor.paused = true;
			}
		}

		public setAnimation (actor: Actor, animation: string) {
			// Determine viewport
			let animViewport = this.calculateAnimationViewport(actor, animation);
			console.log("animViewport: " + JSON.stringify(animViewport));

			actor.animViewport = animViewport;
			actor.defaultBB = this.getDefaultBoundingBox(actor);
			actor.animationState.clearTracks();
			actor.skeleton.setToSetupPose();
			actor.animationState.setAnimation(0, animation, true);

			if (actor.config.animation.includes("Move")) {
				actor.setDestination();
			} else {
				actor.clearDestination();
			}
			if (actor.config.animation == "Sit") {
				actor.position.y = Math.abs(actor.animViewport.y)
			} else {
				actor.position.y = 0;
			}
		}

		private getDefaultBoundingBox (actor: Actor) {
			let animationName;
			let animations = actor.skeleton.data.animations;
			for (let i = 0, n = animations.length; i < n; i++) {
				let animName = animations[i].name;
				if (animName.toLocaleLowerCase().includes("default")) {
					animationName = animName;
					break;
				}
			}

			let animation = actor.skeleton.data.findAnimation(animationName);
			actor.animationState.clearTracks();
			actor.skeleton.setToSetupPose()
			actor.animationState.setAnimationWith(0, animation, true);

			actor.skeleton.x = 0;
			actor.skeleton.y = 0;
			actor.skeleton.scaleX = Math.abs(actor.config.scaleX);
			actor.skeleton.scaleY = Math.abs(actor.config.scaleY);
			actor.animationState.update(0);
			actor.animationState.apply(actor.skeleton);
			actor.skeleton.updateWorldTransform();
			let offset = new spine.Vector2();
			let size = new spine.Vector2();
			actor.skeleton.getBounds(offset, size);

			return {
				x: offset.x,
				y: offset.y,
				width: size.x,
				height: size.y
			}
		}

		private calculateAnimationViewport (actor: Actor, animationName: string) {
			let animation = actor.skeleton.data.findAnimation(animationName);
			actor.animationState.clearTracks();
			actor.skeleton.setToSetupPose()
			actor.animationState.setAnimationWith(0, animation, true);

			let steps = 100;
			let stepTime = animation.duration > 0 ? animation.duration / steps : 0;
			let minX = 100000000;
			let maxX = -100000000;
			let minY = 100000000;
			let maxY = -100000000;
			let offset = new spine.Vector2();
			let size = new spine.Vector2();

			for (var i = 0; i < steps; i++) {
				actor.animationState.update(stepTime);
				actor.animationState.apply(actor.skeleton);

				// TODO: Fix this hack
				actor.setSkeletonMovementData(this.playerConfig.viewport);
				actor.skeleton.x = 0;
				actor.skeleton.y = 0;
				actor.skeleton.scaleX = Math.abs(actor.config.scaleX);
				actor.skeleton.scaleY = Math.abs(actor.config.scaleY);
				actor.skeleton.updateWorldTransform();
				actor.skeleton.getBounds(offset, size);

				minX = Math.min(offset.x, minX);
				maxX = Math.max(offset.x + size.x, maxX);
				minY = Math.min(offset.y, minY);
				maxY = Math.max(offset.y + size.y, maxY);
			}

			offset.x = minX;
			offset.y = minY;
			size.x = maxX - minX;
			size.y = maxY - minY;

			return {
				x: offset.x,
				y: offset.y,
				width: size.x,
				height: size.y
			};
		}
	}

	function isContained(dom: HTMLElement, needle: HTMLElement): boolean {
		if (dom === needle) return true;
		let findRecursive = (dom: HTMLElement, needle: HTMLElement) => {
			for(var i = 0; i < dom.children.length; i++) {
				let child = dom.children[i] as HTMLElement;
				if (child === needle) return true;
				if (findRecursive(child, needle)) return true;
			}
			return false;
		};
		return findRecursive(dom, needle);
	}

	function findWithId(dom: HTMLElement, id: string): HTMLElement[] {
		let found = new Array<HTMLElement>()
		let findRecursive = (dom: HTMLElement, id: string, found: HTMLElement[]) => {
			for(var i = 0; i < dom.children.length; i++) {
				let child = dom.children[i] as HTMLElement;
				if (child.id === id) found.push(child);
				findRecursive(child, id, found);
			}
		};
		findRecursive(dom, id, found);
		return found;
	}

	function findWithClass(dom: HTMLElement, className: string): HTMLElement[] {
		let found = new Array<HTMLElement>()
		let findRecursive = (dom: HTMLElement, className: string, found: HTMLElement[]) => {
			for(var i = 0; i < dom.children.length; i++) {
				let child = dom.children[i] as HTMLElement;
				if (child.classList.contains(className)) found.push(child);
				findRecursive(child, className, found);
			}
		};
		findRecursive(dom, className, found);
		return found;
	}

	function createElement(html: string): HTMLElement {
		let dom = document.createElement("div");
		dom.innerHTML = html;
		return dom.children[0] as HTMLElement;
	}

	function removeClass(elements: HTMLCollection, clazz: string) {
		for (var i = 0; i < elements.length; i++) {
			elements[i].classList.remove(clazz);
		}
	}

	function escapeHtml(str: string) {
		if (!str) return "";
		return str
			 .replace(/&/g, "&amp;")
			 .replace(/</g, "&lt;")
			 .replace(/>/g, "&gt;")
			 .replace(/"/g, "&#34;")
			 .replace(/'/g, "&#39;");
	 }
 }
