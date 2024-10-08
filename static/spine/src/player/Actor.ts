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

import { AnimationStateListener, AnimationState, TrackEntry } from "../core/AnimationState"
import { AnimationStateData } from "../core/AnimationStateData"
import { Skeleton } from "../core/Skeleton"
import { Vector2, TimeKeeper, Map } from "../core/Utils"
import { Vector3 } from "../webgl/Vector3"
import { ActorAction, ParseActionNameToAction } from "./Action"
import { SpinePlayer, BoundingBox } from "./Player"
import { Event } from "../core/Event"

// // module spine {

    export interface ActorUpdateConfig {
        start_pos: Vector2
        dest_pos: Vector2
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

		configScaleX: number
		configScaleY: number
		scaleX: number
		scaleY: number
		maxSizePx: number
		startPosX : number
		startPosY : number
		// Specify extra offset in Pixel for X and Y to fit the chibi on the screen
		extraOffsetX: number,
		extraOffsetY: number,

		defaultMovementSpeedPxX: number,
		defaultMovementSpeedPxY: number,
		defaultMovementSpeedPxZ: number,
		movementSpeedPxX: number,
		movementSpeedPxY: number,
		movementSpeedPxZ: number,

		// Unused?
		defaultScaleX?: number
		defaultScaleY?: number

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
		success?: (widget: SpinePlayer, actor: Actor) => void

		/* Optional: callback when the widget could not be loaded. */
		error?: (widget: SpinePlayer, actor: Actor, msg: string) => void

		/** Optional: Callbacks for the animations */
		animation_listener?: AnimationStateListener

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

		// Velocity is in world coordinates (pixels/second)
		// TODO: Abstract out this position/movement/speed data into a seperate class
		// TODO: Create our own vector3 class
		private movementSpeed: Vector3 = new Vector3();
		// Position is in world coordinates (with dimensions same as viewport)
		//   x-axis is 0 in center of the screen
		//   and y-axis is 0 at the bottom of the screen.
		// We use these coordinates compared to top-left corner.
		private position: Vector3 = new Vector3();
		public scale: Vector2 = new Vector2(1,1);
		private velocity: Vector3 = new Vector3(0, 0, 0);
		public startPosition: Vector2 = null;
		public currentAction: ActorAction = null;

		// Actor loading retry logic
		public load_attempts: number = 0;
		public max_load_attempts: number = 10;
		public load_perma_failed: boolean = false;
		// Record the timestamp when the actor was first loaded. We use this
		// to keep a stable sort order when rendering actors. Actors created
		// first should be rendered first, newer actors are rendered on top.
		public loadedWhen: number = new Date().getTime();

		constructor(config: SpineActorConfig, viewport: BoundingBox) {
			this.loadedWhen = new Date().getTime();
			this.viewport = viewport;
			this.ResetWithConfig(config);
			let x = Math.random()*viewport.width - (viewport.width/2)
			let z = 0;
			this.position = new Vector3(x, 0, z);

			if (config.startPosX || config.startPosY) {
				this.position = new Vector3(
					(config.startPosX* viewport.width) - (viewport.width/2),
					config.startPosY * viewport.height,
					z
				);
			}
		}

		public getPosition(): Vector2 {
			return new Vector2(
				this.position.x,
				this.position.y
			);
		}
		public getPosition3(): Vector3 {
			return new Vector3(
				this.position.x,
				this.position.y,
				this.position.z
			);
		}
		public getPositionX(): number {
			return this.position.x;
		}
		public getPositionY(): number {
			return this.position.y;
		}
		public getPositionZ(): number {
			return this.position.z;
		}
		public setPositionX(x: number) {
			this.position.x = x;
		}
		public setPositionY(y: number) {
			this.position.y = y;
		}
		public setPositionZ(z: number) {
			this.position.z = z;
		}
		public setPosition(x: number, y: number, z ?:number) {
			this.position.x = x;
			this.position.y = y;
			this.position.z = z !== undefined ? z : 0;
		}

		public getMovmentSpeed(): Vector2 {
			return new Vector2(
				this.movementSpeed.x,
				this.movementSpeed.y
			);
		}
		public getMovmentSpeed3(): Vector3 {
			return new Vector3(
				this.movementSpeed.x,
				this.movementSpeed.y,
				this.movementSpeed.z
			);
		}
		public getMovementSpeedX(): number {
			return this.movementSpeed.x;
		}
		public getMovementSpeedY(): number {
			return this.movementSpeed.y;
		}
		public getMovementSpeedZ(): number {
			return this.movementSpeed.z;
		}
		public setMovementSpeedX(x: number) {
			this.movementSpeed.x = x;
		}
		public setMovementSpeedY(y: number) {
			this.movementSpeed.y = y;
		}
		public setMovementSpeedZ(z: number) {
			this.movementSpeed.z = z;
		}
		public setMovementSpeed(x: number, y: number, z ?: number) {			
			this.movementSpeed.x = x;
			this.movementSpeed.y = y;
			this.movementSpeed.z = z !== undefined ? z : this.config.defaultMovementSpeedPxZ;
		}

		public getVelocity(): Vector2 {
			return new Vector2(
				this.velocity.x,
				this.velocity.y
			)
		}
		public getVelocity3(): Vector3 {
			return new Vector3(
				this.velocity.x,
				this.velocity.y,
				this.velocity.z
			)
		}
		public setVelocity(x?: number, y ?: number, z ?: number) {
			this.velocity = new Vector3(
				x != undefined ? x : this.velocity.x,
				y != undefined ? y : this.velocity.y,
				z != undefined ? z : this.velocity.z
			);
		}

		public InitAnimations() {
			this.initAnimationsInternal(this.currentAction.GetAnimations());
		}

		public GetAnimations(): string[] {
			return this.currentAction.GetAnimations()
		}

		public ResetWithConfig(config: SpineActorConfig) {
			this.load_attempts = 0;
			this.load_perma_failed = false;

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

			// Update movement speed from config
			if (config.movementSpeedPxX !== null 
				&& config.movementSpeedPxY !== null
				&& config.movementSpeedPxZ !== null) {
				this.movementSpeed = new Vector3(
					config.movementSpeedPxX, 
					config.movementSpeedPxY,
					config.movementSpeedPxZ,
				);
			} else {
				this.movementSpeed = new Vector3(
					config.defaultMovementSpeedPxX + Math.random()*config.defaultMovementSpeedPxX/2,
					config.defaultMovementSpeedPxY,
					config.defaultMovementSpeedPxZ,
				);
			}
			
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
			this.position.z += this.velocity.z;

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
			let maxSize = Math.sqrt(width*width + height*height);
			// if (height > maxSize) {
			// 	maxSize = height;
			// }
			if (maxSize > this.config.maxSizePx && this.config.chibiId.includes("enemy_")) {
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

			let offsetAvg = new Vector2();
			let sizeAvg = new Vector2();
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
				let offset = new Vector2();
				let size = new Vector2();
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
			let offset = new Vector2();
			let size = new Vector2();

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
//  }
