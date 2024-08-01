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

		/* Optional: the name of the animation to be played. Default: first animation in the skeleton. */
		animations: string[]

		/* Optional: the default mix time used to switch between two animations. */
		defaultMix?: number

		/* Optional: the name of the skin to be set. Default: the default skin. */
		skin?: string

		/* Optional: list of skin names from which the user can choose. */
		skins?: string[]

		/* Optional: whether the skeleton uses premultiplied alpha. Default: true. */
		premultipliedAlpha: boolean

		/** Optional: Animation play speed. 0 to 2.0 */
		animationPlaySpeed: number,

		scaleX: number
		scaleY: number
		defaultScaleX?: number
		defaultScaleY?: number
		maxSizePx: number

		startPosX : number
		startPosY : number
		desiredPositionX?: number,
		desiredPositionY?: number,
		wandering: boolean,

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

		public animViewport: BoundingBox = null;
		public prevAnimViewport: BoundingBox = null;
		public defaultBB: BoundingBox = null;

		// Velcity in world coordinates space (units/second)
		public movementSpeed: spine.Vector2 = new spine.Vector2();
		public position: spine.Vector2 = new spine.Vector2();
		public scale: spine.Vector2 = new spine.Vector2(1,1);
		public startPosition: spine.Vector2 = null;
		public endPosition: spine.Vector2 = null;

		constructor(config: SpineActorConfig, viewport: BoundingBox) {
			this.ResetWithConfig(config);
			this.movementSpeed = new spine.Vector2(80 + Math.random()*40,0);  // ~100 px/ second
			let x = Math.random()*viewport.width - (viewport.width/2)
			this.position = new spine.Vector2(x, 0);

			if (config.startPosX || config.startPosY) {
				this.position = new spine.Vector2(
					(config.startPosX* viewport.width) - (viewport.width/2),
					config.startPosY * viewport.height
				);
			}
			// this.position = new spine.Vector2(0, 0);
		}

		setDestination(viewport: BoundingBox) {
			this.startPosition = new spine.Vector2(this.position.x, this.position.y);
			let half = viewport.width / 2;
			this.endPosition = new spine.Vector2(Math.random()*viewport.width - half, this.position.y);
			// this.startPosition = new spine.Vector2(0, 0);
			// this.endPosition = new spine.Vector2(0, 0);
		}

		setEndPosition(position: spine.Vector2) {
			this.endPosition = position;
		}

		loopPositions() {
			let temp = this.endPosition;
			this.endPosition = new spine.Vector2(this.startPosition.x, this.startPosition.y);
			this.startPosition = new spine.Vector2(temp.x, temp.y);
		}

		clearDestination() {
			this.startPosition = null;
			this.endPosition = null;
		}

		ResetWithConfig(config: SpineActorConfig) {
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

			// current facing should be kept the same, event as we change
			// the configuration/animation
			this.scale.x = Math.sign(this.scale.x) * config.scaleX;
			this.scale.y = Math.sign(this.scale.y) * config.scaleY;
		}

		UpdatePhysics(deltaSecs: number, viewport: BoundingBox) {
			if (this.endPosition != null) {
				let dir = new spine.Vector2(
					this.endPosition.x - this.position.x,
					this.endPosition.y - this.position.y
				).normalize();
				let velocity = new spine.Vector2(
					dir.x * this.movementSpeed.x * deltaSecs,
					dir.y * this.movementSpeed.y * deltaSecs
				)
				this.position.x += velocity.x;
				this.position.y += velocity.y;
	
				if (velocity.x > 0) {
					this.scale.x = this.config.scaleX;
				} else {
					this.scale.x = -this.config.scaleX;
				}

				if (Math.abs(this.endPosition.x - this.position.x) < 5) {
					this.position.x = this.endPosition.x;
					if (this.config.wandering) {
						// Find a new destination;
						this.setDestination(viewport);
						// this.loopPositions();
					} else {
						this.config.animations = ["Idle"];
					}
				}
			}
			this.setSkeletonMovementData(viewport);
		}

		getUsernameHeaderHeight() {
			// let avg_width = (51.620900749705015 * this.config.scaleX) / 0.15;
			// let avg_height = (60.99747806746469 * this.config.scaleY) / 0.15;
			let finalHeight = 0;
			if (this.config.chibiId.includes("enemy_")) {
				finalHeight = this.defaultBB.height;
			} else {
				// HACK: Sometimes the chibi bounding box is too big and
				// doesn't not reasonablly fit the chibi. In order to make
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

		setSkeletonMovementData(viewport: BoundingBox) {
			this.skeleton.x = this.position.x + this.config.extraOffsetX;
			this.skeleton.y = this.position.y + this.config.extraOffsetY;
			this.skeleton.scaleX = this.scale.x;
			this.skeleton.scaleY = this.scale.y;
		}
	}
 }
