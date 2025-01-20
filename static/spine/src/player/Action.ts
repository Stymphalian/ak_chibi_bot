import { Color, Vector2 } from "../core/Utils";
import { SceneRenderer } from "../webgl/SceneRenderer";
import { Vector3 } from "../webgl/Vector3";
import { Actor } from "./Actor";
import { ExperimentFlags } from "./Flags";
import { BoundingBox, SpinePlayer } from "./Player";

export class ActionName {
    static PLAY_ANIMATION = "PLAY_ANIMATION";
    static WANDER = "WANDER";
    static WALK = "WALK";
    static WALK_TO = "WALK_TO";
    static PACE_AROUND = "PACE_AROUND";
    static FOLLOW = "FOLLOW";
}

export interface ActorAction {
    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox): void;
    GetAnimations(): string[]
    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void
    UpdatePhysics(
        actor: Actor,
        deltaSecs: number,
        viewport: BoundingBox,
        player: SpinePlayer): void
}

export function ParseActionNameToAction(actionName: string, actionData: any, flags: Map<string, any>): ActorAction {
    switch (actionName) {
        case ActionName.PLAY_ANIMATION:
            return new PlayAnimationAction(actionData);
        case ActionName.WANDER:
            if (flags.get(ExperimentFlags.WANDER_WITH_STOP)) {
                return new WanderAction(actionData);    
            } else {
                actionData["walk_animation"] = actionData["wander_animation"];
                return new WalkAction(actionData);
            }
        case ActionName.WALK:
            return new WalkAction(actionData);
        case ActionName.WALK_TO:
            return new WalkToAction(actionData);
        case ActionName.PACE_AROUND:
            return new PaceAroundAction(actionData);
        case ActionName.FOLLOW:
            return new FollowAction(actionData);
        default:
            console.log("Unknown action name ", actionName);
            return null;
    }
}


const DIST_TOLERANCE = 0.001;
const DEBUG_HEIGHT_1 = 0.4;
const DEBUG_HEIGHT_2 = 0.5;
const DEBUG_HEIGHT_3 = 0.6;
const DEBUG_HEIGHT_4 = 0.7;
const DEBUG_HEIGHT_5 = 0.8;
const DEBUG_HEIGHT_6 = 0.9;

// function getRandomPosition(currentPos: Vector3, viewport: BoundingBox): Vector3 {
//     let half = viewport.width / 2;
//     let rand = Math.random();
//     // NOTE: We do the rand*width - half because origin is in the center of the screen
//     // We want to ensure that the chibi stays within the bounds of the viewport
//     return new Vector3(rand * viewport.width - half, currentPos.y, currentPos.z);
// }
function getRandomPositionWithZ(currentPos: Vector3, viewport: BoundingBox): Vector3 {
    let half = viewport.width / 2;
    let rand = Math.random();
    let randZ = Math.random() * 10;
    // NOTE: We do the rand*width - half because origin is in the center of the screen
    // We want to ensure that the chibi stays within the bounds of the viewport
    return new Vector3(rand * viewport.width - half, currentPos.y, randZ);
}

function setActorYPositionByAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
    const startPosYScaled = actor.config.startPosY * viewport.height;
    actor.setPositionY(startPosYScaled);
    actor.setSkeletonPositionOffsetY(0);
    let isSitting = animation.toLowerCase().includes("sit")

    if (actor.canvasBB.y < 0) {
        // There is some sprite below the axis
        if (actor.IsEnemySprite()) {
            actor.setSkeletonPositionOffsetY(-actor.canvasBB.y);
        } else if (isSitting) {
            actor.setSkeletonPositionOffsetY(-actor.canvasBB.y);
        }
    } else if (actor.canvasBB.y > 0) {
        // There is some sprite above the axis
        if (actor.IsEnemySprite()) {
            actor.setSkeletonPositionOffsetY(-actor.canvasBB.y);
        }
    }
}

function updateVelocityFromDir(actor: Actor, endPosition: Vector3, dir: Vector3, deltaSecs: number) {
    let actorPosition = actor.getPosition3();
    let step = actor.getMovementSpeed().scale(deltaSecs);
    // Scale the distance of the next velocity so that we don't 
    // overshoot the target position. Otherwise we can get into 
    // cases where the actor has janky oscillating movement as it
    // tries to reach endPosition
    let stepDist = endPosition.subtract(actorPosition);
    if (step.x > Math.abs(stepDist.x)) {
        step.x = Math.abs(stepDist.x);
    }
    if (step.y > Math.abs(stepDist.y)) {
        step.y = Math.abs(stepDist.y);
    }
    if (step.z > Math.abs(stepDist.z)) {
        step.z = Math.abs(stepDist.z);
    }
    actor.setVelocity(dir.x * step.x, dir.y * step.y, dir.z * step.z);
}

export class PlayAnimationAction implements ActorAction {
    public actionData: any
    public startPosition: Vector3
    public endPosition: Vector3
    public currentAnimation: string

    constructor(actionData: any) {
        this.actionData = actionData
        this.currentAnimation = null;
        this.endPosition = null;
    }

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        setActorYPositionByAnimation(actor, animation, viewport)
        this.currentAnimation = animation;
        if (this.currentAnimation.includes("Move")) {
            this.startPosition = null;
            this.endPosition = null;
        }
    }

    GetAnimations(): string[] {
        return this.actionData["animations"];
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void {
    }

    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        if (!this.currentAnimation.includes("Move")) {
            actor.setVelocity(0, 0, 0);
        } else {
            if (this.endPosition == null) {
                this.startPosition = actor.getPosition3();
                this.endPosition = getRandomPositionWithZ(actor.getPosition3(), viewport);
            }

            let dir = this.endPosition.subtract(actor.getPosition3());
            if (dir.length() < 0.0001) {
                // We have reached the target position. Find a new destination
                actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
                this.startPosition = actor.getPosition3();
                this.endPosition = getRandomPositionWithZ(actor.getPosition3(), viewport);
            }
            dir.normalize();
            updateVelocityFromDir(actor, this.endPosition, dir, deltaSecs);
        }
    }
}

export class WalkAction implements ActorAction {
    public actionData: any
    public startPosition: Vector3 = null;
    public endPosition: Vector3 = null;

    constructor(actionData: any) {
        this.actionData = actionData
        this.startPosition = null;
        this.endPosition = null;
    }

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        // TODO: Figure out what this for walking, wander, pace-around actions
        // I know for playAnimation it is used to reset from a sit position
        // actor.setPositionY(actor.config.startPosY * viewport.height);
        setActorYPositionByAnimation(actor, animation, viewport)
    }

    GetAnimations(): string[] {
        return [this.actionData["walk_animation"]];
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void {
        if (this.endPosition) {
            renderer.line(
                this.endPosition.x,
                this.endPosition.y,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_1,
                Color.OXFORD_BLUE,
            )
            renderer.circle(
                false,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_1,
                5,
                Color.OXFORD_BLUE
            );
        }
    }
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        if (this.endPosition == null) {
            this.startPosition = actor.getPosition3();
            this.endPosition = getRandomPositionWithZ(actor.getPosition3(), viewport);
        }

        let dir = this.endPosition.subtract(actor.getPosition3());
        if (dir.length() < DIST_TOLERANCE) {
            // We have reached the target position. Find a new destination
            actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
            this.startPosition = actor.getPosition3();
            this.endPosition = getRandomPositionWithZ(actor.getPosition3(), viewport);
        }
        dir.normalize();
        updateVelocityFromDir(actor, this.endPosition, dir, deltaSecs);
    }
}

export class WanderAction implements ActorAction {
    static WANDER_WALK = "wander_walk";
    static WANDER_IDLE = "wander_idle";
    static WANDER_IDLE_MIN_TIME_SEC = 10;
    static WANDER_IDLE_WAIT_TIME_SEC = 20;
    
    public actionData: any
    public startPosition: Vector3 = null;
    public endPosition: Vector3 = null;
    public state: string = null;
    public idleWaitTimeTotalSecs: number = 0;
    public idleWaitTimeSecs: number = 0;

    constructor(actionData: any) {
        this.actionData = actionData
        this.startPosition = null;
        this.endPosition = null;
        this.state = WanderAction.WANDER_WALK;
        this.idleWaitTimeTotalSecs = 0;
        this.idleWaitTimeSecs = 0;
    }

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        // TODO: Figure out what this is for walking, wander, pace-around actions
        // I know for playAnimation it is used to reset from a sit position
        // actor.setPositionY(actor.config.startPosY * viewport.height);
        setActorYPositionByAnimation(actor, animation, viewport)
    }

    GetAnimations(): string[] {
        if (this.state == WanderAction.WANDER_IDLE) {
            return [this.actionData["wander_animation_idle"]];
        } else {
            return [this.actionData["wander_animation"]];
        }
        // return [this.actionData["wander_animation"]];
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void {
        if (this.endPosition) {
            renderer.line(
                this.endPosition.x,
                this.endPosition.y,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_1,
                Color.OXFORD_BLUE,
            )
            renderer.circle(
                false,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_1,
                (this.idleWaitTimeSecs / this.idleWaitTimeTotalSecs) * 30,
                Color.OXFORD_BLUE
            );
        }
    }
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        if (this.state == WanderAction.WANDER_WALK) {
            if (this.endPosition == null) {
                this.startPosition = actor.getPosition3();
                this.endPosition = getRandomPositionWithZ(actor.getPosition3(), viewport);
            }
    
            let dir = this.endPosition.subtract(actor.getPosition3());
            if (dir.length() < DIST_TOLERANCE) {
                // We have reached the target position. Find a new destination
                actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
                actor.setVelocity(0, 0, 0);
                this.state = WanderAction.WANDER_IDLE;
                this.idleWaitTimeTotalSecs = WanderAction.WANDER_IDLE_MIN_TIME_SEC + Math.random() * WanderAction.WANDER_IDLE_WAIT_TIME_SEC;
                this.idleWaitTimeSecs = this.idleWaitTimeTotalSecs;
                actor.InitAnimationState();
            } else {
                dir.normalize();
                updateVelocityFromDir(actor, this.endPosition, dir, deltaSecs);
            }
            
        } else if (this.state == WanderAction.WANDER_IDLE) {
            this.idleWaitTimeSecs -= deltaSecs;
            if (this.idleWaitTimeSecs < 0) {
                this.state = WanderAction.WANDER_WALK;
                this.startPosition = actor.getPosition3();
                this.endPosition = getRandomPositionWithZ(actor.getPosition3(), viewport);
                actor.InitAnimationState();
            }
        }
    }
}

export class WalkToAction implements ActorAction {
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

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        setActorYPositionByAnimation(actor, animation, viewport);
    }

    GetAnimations(): string[] {
        if (this.reachedDestination) {
            return [this.actionData["walk_to_final_animation"]];
        } else {
            return [this.actionData["walk_to_animation"]];
        }
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void {
        if (this.endPosition) {
            renderer.line(
                this.endPosition.x,
                this.endPosition.y,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_2,
                Color.PENN_BLUE,
            )
            renderer.circle(
                false,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_2,
                10,
                Color.PENN_BLUE);
        }
    }

    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        if (this.reachedDestination) {
            return;
        }

        if (this.endPosition == null) {
            this.startPosition = actor.getPosition3();
            let targetX = this.actionData["target_pos"]["x"];
            let targetY = this.actionData["target_pos"]["y"];
            this.endPosition = new Vector3(
                targetX * viewport.width - (viewport.width / 2),
                targetY,
                0,
            );

            const dist = this.startPosition.subtract(this.endPosition).length();
            if (dist < DIST_TOLERANCE) {
                this.reachedDestination = true;
                actor.InitAnimationState();
                return;
            }
        }

        let dir = this.endPosition.subtract(actor.getPosition3()).normalize();
        let dist = this.endPosition.subtract(actor.getPosition3()).length();
        if (dist < DIST_TOLERANCE) {
            // We have reached the target destination
            actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
            actor.setVelocity(0, 0, 0);
            this.reachedDestination = true;
            actor.InitAnimationState();
        } else {
            updateVelocityFromDir(actor, this.endPosition, dir, deltaSecs);
        }
    }
}

export class PaceAroundAction implements ActorAction {
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

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        setActorYPositionByAnimation(actor, animation, viewport);
    }

    GetAnimations(): string[] {
        return [this.actionData["pace_around_animation"]];
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void {
        if (this.endPosition) {
            renderer.line(
                this.endPosition.x,
                this.endPosition.y,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_3,
                Color.ROYAL_BLUE,
            )
            renderer.circle(
                false,
                this.endPosition.x,
                this.endPosition.y + viewport.height * DEBUG_HEIGHT_3,
                15,
                Color.ROYAL_BLUE);
        }
        if (this.startPosition) {
            renderer.line(
                this.startPosition.x,
                this.startPosition.y,
                this.startPosition.x,
                this.startPosition.y + viewport.height * DEBUG_HEIGHT_3,
                Color.ROYAL_BLUE,
            )
            renderer.circle(
                true,
                this.startPosition.x,
                this.startPosition.y + viewport.height * DEBUG_HEIGHT_3,
                15,
                Color.ROYAL_BLUE);
        }
    }
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
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
            let z1 = Math.random() * 10;
            let z2 = Math.random() * 10;
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
                start.x * viewport.width - (viewport.width / 2),
                start.y * viewport.height,
                start.z
            )
            this.endPosition = new Vector3(
                target.x * viewport.width - (viewport.width / 2),
                target.y * viewport.height,
                target.z
            );
            let temp = this.endPosition.copy();
            if (temp.sub(actor.getPosition3()).length() < DIST_TOLERANCE) {
                this.reachedDestination = true;
            }
        }
        // TODO: Figure out a better way of handling if we have reached an end position.
        let dir = this.endPosition.copy().sub(actor.getPosition3()).normalize();
        let dist_to = this.endPosition.copy().sub(actor.getPosition3()).length();
        if (dist_to < DIST_TOLERANCE || this.reachedDestination) {
            // We have reached the target destination
            actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
            actor.setVelocity(0, 0, 0);
            this.reachedDestination = true;
        } else {
            updateVelocityFromDir(actor, this.endPosition, dir, deltaSecs);
        }
    }
}


export class FollowAction implements ActorAction {
    public actionData: any
    public endPosition: Vector3 = null;
    public noTargetRandomPosition: Vector3 = null;
    public reachedDestination: boolean;

    constructor(actionData: any) {
        this.actionData = actionData;
        this.endPosition = null;
        this.noTargetRandomPosition = null;
        this.reachedDestination = false;
    }

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        setActorYPositionByAnimation(actor, animation, viewport);
    }

    GetAnimations(): string[] {
        if (this.reachedDestination) {
            return [this.actionData["action_follow_idle_animation"]];
        } else {
            return [this.actionData["action_follow_walk_animation"]];
        }
    }

    getTargetPosition(actor: Actor, player: SpinePlayer, viewport: BoundingBox): Vector3 {
        let targetActor = player.getActor(this.actionData["action_follow_target"]);
        if (targetActor == null) {
            // For cases if the target actor doesn't exist (due to GC)
            // we have the actor just walk to random positions.
            if (this.noTargetRandomPosition == null) {
                this.noTargetRandomPosition = getRandomPositionWithZ(
                    actor.getPosition3(), viewport);
            }
            return this.noTargetRandomPosition;
        } else {
            // The target position should be the backside of the 
            // targetActor's bounding box. Plus a buffer of the current
            // actor's frontal bounding box.
            let targetPosX = targetActor.GetBoundingBack();
            let targetPosY = targetActor.GetBoundingBottom();
            let targetPosZ = targetActor.getPositionZ();
            let actorOffsetX = Math.abs(actor.GetBoundingFrontOffset());
            if (targetActor.isFacingRight()) {
                targetPosX -= actorOffsetX;
            } else {
                targetPosX += actorOffsetX;
            }
            return new Vector3(targetPosX, targetPosY, targetPosZ);
        }
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void {
        renderer.line(
            this.endPosition.x,
            this.endPosition.y,
            this.endPosition.x,
            this.endPosition.y + viewport.height * DEBUG_HEIGHT_4,
            Color.PURPLE
        );
        renderer.circle(
            false,
            this.endPosition.x,
            this.endPosition.y + viewport.height * DEBUG_HEIGHT_4,
            20,
            Color.PURPLE
        );
    }

    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        if (this.reachedDestination) {
            let targetActor = player.getActor(this.actionData["action_follow_target"]);
            if (targetActor != null && targetActor.isStationary()) {
                // Target is still stationary.
                return;
            } else {
                // Target is moving again, we need to start moving.
                this.reachedDestination = false;
                actor.InitAnimationState();
            }
        }
        let targetPosition = this.getTargetPosition(actor, player, viewport);
        let actorPosition = actor.getPosition3();
        this.endPosition = targetPosition;

        let dir = this.endPosition.subtract(actorPosition).normalize();
        let dist_to = this.endPosition.distance(actorPosition);
        if (dist_to < 0.001) {
            actor.setPosition(this.endPosition.x, this.endPosition.y, this.endPosition.z);
            actor.setVelocity(0, 0, 0);

            // Transition into idle animation, only if the target is no longer moving
            let targetActor = player.getActor(this.actionData["action_follow_target"]);
            if (targetActor != null) {
                if (targetActor.isStationary()) {
                    this.reachedDestination = true;
                    if (targetActor.isFacingRight()) {
                        actor.setFacingRight();
                    } else {
                        actor.setFacingLeft();
                    }
                    actor.InitAnimationState();
                }
            } else {
                this.noTargetRandomPosition = null;
            }
        } else {
            updateVelocityFromDir(actor, this.endPosition, dir, deltaSecs);
        }
    }
}