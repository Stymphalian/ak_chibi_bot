import { AnimationStateListener, TrackEntry } from "../core/AnimationState";
import { Actor } from "../player/Actor";
import { ManagedWebGLRenderingContext } from "../webgl/WebGL";
import { Runtime } from "./runtime";
import { Event } from "../core/Event";
import "./stym.css";

export class CanvasRecorder {
    private dom: HTMLElement;
    private parent: HTMLElement | null;
    private runtime: Runtime;

    constructor(parent: HTMLElement | string, runtime: Runtime) {
        if (typeof parent === "string") {
            this.parent = document.getElementById(parent);
        } else {
            this.parent = parent;
        }
        this.parent?.appendChild(this.setupDom());
        this.runtime = runtime;
    }

    private getElementById(dom: HTMLElement, id: string): HTMLElement {
        return this.findWithId(this.dom, id)[0];
    }

    private findWithId(dom: HTMLElement, id: string): HTMLElement[] {
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

    private findWithClass(dom: HTMLElement, className: string): HTMLElement[] {
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

    public createElement(html: string): HTMLElement {
        let dom = document.createElement("div");
        dom.innerHTML = html;
        return dom.children[0] as HTMLElement;
    }

    public setupDom(): HTMLElement {
        let dom = this.dom = this.createElement(/*html*/`
				<div class="save_canvas_container">
                    <label>File Format</label>
                    <select id="file_format_select_id">
                    </select>
                    <br />

                    <label>Output Filename</label>
                    <input id="output_filename" type="text" placeholder="video filename" defaultValue="video.mp4"  size=50 />
                    <br />

                    <div>
                        <label>Record Mode</label>
                        <select id="record_mode_select_id">
                            <option value="manual" selected>Manual</option>
                            <option value="actor">Choose Actor</option>
                            <option value="duration">Duration</option>
                        </select>

                        <div class="record_mode_manual">
                            <button class="record-button">Record</button>
                        </div>

                        <div class="record_mode_actor hidden">
                            <label>Choose Actor</label>
                            <select id="actor_select_id">
                            </select>
                            <button class="save-button-actor">Save</button>
                        </div>

                        <div class="record_mode_duration hidden">
                            <label>Record Duration (seconds)</label>
                            <input id="record_duration_id" type="number" placeholder="start time" defaultValue="2" value="2" step="0.001"/>
                            <button class="save-button-duration">Save</button>
                        </div>

                        <br />
                        <div>Status: <span id="record_status"></span></div>
                    </div>
                    
                    <div class="output_container hidden">
                        <p>Output files</p>
                    </div>
                </div>
			`)
        document.body.appendChild(this.dom);
        this.setupData(this.dom);
        this.addListeners(this.dom);
        return dom;
    }

    private setupData(dom: HTMLElement) {
        let selectElem = this.getElementById(dom, "file_format_select_id");
        const browserSupportedMimeTypes = getAllSupportedMimeTypes("video");
        browserSupportedMimeTypes.forEach(mimeType => {
            selectElem.appendChild(
                this.createElement(`<option value="${mimeType}">${mimeType}</option>`)
            )
        });

        // set default download file name
        (this.getElementById(dom, "output_filename") as HTMLInputElement).value = `vid_${crypto.randomUUID()}.mp4`;
    }

    private setStatus(txt: string, clear ?:boolean) {
        this.getElementById(this.dom, "record_status").innerText = txt;
        if (clear) {
            setTimeout(() => {
                this.getElementById(this.dom, "record_status").innerText = "";
            }, 5000);
        }
    }

    private getOutputFilename(): string {
        return (this.getElementById(this.dom, "output_filename") as HTMLInputElement).value;
    }

    private recordManualStart() {
        let saveContext: SaveContextManual = null;
        return (event: any) => {
            console.log("Record button");
            let dom = this.dom;
            if (saveContext == null) {
                console.log("start manual recording");
                event.target.innerText = "Stop";
                // Start the recording
                let outputFilename = this.getOutputFilename();
                let selectedMimeType = (this.getElementById(dom, "file_format_select_id") as HTMLInputElement).value;
                saveContext = new SaveContextManual(
                    outputFilename,
                    selectedMimeType,
                    this.saveBlob(),
                    this.setStatus.bind(this)
                );
                let removalFn = this.runtime.spinePlayer.registerRenderCallback(
                    saveContext.renderCallback.bind(saveContext)
                );
                saveContext.removalFn = removalFn;

            } else {
                console.log("Stop manual recording");
                event.target.innerText = "Record";
                saveContext.recorder.stop();
                saveContext.removalFn();
                saveContext = null;
            }
        };
    }

    private addListeners(dom: HTMLElement) {
        this.getElementById(dom, "record_mode_select_id").addEventListener("change", ((event: any) => {
            let selection = (event.target as HTMLSelectElement).value;
            let recordModeActor = this.findWithClass(dom, "record_mode_actor")[0];
            let recordModeDuration = this.findWithClass(dom, "record_mode_duration")[0];
            let recordModeManual = this.findWithClass(dom, "record_mode_manual")[0];

            if (selection === "actor") {
                recordModeActor.classList.remove("hidden");
                recordModeDuration.classList.add("hidden");
                recordModeManual.classList.add("hidden");

                recordModeActor.querySelector("select").innerHTML = "";
                this.runtime.spinePlayer.getActorNames().forEach(actorName => {
                    let encodedName = encodeURIComponent(actorName);
                    let opt = this.createElement(`<option value="${encodedName}"></option>`)
                    opt.innerText = `${encodedName}`;
                    recordModeActor.querySelector("select").appendChild(opt);
                })
            } else if (selection === "duration") {
                recordModeActor.classList.add("hidden");
                recordModeDuration.classList.remove("hidden");
                recordModeManual.classList.add("hidden");
            } else if (selection === "manual") {
                recordModeActor.classList.add("hidden");
                recordModeDuration.classList.add("hidden");
                recordModeManual.classList.remove("hidden");
            }
        }).bind(this))

        this.findWithClass(dom, "record-button")[0].addEventListener(
            "click", 
            this.recordManualStart(),
        );
        this.findWithClass(dom, "save-button-actor")[0].addEventListener("click", (event) => {
            let outputFilename = this.getOutputFilename();
            let selectedMimeType = (this.getElementById(dom, "file_format_select_id") as HTMLInputElement).value;
            let encodedSelectedActor = (this.getElementById(dom, "actor_select_id") as HTMLSelectElement).value;
            let decodedSelectedActor = decodeURIComponent(encodedSelectedActor);

            let actor = this.runtime.spinePlayer.getActor(decodedSelectedActor);
            if (!actor || actor === undefined) {
                console.log("Actor not found");
                return;
            }
            const saveContext = new SaveContextActor(
                outputFilename,
                selectedMimeType,
                this.saveBlob(),
                this.setStatus.bind(this)
            );
            saveContext.setup(actor);
            let removalFn = this.runtime.spinePlayer.registerRenderCallback(
                saveContext.renderCallback.bind(saveContext)
            );
            saveContext.removalFn = removalFn;
        });
        this.findWithClass(dom, "save-button-duration")[0].addEventListener("click", (event) => {
            let outputFilename = this.getOutputFilename();
            let selectedMimeType = (this.getElementById(dom, "file_format_select_id") as HTMLInputElement).value;
            let duration = (this.getElementById(dom, "record_duration_id") as HTMLInputElement).valueAsNumber;

            const saveContext = new SaveContextDuration(
                outputFilename,
                selectedMimeType,
                this.saveBlob(),
                this.setStatus.bind(this)
            );
            saveContext.setup(duration);
            let removalFn = this.runtime.spinePlayer.registerRenderCallback(
                saveContext.renderCallback.bind(saveContext)
            );
            saveContext.removalFn = removalFn;
        });
    }

    private saveBlob() {
        const a = document.createElement('a');
        let outputContainer = this.findWithClass(this.dom, "output_container")[0];
        outputContainer.appendChild(a);

        return function saveData(blob: Blob, fileName: string) {
            const url = window.URL.createObjectURL(blob);
            a.href = url;
            a.download = fileName;
            a.click();
            setTimeout(() => { 
                outputContainer.removeChild(a)
            });
        };
    };
}

// @ts-ignore
function getAllSupportedMimeTypes(...mediaTypes) {
    if (!mediaTypes.length) mediaTypes.push('video', 'audio')
    const CONTAINERS = ['webm', 'ogg', 'mp3', 'mp4', 'x-matroska', '3gpp', '3gpp2', '3gp2', 'quicktime', 'mpeg', 'aac', 'flac', 'x-flac', 'wave', 'wav', 'x-wav', 'x-pn-wav', 'not-supported']
    const CODECS = ['vp9', 'vp9.0', 'vp8', 'vp8.0', 'avc1', 'av1', 'h265', 'h.265', 'h264', 'h.264', 'opus', 'vorbis', 'pcm', 'aac', 'mpeg', 'mp4a', 'rtx', 'red', 'ulpfec', 'g722', 'pcmu', 'pcma', 'cn', 'telephone-event', 'not-supported']

    return [...new Set(
        CONTAINERS.flatMap(ext =>
            mediaTypes.flatMap(mediaType => [
                `${mediaType}/${ext}`,
            ]),
        ),
    ), ...new Set(
        CONTAINERS.flatMap(ext =>
            CODECS.flatMap(codec =>
                mediaTypes.flatMap(mediaType => [
                    // NOTE: 'codecs:' will always be true (false positive)
                    `${mediaType}/${ext};codecs=${codec}`,
                ]),
            ),
        ),
    ), ...new Set(
        CONTAINERS.flatMap(ext =>
            CODECS.flatMap(codec1 =>
                CODECS.flatMap(codec2 =>
                    mediaTypes.flatMap(mediaType => [
                        `${mediaType}/${ext};codecs="${codec1}, ${codec2}"`,
                    ]),
                ),
            ),
        ),
    )].filter(variation => MediaRecorder.isTypeSupported(variation))
}

function hexToNumber(hex: string): number {
    const trimmed = hex.trim();
    if (trimmed.length === 0) {
        return 0;
    }
    return parseInt(trimmed, 16);
}

class SaveContext {
    public removalFn: () => void;
    public saveBlobFn: (blob: Blob, fileName: string) => void = null;
    public setStatusFn: (txt:string, clear ?: boolean) => void = null;
    public outputFilename: string = '';
    public mimeType: string = 'video/mp4;codec=vp9';
    public chunks: Blob[] = [];
    public recorder: MediaRecorder;
    public frameCount: number = 0;

    constructor(
        outputFilename: string,
        mimeType: string,
        saveBlobFn: (blob: Blob, fileName: string) => void,
        setStatusFn: (txt: string, clear ?: boolean) => void
    ) {
        this.removalFn = null;
        this.saveBlobFn = saveBlobFn;
        this.setStatusFn = setStatusFn;
        this.outputFilename = outputFilename;
        this.mimeType = mimeType;
    }

    public startRecording(context: ManagedWebGLRenderingContext) {
        const mimetype = this.mimeType;
        if (!MediaRecorder.isTypeSupported(mimetype)) {
            console.log("Failed to start recording. unsupported file format", mimetype);
            return;
        }

        if (this.setStatusFn) {
            this.setStatusFn("Recording");
        }

        try {
            const canvas = context.canvas as HTMLCanvasElement;
            const chunks: Blob[] = []; // here we will store our recorded media chunks (Blobs)
            const stream = canvas.captureStream(); // grab our canvas MediaStream
            const rec = new MediaRecorder(stream, {
                mimeType: mimetype,
            }); // init the recorder

            this.chunks = chunks;
            this.recorder = rec;

            // every time the recorder has new data, we will store it in our array
            rec.ondataavailable = e => chunks.push(e.data);
            // only when the recorder stops, we construct a complete Blob from all the chunks
            rec.onstop = e => {
                this.saveBlobFn(
                    new Blob(chunks, { type: mimetype }),
                    this.outputFilename
                );
                if (this.setStatusFn) {
                    this.setStatusFn("Finished recording", true);
                }
            }
            rec.onerror = (e) => console.log("MediaRecorder error", e);
            rec.start();
        } catch (e) {
            console.log("Failed to start recording", e);
        }
    }

    public renderCallback(context: ManagedWebGLRenderingContext) {
        throw new Error("Method not implemented.");
    }
}

class SaveContextDuration extends SaveContext {
    public durationMs: number = 0;
    public startTimeMsec: number = 0;

    public setup(numSeconds: number) {
        this.durationMs = numSeconds * 1000;
    }

    public renderCallback(context: ManagedWebGLRenderingContext) {
        if (this.frameCount == 0) {
            console.log("Start recording video");
            this.startRecording(context);
            this.startTimeMsec = (new Date().getTime());
        }

        this.frameCount += 1
        let currentTimeMsec = (new Date().getTime());
        let timePassedMsec = currentTimeMsec - this.startTimeMsec;
        if (timePassedMsec >= this.durationMs) {
            console.log("Finished recording video");
            this.removalFn();
            this.recorder.stop();
            return;
        }
    }
}

enum SaveContextActorState {
    FIRST_RENDER_CALL = 0,
    WAIT_FIRST_COMPLETE = 1,
    WAIT_SECOND_COMPLETE = 2,
    WAIT_LAST_FRAME = 3,
}

class SaveContextActor extends SaveContext {
    public actor: Actor;
    public state: SaveContextActorState = SaveContextActorState.FIRST_RENDER_CALL;
    public animationListener: AnimationStateListener = null;

    public setup(actor: Actor) {
        this.actor = actor;
        this.state = SaveContextActorState.FIRST_RENDER_CALL;
        this.animationListener = null;
    }

    public renderCallback(context: ManagedWebGLRenderingContext) {
        if (this.state == SaveContextActorState.FIRST_RENDER_CALL) {
            this.animationListener = {
                complete: (entry: TrackEntry) => {
                    if (this.state == SaveContextActorState.WAIT_FIRST_COMPLETE) {
                        console.log("start recording");
                        this.state = SaveContextActorState.WAIT_SECOND_COMPLETE;
                        this.startRecording(context);
                    } else if (this.state == SaveContextActorState.WAIT_SECOND_COMPLETE) {
                        this.state = SaveContextActorState.WAIT_LAST_FRAME;
                    }
                },
                event: (entry: TrackEntry, event: Event) => { },
                interrupt: (entry: TrackEntry) => { },
                dispose: (entry: TrackEntry) => { },
                start: (entry: TrackEntry) => { },
                end: (entry: TrackEntry) => { },
            }
            this.actor.animationState.addListener(this.animationListener);
            this.frameCount += 1;
            this.state = SaveContextActorState.WAIT_FIRST_COMPLETE;
            return;
        }

        if (this.state == SaveContextActorState.WAIT_SECOND_COMPLETE) {
            this.frameCount += 1;
        } else if (this.state == SaveContextActorState.WAIT_LAST_FRAME) {
            // Done with recording.
            console.log("Done with recording after ", this.frameCount, " frames");
            this.actor.animationState.removeListener(this.animationListener);
            this.removalFn();
            this.recorder.stop();
        }
    }
}

class SaveContextManual extends SaveContext {
    public renderCallback(context: ManagedWebGLRenderingContext) {
        if (this.frameCount == 0) {
            this.startRecording(context);
        }
        this.frameCount += 1;
    }
}