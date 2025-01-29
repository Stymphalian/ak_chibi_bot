import { Vector2 } from "../core/Utils";
import { Camera } from "../webgl/Camera";
import { Vector3 } from "../webgl/Vector3";
import { BoundingBox } from "./Utils";

export class SpeechBubble {
    nameTagSize: any = null;

    public Reset() {
        this.nameTagSize = null;
    }

    public NameTag(
        viewport: BoundingBox,
        camera: Camera,
        ctx: CanvasRenderingContext2D,
        text: string,
        xpx: number, ypx: number, zpx: number = 0) {
        let tt = camera.worldToScreen(new Vector3(xpx, ypx, zpx));
        let textpos = new Vector2(tt.x, tt.y);
        textpos.y = viewport.height - textpos.y;
        xpx = textpos.x;
        ypx = textpos.y;

        ctx.save();
        ctx.translate(xpx, ypx);

        // Measure how much space is required for the speech bubble
        if (this.nameTagSize == null) {
            let data = ctx.measureText(text);
            let h = data.actualBoundingBoxAscent - data.actualBoundingBoxDescent;
            let w = data.width;
            this.nameTagSize = {
                width: w,
                height: h
            }
        }
        let width = this.nameTagSize.width;
        let height = this.nameTagSize.height;

        this._drawSpeechBubble(ctx, width, height);

        // Draw the text
        ctx.fillStyle = "white";
        ctx.fillText(text, -width / 2, 0);

        ctx.restore();
    }

    public ChatText(
        viewport: BoundingBox,
        camera: Camera,
        ctx: CanvasRenderingContext2D,
        texts: string[],
        xpx: number, ypx: number, zpx: number = 0) {
        let tt = camera.worldToScreen(new Vector3(xpx, ypx, zpx));
        let textpos = new Vector2(tt.x, tt.y);
        textpos.y = viewport.height - textpos.y;
        xpx = textpos.x;
        ypx = textpos.y;

        ctx.save();
        ctx.translate(xpx, ypx);

        // Measure how much space is required for the speech bubble
        let width = 0;
        let height = 0;
        let heights = [];
        for (let i = 0; i < texts.length; i++) {
            let data = ctx.measureText(texts[i]);
            let h = data.actualBoundingBoxAscent - data.actualBoundingBoxDescent;
            let w = data.width;
            width = Math.max(width, w);
            height += h;
            heights.push(h);
        }

        this._drawSpeechBubble(ctx, width, height);

        // Draw the text
        ctx.fillStyle = "white";
        let y = -height;
        for (let i = 0; i < texts.length; i++) {
            y += heights[i];
            let text = texts[i];
            ctx.fillText(text, -width / 2, y);
        }

        ctx.restore();
    }

    private _drawSpeechBubble(ctx: CanvasRenderingContext2D, width: number, height: number) {
        // Draw a speech bubble box
        ctx.beginPath();
        let pad = 5;
        ctx.fillStyle = "black";
        ctx.strokeStyle = "#333";
        ctx.lineWidth = 5;
        ctx.roundRect(-width / 2 - pad, pad, width + 2 * pad, -height - 2 * pad, 3);
        ctx.stroke();
        ctx.fill();

        // Draw a small triangle below the rect
        ctx.beginPath();
        ctx.fillStyle = "black";
        ctx.moveTo(-5, pad);
        ctx.lineTo(0, pad + 5);
        ctx.lineTo(5, pad);
        ctx.closePath();
        ctx.fill();

        // We want the border to blend in with the box and triangle
        ctx.beginPath();
        ctx.strokeStyle = "#333";
        ctx.lineWidth = 2;
        ctx.moveTo(-5, pad);
        ctx.lineTo(0, pad + 5);
        ctx.lineTo(5, pad);
        ctx.stroke();
        ctx.closePath();
    }
};