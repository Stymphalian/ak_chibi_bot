import { Color, Vector2 } from "../core/Utils";
import { SceneRenderer } from "../webgl/SceneRenderer";
import { Vector3 } from "../webgl/Vector3";
import { Actor } from "./Actor";
import { BoundingBox, SpinePlayer } from "./Player";


export class ActionName {
    static PLAY_ANIMATION = "PLAY_ANIMATION";
    static WANDER = "WANDER";
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

export function ParseActionNameToAction(actionName: string, actionData: any): ActorAction {
    switch (actionName) {
        case ActionName.PLAY_ANIMATION:
            return new PlayAnimationAction(actionData);
        case ActionName.WANDER:
            return new WanderAction(actionData);
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

function setActorYPositionByAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
    const startPosYScaled = actor.config.startPosY * viewport.height;
    actor.setPositionY(startPosYScaled);
    let isSitting = animation.toLowerCase().includes("sit")

    if (actor.canvasBB.y < 0) {
        // There is some sprite above the axis
        if (actor.IsEnemySprite()) {
            actor.setPositionY(startPosYScaled - actor.canvasBB.y);
        } else if (isSitting) {
            actor.setPositionY(startPosYScaled - actor.canvasBB.y);
        }
    } else if (actor.canvasBB.y > 0) {
        // There is some sprite below the axis
        if (actor.IsEnemySprite()) {
            actor.setPositionY(startPosYScaled - actor.canvasBB.y);
        }
    }
}

export class PlayAnimationAction implements ActorAction {
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
        return new Vector2(rand * viewport.width - half, currentPos.y);
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

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void { }

    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        if (!this.currentAnimation.includes("Move")) {
            actor.setVelocity(0, 0, 0);
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

export class WanderAction implements ActorAction {
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
        return new Vector2(Math.random() * viewport.width - half, currentPos.y);
    }

    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        // TODO: Figure out what this for walking, wander, pace-around actions
        // I know for playAnimation it is used to reset from a sit position
        // actor.setPositionY(actor.config.startPosY * viewport.height);
        setActorYPositionByAnimation(actor, animation, viewport)
    }

    GetAnimations(): string[] {
        return [this.actionData["wander_animation"]];
    }

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void { }
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
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


export class WalkToAction implements ActorAction {
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

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void { }
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
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
                target.x * viewport.width - (viewport.width / 2),
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
            actor.setPosition(this.endPosition.x, this.endPosition.y, 0);
            this.startPosition = actor.getPosition();
            actor.setVelocity(0, 0, 0);
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

    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void { }
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
            let z1 = Math.random() * 40;
            let z2 = Math.random() * 40;
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
            actor.setVelocity(0, 0, 0);
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

    getRandomPosition(currentPos: Vector3, viewport: BoundingBox): Vector3 {
        let half = viewport.width / 2;
        let rand = Math.random();
        return new Vector3(rand * viewport.width - half, currentPos.y, currentPos.z);
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
                this.noTargetRandomPosition = this.getRandomPosition(
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
            this.endPosition.x, 0,
            this.endPosition.x, viewport.height / 2.0,
            Color.PURPLE
        );
        renderer.circle(true, this.endPosition.x, viewport.height / 2.0, 5, Color.PURPLE);
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
            let step = actor.getMovementSpeed().scale(deltaSecs);
            // Scale the distance of the next velocity so that we don't 
            // overshoot the target position. Otherwise we can get into 
            // cases where the actor has janky oscillating movement as it
            // tries to reach endPosition
            let stepDist = this.endPosition.subtract(actorPosition);
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
    }
}