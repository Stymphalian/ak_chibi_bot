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
		jsonUrl: string

		/* the URL of the skeleton .skel file */
		skelUrl: string

		/* the URL of the skeleton .atlas file. Atlas page images are automatically resolved. */
		atlasUrl: string

		/* Raw data URIs, mapping from a path to base 64 encoded raw data. When the player
		   resolves a path of the `jsonUrl`, `skelUrl`, `atlasUrl`, or the image paths
		   referenced in the atlas, it will first look for that path in this array of
		   raw data URIs. This allows embedding of resources directly in HTML/JS. */
		rawDataURIs: Map<string>

		/* Optional: the name of the animation to be played. Default: first animation in the skeleton. */
		animation: string

		/* Optional: list of animation names from which the user can choose. */
		animations: string[]

		/* Optional: the default mix time used to switch between two animations. */
		defaultMix: number

		/* Optional: the name of the skin to be set. Default: the default skin. */
		skin: string

		/* Optional: list of skin names from which the user can choose. */
		skins: string[]

		/* Optional: whether the skeleton uses premultiplied alpha. Default: true. */
		premultipliedAlpha: boolean

		scaleX: number
		scaleY: number
		defaultScaleX: number
		defaultScaleY: number
		maxSizePx: number

		startPosX : number
		startPosY : number
		desiredPositionX: number,
		wandering: boolean,

		/* Optional: the position and size of the viewport in world coordinates of the skeleton. Default: the setup pose bounding box. */
		viewport: {
			x: number
			y: number
			minX: number
			minY: number
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

		// /* Optional: the background color used in fullscreen mode. Must be given in the format #rrggbbaa. Default: backgroundColor. */
		// fullScreenBackgroundColor: string

		// /* Optional: list of bone names that the user can control by dragging. */
		// controlBones: string[]

		/* Optional: callback when the widget and its assets have been successfully loaded. */
		success: (widget: SpinePlayer) => void

		/* Optional: callback when the widget could not be loaded. */
		error: (widget: SpinePlayer, msg: string) => void

		/** Optional: Callbacks for the animations */
		animation_listener: spine.AnimationStateListener

		userDisplayName: string,
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

		public animViewport: BoundingBox = null;
		public prevAnimViewport: BoundingBox = null;
		public defaultBB: BoundingBox = null;

		// Velcity in world coordinates space (units/second)
		public movementSpeed: spine.Vector2 = new spine.Vector2();
		public position: spine.Vector2 = new spine.Vector2();
		public scale: spine.Vector2 = new spine.Vector2(1,1);
		public startPosition: spine.Vector2 = null;
		public endPosition: spine.Vector2 = null;

		constructor(config: SpineActorConfig) {
			this.ResetWithConfig(config);
			this.movementSpeed = new spine.Vector2(80 + Math.random()*40,0);  // ~100 px/ second
			let x = Math.random()*1920 - (1920/2)
			this.position = new spine.Vector2(x, 0);

			if (config.startPosX || config.startPosY) {
				this.position = new spine.Vector2(
					config.startPosX,
					config.startPosY
				);
			}
			// this.position = new spine.Vector2(0, 0);
		}

		setDestination() {
			// TODO: Need to get the starting position from the canvas size
			// instead of hardcoding it here.
			this.startPosition = new spine.Vector2(this.position.x, this.position.y);
			let half = 1920 / 2;
			this.endPosition = new spine.Vector2(Math.random()*1920 - half, this.position.y);
			// this.startPosition = new spine.Vector2(0, 0);
			// this.endPosition = new spine.Vector2(0, 0);
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
			if (this.endPosition) {
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
						this.setDestination();
						// this.loopPositions();
					} else {
						this.config.animation = "Idle";
					}
				}
			}
			this.setSkeletonMovementData(viewport);
		}

		setSkeletonMovementData(viewport: BoundingBox) {
			this.skeleton.x = this.position.x;
			this.skeleton.y = this.position.y;
			this.skeleton.scaleX = this.scale.x;
			this.skeleton.scaleY = this.scale.y;
		}
	}
 }
