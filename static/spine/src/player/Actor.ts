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

    export interface ActorUpdateConfig {
        start_pos: spine.Vector2
        dest_pos: spine.Vector2
        loop_positions: boolean
        wander: boolean
        animation: string
    }

	export interface SpineActorConfig {
		/* the URL of the skeleton .json file */
		jsonUrl?: string

		/* the URL of the skeleton .skel file */
		skelUrl: string

		/* the URL of the skeleton .atlas file. Atlas page images are automatically resolved. */
		atlasUrl: string

		/* Raw data URIs, mapping from a path to base 64 encoded raw data. When the player
		   resolves a path of the `jsonUrl`, `skelUrl`, `atlasUrl`, or the image paths
		   referenced in the atlas, it will first look for that path in this array of
		   raw data URIs. This allows embedding of resources directly in HTML/JS. */
		rawDataURIs?: Map<string>

		/* Optional: the default mix time used to switch between two animations. */
		defaultMix?: number

		/* Optional: the name of the skin to be set. Default: the default skin. */
		skin?: string

		/* Optional: list of skin names from which the user can choose. */
		skins?: string[]

		/* Optional: whether the skeleton uses premultiplied alpha. Default: true. */
		premultipliedAlpha: boolean

		/** Optional: Animation play speed. 0 to 3.0 */
		animationPlaySpeed: number,

		scaleX: number
		scaleY: number
		maxSizePx: number
		startPosX : number
		startPosY : number

		// Unused?
		defaultScaleX?: number
		defaultScaleY?: number
		extraOffsetX: number,
		extraOffsetY: number,

		/* Optional: the background image. Default: none. */
		backgroundImage?: {
			/* The URL of the background image */
			url: string

			/* Optional: the position and size of the background image in world coordinates. Default: viewport. */
			x: number
			y: number
			width: number
			height: number
		}

		// /* Optional: the background color used in fullscreen mode. Must be given in the format #rrggbbaa. Default: backgroundColor. */
		// fullScreenBackgroundColor: string

		// /* Optional: list of bone names that the user can control by dragging. */
		// controlBones: string[]

		/* Optional: callback when the widget and its assets have been successfully loaded. */
		success?: (widget: SpinePlayer) => void

		/* Optional: callback when the widget could not be loaded. */
		error?: (widget: SpinePlayer, msg: string) => void

		/** Optional: Callbacks for the animations */
		animation_listener?: spine.AnimationStateListener

		userDisplayName: string,

		/** char_002_amiya, enemy_1526_sfsui, etc */
		chibiId: string,

		/** Required: action name */
		action: string
		/** Required: A json object of the action data */
		action_data: any
	}


	export class Actor {
		public loaded: boolean;
		public skeleton: Skeleton;
		public animationState: AnimationState;
		public time = new TimeKeeper();
		public paused = false;
		public playTime = 0;
		public speed = 1;
		public config: SpineActorConfig;
		public lastAnimation: string = null;
		// public animations: string[]

		public viewport: BoundingBox = null;
		public animViewport: BoundingBox = null;
		public prevAnimViewport: BoundingBox = null;
		public defaultBB: BoundingBox = null;

		// Velcity in world coordinates space (units/second)
		public movementSpeed: spine.Vector2 = new spine.Vector2();
		public position: spine.Vector2 = new spine.Vector2();
		public scale: spine.Vector2 = new spine.Vector2(1,1);
		public velocity: spine.Vector2 = new spine.Vector2(0, 0);
		public startPosition: spine.Vector2 = null;
		public currentAction: ActorAction = null;

		constructor(config: SpineActorConfig, viewport: BoundingBox) {
			this.ResetWithConfig(config);
			this.movementSpeed = new spine.Vector2(80 + Math.random()*40,0);  // ~100 px/ second
			let x = Math.random()*viewport.width - (viewport.width/2)
			this.position = new spine.Vector2(x, 0);
			this.viewport = viewport;

			if (config.startPosX || config.startPosY) {
				this.position = new spine.Vector2(
					(config.startPosX* viewport.width) - (viewport.width/2),
					config.startPosY * viewport.height
				);
			}
		}

		public InitAnimations() {
			this.initAnimationsInternal(this.currentAction.GetAnimations());
		}

		public GetAnimations(): string[] {
			return this.currentAction.GetAnimations()
		}

		public ResetWithConfig(config: SpineActorConfig) {
			this.loaded = false;
			this.skeleton = null;
			this.animationState = null;
			this.time = new TimeKeeper();
			this.paused = false;
			this.playTime = 0;
			this.speed = 1;
			this.config = config;

			this.animViewport = null;
			this.prevAnimViewport = null;
			this.defaultBB = null;

			// current bearing should be kept the same, event as we change
			// the configuration/animation
			this.scale.x = Math.sign(this.scale.x) * config.scaleX;
			this.scale.y = Math.sign(this.scale.y) * config.scaleY;

			this.currentAction = ParseActionNameToAction(
				this.config.action,
				this.config.action_data
			);
		}

		public UpdatePhysics(deltaSecs: number, viewport: BoundingBox) {
			this.currentAction.UpdatePhysics(this, deltaSecs, viewport);
			this.position.x += this.velocity.x;
			this.position.y += this.velocity.y;

			// Make the Actor face right or left depending on velocity
			if (this.velocity.x != 0) {
				if (this.velocity.x > 0) {
					this.scale.x = this.config.scaleX;
				} else {
					this.scale.x = -this.config.scaleX;
				}
			}
			this.setSkeletonMovementData(viewport);
		}

		public InitAnimationState() {
			let skeletonData = this.skeleton.data;
			let stateData = new AnimationStateData(skeletonData);
			stateData.defaultMix = this.config.defaultMix;
			this.animationState = new AnimationState(stateData);
			if (this.config.animation_listener) {
				this.animationState.clearListeners();
				this.animationState.addListener(this.config.animation_listener);
			}
			if(this.GetAnimations()) {
				// TODO: Verify all the animations are found in the skeleton data
				if (this.animationState) {
					if (!this.animationState.getCurrent(0)) {
						this.InitAnimations();
					}
				}
			}
		}
		
		public GetUsernameHeaderHeight() {
			// let avg_width = (51.620900749705015 * this.config.scaleX) / 0.15;
			// let avg_height = (60.99747806746469 * this.config.scaleY) / 0.15;
			let finalHeight = 0;
			if (this.config.chibiId.includes("enemy_")) {
				finalHeight = this.defaultBB.height;
			} else {
				// HACK: Sometimes the chibi bounding box is too big and
				// doesn't reasonablly fit the chibi. In order to make
				// sure the name tags are placed at a reasonable height to
				// the sprite we use a heuristic which was empiracally found
				// based off the average height of all the operator chibis.
				let avg_height = (70.0 * this.config.scaleY) / 0.15;
				let height = this.defaultBB.height;
				if (height > avg_height) {
					finalHeight =  avg_height;
				} else {
					finalHeight =  height;
				}
			}
			return finalHeight + Math.abs(this.config.extraOffsetY);
		}

		// Privates
		// ---------------------------------------------------------------------

		private setSkeletonMovementData(viewport: BoundingBox) {
			this.skeleton.x = this.position.x + this.config.extraOffsetX;
			this.skeleton.y = this.position.y + this.config.extraOffsetY;
			this.skeleton.scaleX = this.scale.x;
			this.skeleton.scaleY = this.scale.y;
		}

		private recordAnimation(animation: string) {
			this.currentAction.SetAnimation(this, animation, this.viewport);
			this.lastAnimation = animation;
		}

		private initAnimationsInternal(animations: string[]) {
			this.setAnimationState(animations);
			this.recordAnimation(animations[0]);
						
			// Resize very large chibis to more reasonable sizes
			let width = this.defaultBB.width;
			let height = this.defaultBB.height;
			let maxSize = width;
			if (height > maxSize) {
				maxSize = height;
			}
			if (maxSize> this.config.maxSizePx && this.config.chibiId.includes("enemy_")) {
				console.log("Resizing actor to " + this.config.maxSizePx);
				let ratio = width / height;

				let xNew = 0;
				let yNew = 0;
				if (height > width) {
					xNew = ratio * this.config.maxSizePx;
					yNew = this.config.maxSizePx;
				} else {
					xNew = this.config.maxSizePx;
					yNew = this.config.maxSizePx / ratio;
				}
				let newScaleX = (xNew * this.config.scaleX) / width;
				let newScaleY = (yNew * this.config.scaleY) / height;

				// TODO: setting defaultScale is a bug. We need to save it only once
				this.config.defaultScaleX = this.config.scaleX;
				this.config.defaultScaleY = this.config.scaleY;
				this.config.scaleX = newScaleX;
				this.config.scaleY = newScaleY;
				this.scale.x = Math.sign(this.scale.x) * this.config.scaleX;
				this.scale.y = Math.sign(this.scale.y) * this.config.scaleY;

				this.setAnimationState(animations);
				this.recordAnimation(animations[0]);
				// this.startAnimCallback(actor, animations[0]);
			}
		}

		private setAnimationState(animations: string[]) {
			let animation = animations[0];
			// Determine viewport
			this.animViewport = this.calculateAnimationViewport(animation);
			this.defaultBB = this.getDefaultBoundingBox();
			this.animationState.timeScale = this.config.animationPlaySpeed;
			this.animationState.clearTracks();
			this.skeleton.setToSetupPose();
			
			this.animationState.setAnimation(0, animation, true);
			let lastTrackEntry = null;
			for (let i = 1; i < animations.length; i++) {
				lastTrackEntry = this.animationState.addAnimation(0, animations[i], false, 0);
			}

			if (lastTrackEntry) {
				this.animationState.addListener({
					start: (
						(trackEntry: TrackEntry) => {
							this.recordAnimation(trackEntry.animation.name,)
						}
					).bind(this),
					interrupt: function (trackEntry) { },
					end: function (trackEntry) {},
					complete: (
						(trackEntry: TrackEntry) => { 
							// Loop through all the animations
							if (trackEntry.next == null) {
								this.animationState.setAnimation(0, animation, true);
								for (let i = 1; i < animations.length; i++) {
									lastTrackEntry = this.animationState.addAnimation(0, animations[i], false, 0);
								}
							}
						}
					).bind(this),
					dispose: function (trackEntry) {},
					event: function (entry: TrackEntry, event: Event): void {}
				});
			}
		}

		private getDefaultBoundingBox() {
			let animations = this.skeleton.data.animations;

			let offsetAvg = new spine.Vector2();
			let sizeAvg = new spine.Vector2();
			let num_processed = 0;
			for (let i = 0, n = animations.length; i < n; i++) {
				let animationName = animations[i].name;
				// if (!animationName.toLocaleLowerCase().includes("default")) {
				// 	continue;
				// }
				let animation = this.skeleton.data.findAnimation(animationName);
				this.animationState.clearTracks();
				this.skeleton.setToSetupPose()
				this.animationState.setAnimationWith(0, animation, true);

				let savedX = this.skeleton.x;
				let savedY = this.skeleton.y;
				this.skeleton.x = 0;
				this.skeleton.y = 0;
				this.skeleton.scaleX = Math.abs(this.config.scaleX);
				this.skeleton.scaleY = Math.abs(this.config.scaleY);
				this.animationState.update(0);
				this.animationState.apply(this.skeleton);
				this.skeleton.updateWorldTransform();
				let offset = new spine.Vector2();
				let size = new spine.Vector2();
				this.skeleton.getBounds(offset, size);

				this.skeleton.x = savedX;
				this.skeleton.y = savedY;

				if (Number.isFinite(offset.x) && Number.isFinite(offset.y) && Number.isFinite(size.x) && Number.isFinite(size.y)) {
					num_processed += 1;
					offsetAvg.x += offset.x;
					offsetAvg.y += offset.y;
					sizeAvg.x += size.x;
					sizeAvg.y += size.y;
				}
			}

			offsetAvg.x /= num_processed;
			offsetAvg.y /= num_processed;
			sizeAvg.x /= num_processed;
			sizeAvg.y /= num_processed;

			// this.average_width += sizeAvg.x;
			// this.average_height += sizeAvg.y;
			// this.average_count += 1;

			let ret = {
				x: offsetAvg.x,
				y: offsetAvg.y,
				width: sizeAvg.x,
				height: sizeAvg.y
			}

			// HACK: 
			// For enemies whose y bounding box is too far from the bottom
			// We add an offset so that it is rendered where we want it
			if (this.config.chibiId.includes("enemy_")) {
				if (ret.y < -15) {
					this.config.extraOffsetY = -ret.y;
				}
			}
			return ret;
		}

		private calculateAnimationViewport (animationName: string) {
			let animation = this.skeleton.data.findAnimation(animationName);
			this.animationState.clearTracks();
			this.skeleton.setToSetupPose()
			this.animationState.setAnimationWith(0, animation, true);

			let steps = 100;
			let stepTime = animation.duration > 0 ? animation.duration / steps : 0;
			let minX = 100000000;
			let maxX = -100000000;
			let minY = 100000000;
			let maxY = -100000000;
			let offset = new spine.Vector2();
			let size = new spine.Vector2();

			let savedX = this.skeleton.x;
			let savedY = this.skeleton.y;
			for (var i = 0; i < steps; i++) {
				this.animationState.update(stepTime);
				this.animationState.apply(this.skeleton);

				// TODO: Fix this hack
				// this.SetSkeletonMovementData(this.playerConfig.viewport);
				this.skeleton.x = 0;
				this.skeleton.y = 0;
				this.skeleton.scaleX = Math.abs(this.config.scaleX);
				this.skeleton.scaleY = Math.abs(this.config.scaleY);
				this.skeleton.updateWorldTransform();
				this.skeleton.getBounds(offset, size);

				minX = Math.min(offset.x, minX);
				maxX = Math.max(offset.x + size.x, maxX);
				minY = Math.min(offset.y, minY);
				maxY = Math.max(offset.y + size.y, maxY);
			}

			this.skeleton.x = savedX;
			this.skeleton.y = savedY;

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
 }