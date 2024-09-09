module spine {
    export class ActionName {
		static PLAY_ANIMATION = "PLAY_ANIMATION";
		static WANDER = "WANDER";
		static WALK_TO = "WALK_TO";
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
			return new spine.Vector2(rand*viewport.width - half, currentPos.y);
        }

        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox) {
            const startPosYScaled = actor.config.startPosY * viewport.y;
            if (animation == "Sit") {
				actor.position.y = startPosYScaled + Math.abs(actor.animViewport.y)
			} else {
				actor.position.y = startPosYScaled;
			}
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
                actor.velocity.x = 0;
                actor.velocity.y = 0;
            } else {
                if (this.endPosition == null) {
                    this.startPosition = actor.position;
                    this.endPosition = this.getRandomPosition(actor.position, viewport);
                }
    
                let dir = this.endPosition.subtract(actor.position);
                if (dir.length() < 5) {
                    // We have reached the target position. Find a new destination
                    actor.position.x = this.endPosition.x;
                    actor.position.y = this.endPosition.y;
                    this.startPosition = actor.position;
                    this.endPosition = this.getRandomPosition(actor.position, viewport);
                }    
                dir.normalize();
                actor.velocity =  new Vector2(
                    dir.x * actor.movementSpeed.x * deltaSecs,
                    dir.y * actor.movementSpeed.y * deltaSecs
                )
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
			return new spine.Vector2(Math.random()*viewport.width - half, currentPos.y);
        }

        SetAnimation(actor:Actor, animation:string, viewport: BoundingBox) {
            actor.position.y = actor.config.startPosY * viewport.y;
        }

        GetAnimations(): string[] {
            return [this.actionData["wander_animation"]];
        }

        UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox){
            if (this.endPosition == null) {
                this.startPosition = actor.position;
                this.endPosition = this.getRandomPosition(actor.position, viewport);
            }

            let dir = this.endPosition.subtract(actor.position);
            if (dir.length() < 5) {
                // We have reached the target position. Find a new destination
                actor.position.x = this.endPosition.x;
                actor.position.y = this.endPosition.y;
                this.startPosition = actor.position;
                this.endPosition = this.getRandomPosition(actor.position, viewport);
            }    
            dir.normalize();
            actor.velocity =  new Vector2(
                dir.x * actor.movementSpeed.x * deltaSecs,
                dir.y * actor.movementSpeed.y * deltaSecs
            )
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
            actor.position.y = actor.config.startPosY * viewport.y;
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
                this.startPosition = actor.position;
                let target = new Vector2(
                    this.actionData["target_pos"]["x"],
                    this.actionData["target_pos"]["y"],
                );
                this.endPosition = new spine.Vector2(
                    target.x * viewport.width - (viewport.width/2),
                    0,
                );
                this.startDir = this.endPosition.subtract(this.startPosition).normalize();
            }

            let dir = this.endPosition.subtract(actor.position).normalize();
            let angle = this.startDir.angle(dir);
            let reached = Math.abs(Math.PI - angle) < 0.001;
            if (reached) {
                // We have reached the target destination
                actor.position.x = this.endPosition.x;
                actor.position.y = this.endPosition.y;
                this.startPosition = actor.position;
                actor.velocity.x = 0;
                actor.velocity.y = 0;
                this.reachedDestination = true;
                actor.InitAnimationState();
            } else {
                actor.velocity =  new Vector2(
                    dir.x * actor.movementSpeed.x * deltaSecs,
                    dir.y * actor.movementSpeed.y * deltaSecs
                );
            }
        }
    }
}