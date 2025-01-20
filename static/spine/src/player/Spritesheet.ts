import { Color } from "../core/Utils"

export class SpritesheetAnimationConfig  {
    public filepath: string
    public rows : number
    public cols :  number
    public scaleX : number
    public scaleY : number
    public width : number
    public height: number
    public frames: number 
    public fps: number
}

export class SpritesheetConfig {
    public animations: Map<string, SpritesheetAnimationConfig>
}

export class UVCoords {
    public U1 : number
    public U2 : number
    public V1 : number
    public V2 : number
}

export function readSpritesheetJsonConfig(json: string| any) : SpritesheetConfig {
    let data = new SpritesheetConfig();
    data.animations = new Map<string, SpritesheetAnimationConfig>();

    let root = typeof(json) === "string" ? JSON.parse(json) : json;
    try {
        for (let key in root.animations) {
            let animation = root.animations[key];
            let animationConfig = new SpritesheetAnimationConfig();
            animationConfig.filepath = animation.filepath;
            animationConfig.rows = animation.rows;
            animationConfig.cols = animation.cols;
            animationConfig.scaleX = animation.scaleX;
            animationConfig.scaleY = animation.scaleY;
            animationConfig.width = animation.width;
            animationConfig.height = animation.height;
            animationConfig.frames = animation.frames;
            animationConfig.fps = animation.fps;
            data.animations.set(key, animationConfig);
        }
    } catch (e) {
        console.log(e);
    }
    
    return data;
}

export class SpritesheetActor {
    public config: SpritesheetConfig
    public textures: Map<string, any>;
    private animationName: string;
    public animationConfig: SpritesheetAnimationConfig
    public highlightColor: Color = Color.WHITE;

    // animation state
    public timeScale = 1;
    private currentFrame = 0;
    private trackTime = 0;
    private durationPerFrame: number = 1.0;
    private totalAnimationDuration: number = 0.0;

    constructor(config: SpritesheetConfig, textures: Map<string, any>) {
        this.config = config;
        this.textures = textures;
    }

    public Update(delta: number) {
        delta *= this.timeScale;
        this.trackTime += delta;
        const trackTimeLocal = this.trackTime % this.totalAnimationDuration;
        this.currentFrame = Math.floor(trackTimeLocal / this.durationPerFrame);
    }

    public SetAnimation(animationName: string) {
        this.animationName = animationName;
        this.animationConfig = this.config.animations.get(animationName);
        this.durationPerFrame = 1.0 / this.animationConfig.fps;
        this.totalAnimationDuration = this.animationConfig.frames * this.durationPerFrame;
    }

    public GetTexture(): any {
        return this.textures.get(this.animationName);
    }

    public GetUVFromFrame() : UVCoords {
        let coords = new UVCoords();
        const frameIndex = this.currentFrame;

        const row = Math.floor(frameIndex / this.animationConfig.cols);
        const col = frameIndex % this.animationConfig.cols;
        const x = col / this.animationConfig.cols;
        const y = row / this.animationConfig.rows;
        const x1 = (col + 1) / this.animationConfig.cols;
        const y1 = (row + 1) / this.animationConfig.rows;

        coords.U1 = x;
        coords.V1 = y1;
        coords.U2 = x1;
        coords.V2 = y;
        return coords;
    }

}