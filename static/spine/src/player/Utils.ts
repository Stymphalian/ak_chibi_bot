import { PerspectiveCamera, Camera } from "../webgl/Camera";
import { M00 } from "../webgl/Matrix4";
import { Vector3 } from "../webgl/Vector3";
import { Actor } from "./Actor";


export function getPerspectiveCameraZOffset(viewport: BoundingBox, near: number, far: number, fovY: number): number {
    let width = viewport.width;
    let height = viewport.height;
    let cam = new PerspectiveCamera(width, height);
    cam.near = near;
    cam.far = far;
    cam.fov = fovY;
    cam.zoom = 1;
    cam.position.x = 0.0;
    cam.position.y = 0.0;
    cam.position.z = 0.0;
    cam.update();

    let a = cam.projectionView.values[M00];
    let w = -a * (width / 2);
    return w;
}

export function configurePerspectiveCamera(cam: Camera, near: number, far: number, viewport: BoundingBox) {
    cam.near = near;
    cam.far = far;
    cam.zoom = 1;
    cam.position.x = 0;
    cam.position.y = viewport.height / 2;
    // TODO: Negative so that the view is not flipped?
    cam.position.z = -getPerspectiveCameraZOffset(
        viewport, cam.near, cam.far, cam.fov
    );
    cam.direction = new Vector3(0, 0, -1);
    cam.update();
}

export function updateCameraSettings(cam: Camera, actor: Actor, viewport: BoundingBox) {
    // let cam = this.sceneRenderer.camera;
    cam.position.x = 0;
    cam.position.y = viewport.height / 2;
    // TODO: Negative so that the view is not flipped?
    cam.position.z = -getPerspectiveCameraZOffset(viewport, cam.near, cam.far, cam.fov);
    // cam.position.z += actor.getPositionZ();
    cam.direction = new Vector3(0, 0, -1);
    // let origin = new Vector3(0,0,0);
    // let pos = new Vector3(0,0,cam.position.z)
    // let dir = origin.sub(pos).normalize();
    // cam.direction = dir;
    cam.update();
}

export function isContained(dom: HTMLElement, needle: HTMLElement): boolean {
    if (dom === needle) return true;
    let findRecursive = (dom: HTMLElement, needle: HTMLElement) => {
        for (var i = 0; i < dom.children.length; i++) {
            let child = dom.children[i] as HTMLElement;
            if (child === needle) return true;
            if (findRecursive(child, needle)) return true;
        }
        return false;
    };
    return findRecursive(dom, needle);
}

export function findWithId(dom: HTMLElement, id: string): HTMLElement[] {
    let found = new Array<HTMLElement>()
    let findRecursive = (dom: HTMLElement, id: string, found: HTMLElement[]) => {
        for (var i = 0; i < dom.children.length; i++) {
            let child = dom.children[i] as HTMLElement;
            if (child.id === id) found.push(child);
            findRecursive(child, id, found);
        }
    };
    findRecursive(dom, id, found);
    return found;
}

export function findWithClass(dom: HTMLElement, className: string): HTMLElement[] {
    let found = new Array<HTMLElement>()
    let findRecursive = (dom: HTMLElement, className: string, found: HTMLElement[]) => {
        for (var i = 0; i < dom.children.length; i++) {
            let child = dom.children[i] as HTMLElement;
            if (child.classList.contains(className)) found.push(child);
            findRecursive(child, className, found);
        }
    };
    findRecursive(dom, className, found);
    return found;
}

export function createElement(html: string): HTMLElement {
    let dom = document.createElement("div");
    dom.innerHTML = html;
    return dom.children[0] as HTMLElement;
}

export function removeClass(elements: HTMLCollection, clazz: string) {
    for (var i = 0; i < elements.length; i++) {
        elements[i].classList.remove(clazz);
    }
}

export function escapeHtml(str: string) {
    if (!str) return "";
    return str
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&#34;")
        .replace(/'/g, "&#39;");
}

export function isAlphanumeric(str: string): boolean {
    return /^[a-zA-Z0-9_-]{1,100}$/.test(str)
}
export function isAlphanumericWithSpace(str: string): boolean {
    return /^[a-zA-Z0-9_-\s]{1,100}$/.test(str)
}

export interface Viewport {
	x: number,
	y: number,
	width: number,
	height: number,
	padLeft: string | number
	padRight: string | number
	padTop: string | number
	padBottom: string | number
	debugRender: boolean
}

export interface BoundingBox {
	x: number,
	y: number,
	width: number,
	height: number,
}