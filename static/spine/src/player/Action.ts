import { Vector2 } from "../core/Utils";
import { Vector3 } from "../webgl/Vector3";
import { Actor } from "./Actor";
import { BoundingBox } from "./Player";

// // module spine {
    export class ActionName {
		static PLAY_ANIMATION = "PLAY_ANIMATION";
		static WANDER = "WANDER";
		static WALK_TO = "WALK_TO";
        static PACE_AROUND = "PACE_AROUND";
    }

    export interface ActorAction {
        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox): void;
        GetAnimations(): string[]
        UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox): void
    }

    export function ParseActionNameToAction(actionName: string, actionData: any) : ActorAction 
    {
        switch (actionName) {
            case ActionName.PLAY_ANIMATION:
                return new PlayAnimationAction(actionData);
            case ActionName.WANDER:
                return new WanderAction(actionData);
            case ActionName.WALK_TO:
                return new WalkToAction(actionData);
            case ActionName.PACE_AROUND:
                return new PaceAroundAction(actionData);
            default:
                console.log("Unknown action name ", actionName);
                return null;
        }
    }

    export class PlayAnimationAction implements ActorAction{
        public actionData: any
        public startPosition: Vector2
        public endPosition: Vector2
        public currentAnimation: string

        constructor(actionData: any) {
            this.actionData = actionData
            this.currentAnimation = null;
            this.endPosition = null;
        }

        getRandomPosition(currentPos: Vector2, viewport: BoundingBox): Vector2 {
            let half = viewport.width / 2;
            let rand = Math.random();
			return new Vector2(rand*viewport.width - half, currentPos.y);
        }

        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox) {
            const startPosYScaled = actor.config.startPosY * viewport.height;
            actor.setPositionY(startPosYScaled - actor.canvasBB.y)
            this.currentAnimation = animation;
            if (this.currentAnimation.includes("Move")) {
                this.startPosition = null;
                this.endPosition = null;
            }
        }

        GetAnimations(): string[] {
            return this.actionData["animations"];
        }

        UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox) {
            if (!this.currentAnimation.includes("Move")) {
                actor.setVelocity(0,0,0);
            } else {
                if (this.endPosition == null) {
                    this.startPosition = actor.getPosition();
                    this.endPosition = this.getRandomPosition(actor.getPosition(), viewport);
                }
    
                let dir = this.endPosition.subtract(actor.getPosition());
                if (dir.length() < 5) {
                    // We have reached the target position. Find a new destination
                    actor.setPosition(this.endPosition.x, this.endPosition.y, 0);
                    this.startPosition = actor.getPosition();
                    this.endPosition = this.getRandomPosition(actor.getPosition(), viewport);
                }    
                dir.normalize();
                actor.setVelocity(
                    dir.x * actor.getMovementSpeedX() * deltaSecs,
                    dir.y * actor.getMovementSpeedY() * deltaSecs,
                    0
                );
            }
        }
    }

    export class WanderAction implements ActorAction{
        public actionData: any
        public startPosition: Vector2 = null;
        public endPosition: Vector2 = null;

        constructor(actionData: any) {
            this.actionData = actionData
            this.startPosition = null;
            this.endPosition = null;
        }

        getRandomPosition(currentPos: Vector2, viewport: BoundingBox): Vector2 {
            let half = viewport.width / 2;
			return new Vector2(Math.random()*viewport.width - half, currentPos.y);
        }

        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox) {
            // TODO: Figure out what this for walking, wander, pace-around actions
            // I know for playAnimation it is used to reset from a sit position
            // actor.setPositionY(actor.config.startPosY * viewport.height);
            const startPosYScaled = actor.config.startPosY * viewport.height;
            actor.setPositionY(startPosYScaled - actor.canvasBB.y)
        }

        GetAnimations(): string[] {
            return [this.actionData["wander_animation"]];
        }

        UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox){
            if (this.endPosition == null) {
                this.startPosition = actor.getPosition();
                this.endPosition = this.getRandomPosition(actor.getPosition(), viewport);
            }

            let dir = this.endPosition.subtract(actor.getPosition());
            if (dir.length() < 5) {
                // We have reached the target position. Find a new destination
                actor.setPosition(this.endPosition.x, this.endPosition.y, 0);
                this.startPosition = actor.getPosition();
                this.endPosition = this.getRandomPosition(actor.getPosition(), viewport);
            }    
            dir.normalize();
            actor.setVelocity(
                dir.x * actor.getMovementSpeedX() * deltaSecs,
                dir.y * actor.getMovementSpeedY() * deltaSecs,
                0
            );
        }
    }


    export class WalkToAction implements ActorAction{
        public actionData: any
        public startPosition: Vector2 = null;
        public endPosition: Vector2 = null;
        public startDir: Vector2 = null;
        public reachedDestination: boolean;

        constructor(actionData: any) {
            this.actionData = actionData;
            this.startPosition = null;
            this.endPosition = null;
            this.startDir = null;
            this.reachedDestination = false;
        }

        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox) {
            const startPosYScaled = actor.config.startPosY * viewport.height;
            actor.setPositionY(startPosYScaled - actor.canvasBB.y)
        }

        GetAnimations(): string[] {
            if (this.reachedDestination) {
                return [this.actionData["walk_to_final_animation"]];
            } else {
                return [this.actionData["walk_to_animation"]];
            }
        }

        UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox) {
            if (this.reachedDestination) {
                return;
            }

            if (this.endPosition == null) {
                this.startPosition = actor.getPosition();
                let target = new Vector2(
                    this.actionData["target_pos"]["x"],
                    this.actionData["target_pos"]["y"],
                );
                this.endPosition = new Vector2(
                    target.x * viewport.width - (viewport.width/2),
                    0,
                );
                this.startDir = this.endPosition.subtract(this.startPosition);
                if (this.startDir.length() < 10) {
                    this.reachedDestination = true;
                }
                this.startDir = this.startDir.normalize()
            }

            let dir = this.endPosition.subtract(actor.getPosition()).normalize();
            let angle = this.startDir.angle(dir);
            let reached = Math.abs(Math.PI - angle) < 0.001;
            if (reached || this.reachedDestination) {
                // We have reached the target destination
                actor.setPosition(this.endPosition.x, this.endPosition.y,0);
                this.startPosition = actor.getPosition();
                actor.setVelocity(0,0,0);
                this.reachedDestination = true;
                actor.InitAnimationState();
            } else {
                actor.setVelocity(
                    dir.x * actor.getMovementSpeedX() * deltaSecs,
                    dir.y * actor.getMovementSpeedY() * deltaSecs,
                    0
                );
            }
        }
    }

    export class PaceAroundAction implements ActorAction{
        public actionData: any
        public startPosition: Vector3 = null;
        public endPosition: Vector3 = null;
        public reachedDestination: boolean;

        constructor(actionData: any) {
            this.actionData = actionData;
            this.startPosition = null;
            this.endPosition = null;
            this.reachedDestination = false;
        }

        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox) {
            const startPosYScaled = actor.config.startPosY * viewport.height;
            actor.setPositionY(startPosYScaled - actor.canvasBB.y)
        }

        GetAnimations(): string[] {
            return [this.actionData["pace_around_animation"]];
        }

        UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox) {
            if (this.reachedDestination) {
                this.reachedDestination = false;

                // Swap the positions
                let tempPosition = this.startPosition;
                this.startPosition = this.endPosition;
                this.endPosition = tempPosition;
                return;
            }

            if (this.endPosition == null) {
                // TODO: Make the z range configurable
                let z1 = Math.random()*40;
                let z2 = Math.random()*40;
                let start = new Vector3(
                    this.actionData["pace_start_pos"]["x"],
                    this.actionData["pace_start_pos"]["y"],
                    z1
                )
                let target = new Vector3(
                    this.actionData["pace_end_pos"]["x"],
                    this.actionData["pace_end_pos"]["y"],
                    z2
                );
                this.startPosition = new Vector3(
                    start.x * viewport.width - (viewport.width/2),
                    start.y * viewport.height,
                    start.z
                )
                this.endPosition = new Vector3(
                    target.x * viewport.width - (viewport.width/2),
                    target.y * viewport.height,
                    target.z
                );
                let temp = this.endPosition.copy();
                if (temp.sub(actor.getPosition3()).length() < 10) {
                    this.reachedDestination = true;
                }
            }
            // TODO: Figure out a better way of handling if we have reached an end position.
            let dir = this.endPosition.copy().sub(actor.getPosition3()).normalize();
            let dist_to = this.endPosition.copy().sub(actor.getPosition3()).length();
            if (dist_to < 10 || this.reachedDestination) {
                // We have reached the target destination
                actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
                actor.setVelocity(0,0,0);
                this.reachedDestination = true;
            } else {
                actor.setVelocity(
                    dir.x * actor.getMovementSpeedX() * deltaSecs,
                    dir.y * actor.getMovementSpeedY() * deltaSecs,
                    dir.z * actor.getMovementSpeedZ() * deltaSecs
                )
            }
        }
    }
// }