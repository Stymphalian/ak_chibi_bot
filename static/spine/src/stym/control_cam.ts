import { PerspectiveCamera, OrthoCamera } from "../webgl/Camera";
import { M00 } from "../webgl/Matrix4";
import { Vector3 } from "../webgl/Vector3";
import {Runtime} from "./runtime";
import "./stym.css";

export class ControlCam {
    private dom: HTMLElement;
    private parent: HTMLElement | null;
    private width: number;
    private height: number;
    private runtime: Runtime;

    constructor(parent: HTMLElement | string, width: number, height: number, runtime: Runtime) {
        if (typeof parent === "string") {
            this.parent = document.getElementById(parent);
        } else  {
            this.parent = parent;
        }
        this.parent?.appendChild(this.setupDom());
        this.width = width;
        this.height = height;
        this.runtime = runtime;
    }

    private getElementById(dom: HTMLElement, id: string): HTMLElement {
        return this.findWithId(this.dom, id)[0];
    }

    private findWithId(dom: HTMLElement, id: string): HTMLElement[] {
        let found = new Array<HTMLElement>()
        let findRecursive = (dom: HTMLElement, id: string, found: HTMLElement[]) => {
            for(var i = 0; i < dom.children.length; i++) {
                let child = dom.children[i] as HTMLElement;
                if (child.id === id) found.push(child);
                findRecursive(child, id, found);
            }
        };
        findRecursive(dom, id, found);
        return found;
    }

    private findWithClass(dom: HTMLElement, className: string): HTMLElement[] {
        let found = new Array<HTMLElement>()
        let findRecursive = (dom: HTMLElement, className: string, found: HTMLElement[]) => {
            for(var i = 0; i < dom.children.length; i++) {
                let child = dom.children[i] as HTMLElement;
                if (child.classList.contains(className)) found.push(child);
                findRecursive(child, className, found);
            }
        };
        findRecursive(dom, className, found);
        return found;
    }

    public createElement(html: string): HTMLElement {
        let dom = document.createElement("div");
        dom.innerHTML = html;
        return dom.children[0] as HTMLElement;
    }

    public setupDom() : HTMLElement {
        let dom = this.dom = this.createElement(/*html*/`
            <div class="cam_container">
                <label class="large_font">Camera Near/Far</label>
                <input class="cam_input watch" id="near-value" type="number" value="100" step="0.00001" size="10"></input>
                <input class="cam_input watch" id="far-value" type="number" value="600" step="100" size="10"></input>
                <br />
                <label class="large_font">Camera X,Y,Z</label>
                <input class="cam_input watch" id="cam-x" type="number" value="0" step="0.01" size="10"></input>
                <input class="cam_input watch" id="cam-y" type="number" value="0" step="1" size="10"></input>
                <input class="cam_input watch" id="cam-z" type="number" value="0" step="1" size="10"></input>
                <br />
                <label class="large_font">Camera FOV</label>
                <input class="cam_input watch" id="fov" type="number" value="45" step="5" size="10"></input>
                <div>
                    <label class="large_font" id="camera_type_label">Camera Type</label>
                    <select class="cam_input" id="camera_type">
                        <option value="orthographic">Orthographic</option>
                        <option value="perspective" selected="selected">Perspective</option>
                    </select>
                </div>

                <hr />
                <input class="cam_input watch" id="test_x" type="number" value="0" step="0.01" size="10"></input>
                <input class="cam_input watch" id="test_y" type="number" value="0" step="0.01" size="10"></input>
                <input class="cam_input watch" id="test_z" type="number" value="0" step="1" size="10"></input>
                <hr />
                <div>
                    <label class="large_font" id="transform_dir_label">Coordinate System</label>
                    <select class="cam_input watch" id="transform_dir">
                        <option value="world">World</option>
                        <option value="screen">Screen</option>
                    </select>
                    <br />

                    <label class="large_font" id="camera_choice_label">Perspective</label>
                    <select class="cam_input watch" id="camera_choice">
                        <option value="perspective" selected="selected">Perspective</option>
                        <option value="orthographic">Orthographic</option>
                    </select>
                    <br />

                    <label class="large_font" id="correct_cam_z">Correct Cam Z: </label>
                </div>
                <div>
                    <label class="large_font">World to Screen</label>
                    <p class="cam_input" id="world_to_screen"></p>
                </div>
                <div>
                    <label class="large_font">Screen To World</label>
                    <p class="cam_input" id="screen_to_world"></p>
                </div>
                <div>
                    <label class="large_font">Output</label>
                    <p class="cam_input" id="output1"></p>
                    <p class="cam_input" id="outpu2"></p>
                    <p class="cam_input" id="output3"></p>
                </div>
            </div>
        `)
        document.body.appendChild(this.dom);
        this.addListeners(this.dom);
        return dom;
    }


    private getCorrectCameraPerspectiveZ(near: number, far: number, fovY: number) {
        let cam = new PerspectiveCamera(this.width, this.height);
        cam.near = near;
        cam.far = far;
        cam.fov = fovY;
        cam.zoom = 1;
        cam.position.x = 0.0;
        cam.position.y = 0.0;
        cam.position.z = 0.0;
        cam.update();

        let a = cam.projectionView.values[M00];
        let w = -a * (this.width/2);
        return w;
    }

    private addListeners(dom: HTMLElement) {
        this.findWithClass(dom, "watch").forEach(input => {
            input.addEventListener("change", (event) => {
                try {
                    let width = this.width;
                    let height = this.height;
                    let cam = null;
                    if ((this.getElementById(dom, "camera_choice") as HTMLSelectElement).selectedIndex == 0) {
                        cam = new PerspectiveCamera(width, height);
                        cam.near = (this.getElementById(dom, "near-value") as HTMLInputElement).valueAsNumber;
                        cam.far = (this.getElementById(dom, "far-value") as HTMLInputElement).valueAsNumber;
                        cam.fov = (this.getElementById(dom, "fov") as HTMLInputElement).valueAsNumber;
                        cam.zoom = 1
                        cam.position.x = (this.getElementById(dom, "cam-x") as HTMLInputElement).valueAsNumber;
                        cam.position.y = (this.getElementById(dom, "cam-y") as HTMLInputElement).valueAsNumber;
                        cam.position.z = (this.getElementById(dom, "cam-z") as HTMLInputElement).valueAsNumber;

                        let origin = new Vector3(0, 0, 0);
                        let pos = new Vector3(0, 0, cam.position.z)
                        let dir = origin.sub(pos).normalize();
                        cam.direction = new Vector3(0, 0, -1);
                        cam.update(true);
                    } else {
                        cam = new OrthoCamera(width, height);
                        cam.near = 0;
                        cam.far = 200;
                        cam.direction = new Vector3(0, 0, -1);
                        cam.zoom = 1.0;
                        cam.position.x = 0;
                        cam.position.y = height / 2;
                        cam.position.z = 0;
                        cam.update();
                    }

                    if (this.runtime) {
                        // HACK to override the spinePlay updateCamera function
                        // so that it only uses our own camera control logic
                        this.runtime.spinePlayer.getSceneRenderer().camera = cam;
                        this.runtime.spinePlayer.updateCameraSettings = () => {
                        }
                    }

                    const x = (this.getElementById(dom, "test_x") as HTMLInputElement).valueAsNumber;
                    const y = (this.getElementById(dom, "test_y") as HTMLInputElement).valueAsNumber;
                    const z = (this.getElementById(dom, "test_z") as HTMLInputElement).valueAsNumber;
                    const world_to_screen = this.getElementById(dom, "world_to_screen");
                    const screen_to_world = this.getElementById(dom, "screen_to_world");

                    let w = this.getCorrectCameraPerspectiveZ(cam.near, cam.far, cam.fov);
                    this.getElementById(dom, "correct_cam_z").innerText = `Correct Cam Z: ${w}`;

                    if ((this.getElementById(dom, "transform_dir") as HTMLSelectElement).selectedIndex == 0) {
                        // this.getElementById(dom, "transform_dir_label").innerHTML = "World Coordinates";
                        // Using world coordinates
                        let r1 = cam.worldToScreen(
                            new Vector3(x, y, z),
                        )
                        let r2 = cam.screenToWorld(
                            new Vector3(r1.x, r1.y, r1.z),
                            width, height
                        )
                        world_to_screen.innerText = `${r1.x}, ${r1.y}, ${r1.z}`;
                        screen_to_world.innerText = `${r2.x}, ${r2.y}, ${r2.z}`;
                    } else {
                        // 1000, 540, 0
                        // 1954.1125, 539.99999, 0.16859
                        // 1000, 540, 0.00000						
                        // Using screen coordinates
                        // this.getElementById(dom, "transform_dir_label").innerHTML = "Screen Coordinates";
                        let r2 = cam.screenToWorld(
                            new Vector3(x, y, z),
                            width, height
                        )
                        let r1 = cam.worldToScreen(
                            new Vector3(r2.x, r2.y, r2.z),
                        )
                        world_to_screen.innerText = `${r1.x}, ${r1.y}, ${r1.z}`;
                        screen_to_world.innerText = `${r2.x}, ${r2.y}, ${r2.z}`;
                    }

                } catch (e) {
                    console.log(e);
                    const world_to_screen = this.getElementById(dom, "world_to_screen");
                    world_to_screen.innerText = "Error: " + e;
                }
            });
        });
    }
}