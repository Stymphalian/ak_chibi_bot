# Spine Runtime Architecture Guide (AGENTS.md)

## Overview

The `static/spine` directory contains a **TypeScript WebGL application** that renders animated Arknights chibi characters (sprites) on an OBS browser source. This is a real-time rendering engine that:
- Connects to the Go backend via WebSocket
- Receives commands to spawn/update/remove chibi characters
- Renders Spine 2D skeletal animations or spritesheet animations using WebGL
- Displays chat messages as speech bubbles above characters
- Supports multiple actors (chibis) moving and animating simultaneously on screen
- Shows optional performance panel with FPS and GPU timing (enabled via `?fps=1`)

This is **separate** from the React web app (`static/web_app`) which is the control panel. The Spine app is what streamers add to OBS as a browser source.

---

## Technology Stack

- **Language**: TypeScript
- **Graphics**: WebGL with custom Spine Runtime (3.8)
- **Build Tool**: Webpack 5
- **Animation**: Spine 2D skeletal animation + custom spritesheet system
- **Communication**: WebSocket (connects to Go backend)
- **Rendering**: Custom SceneRenderer with perspective camera

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Entry Point (index.ts)                │
│  - Parse URL query params (channelName, config)          │
│  - Create Runtime instance                               │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│                Runtime (runtime.ts)                      │
│  - WebSocket connection to backend                       │
│  - Message handler (SET_OPERATOR, REMOVE_OPERATOR, etc.) │
│  - Creates SpinePlayer                                   │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│              SpinePlayer (Player.ts)                     │
│  - Asset loading (skeletons, atlases, textures)          │
│  - Actor registry (Map<username, Actor>)                 │
│  - Rendering loop (drawFrame)                            │
│  - Canvas management (WebGL + 2D overlay)                │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│                Actor (Actor.ts)                          │
│  - Individual chibi character                            │
│  - Skeleton/spritesheet data                             │
│  - Position, velocity, movement logic                    │
│  - Animation state management                            │
│  - Chat message queue                                    │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│             Actions (Action.ts)                          │
│  - Movement behaviors (Wander, Walk, Follow, etc.)       │
│  - Physics updates                                       │
│  - Animation selection                                   │
└──────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. **Entry Point** ([index.ts](src/index.ts))

**Purpose**: Initialize the application from URL parameters.

**Key Responsibilities**:
- Parse `channelName` from URL query string
- Extract configuration (width, height, debug mode, scale, etc.)
- Create `Runtime` instance
- Attach to `window.SpineRuntime` for debugging

**URL Parameters**:
```
?channelName=stymphalian2__
&width=1920
&height=1080
&debug=true
&scale=1.0
&show_chat=true
&chibi_ocean=true
&blacklist=user1,user2
&premultiplied_alpha=true
```

**Code Flow**:
```typescript
// Parse URL params
const channelName = searchParams.get('channelName')
const config = getRuntimeConfigFromQueryParams(searchParams)

// Create runtime
window.SpineRuntime = new Runtime(channelName, config)
```

**Key Files**:
- [index.ts](src/index.ts): Entry point
- [stym/utils.ts](src/stym/utils.ts): `getRuntimeConfigFromQueryParams()`

---

### 2. **Runtime System** ([stym/runtime.ts](src/stym/runtime.ts))

**Purpose**: Manage WebSocket connection and coordinate message handling.

**Key Data Structure**:
```typescript
interface RuntimeConfig {
    width: number                    // Canvas width
    height: number                   // Canvas height
    debugMode: boolean              // Show debug overlays
    chibiScale: number              // Global chibi scaling (0.1 - 3.0)
    showChatMessagesFlag: boolean   // Display chat bubbles
    usernameBlacklist: string[]     // Users to ignore
    excessiveChibiMitigations: boolean  // Reduce visual clutter
    usePremultipliedAlpha: boolean  // Alpha blending mode
}
```

**WebSocket Protocol**:

The runtime connects to: `ws://host/ws/?channelName={channelName}`

**Message Types**:

1. **SET_OPERATOR** - Add or update a chibi
```json
{
  "type_name": "SET_OPERATOR",
  "user_name": "stymphalian2__",
  "user_name_display": "Stymphalian2__",
  "operator_id": "char_002_amiya",
  "skel_file": "/assets/characters/char_002_amiya.skel",
  "atlas_file": "/assets/characters/char_002_amiya.atlas",
  "spritesheet_data_filepath": null,
  "start_pos": {"x": 0.5, "y": 0.0},
  "sprite_scale": {"x": 1.0, "y": 1.0},
  "max_sprite_pixel_size": 350,
  "movement_speed": {"x": 1.0, "y": 1.0},
  "animation_speed": 1.0,
  "action": "WANDER",
  "action_data": {...}
}
```

2. **REMOVE_OPERATOR** - Remove a chibi
```json
{
  "type_name": "REMOVE_OPERATOR",
  "user_name": "stymphalian2__"
}
```

3. **SHOW_CHAT_MESSAGE** - Display chat bubble
```json
{
  "type_name": "SHOW_CHAT_MESSAGE",
  "user_name": "stymphalian2__",
  "message": "Hello world!"
}
```

4. **FIND_OPERATOR** - Highlight chibi (flash effect)
```json
{
  "type_name": "FIND_OPERATOR",
  "user_name": "stymphalian2__"
}
```

**Message Handler Flow**:
```typescript
messageHandler(event: MessageEvent) {
    let requestData = JSON.parse(event.data)
    switch(requestData["type_name"]) {
        case "SET_OPERATOR":
            this.swapCharacter(requestData)
            break
        case "REMOVE_OPERATOR":
            this.removeCharacter(requestData)
            break
        case "SHOW_CHAT_MESSAGE":
            this.showChatMessage(requestData)
            break
        case "FIND_OPERATOR":
            this.findCharacter(requestData)
            break
    }
}
```

**Connection Management**:
- **Exponential backoff**: Starts at 15s, doubles on failure, max 5 minutes
- **Auto-reconnect**: Automatically retries on disconnect
- **State sync**: Notifies SpinePlayer of connection status

**Code Pointers**:
- [runtime.ts:72-110](src/stym/runtime.ts): `openWebSocket()` - Connection setup
- [runtime.ts:112-129](src/stym/runtime.ts): `messageHandler()` - Message router
- [runtime.ts:131-226](src/stym/runtime.ts): `swapCharacter()` - Process SET_OPERATOR

---

### 3. **SpinePlayer** ([player/Player.ts](src/player/Player.ts))

**Purpose**: Main rendering orchestrator - manages actors, assets, and rendering loop.

**Key Data Structures**:
```typescript
class SpinePlayer {
    private actors: Map<string, Actor>  // username -> Actor
    private assetManager: AssetManager  // Asset loading
    private sceneRenderer: SceneRenderer // WebGL renderer
    private canvas: HTMLCanvasElement   // WebGL canvas
    private textCanvas: HTMLCanvasElement // 2D overlay for text
    private actorQueue: Array<{actorName, config}>  // Deferred actor updates
    private webSocket: WebSocket | null
}
```

**Rendering Pipeline**:

```
┌─────────────────────┐
│  drawFrame()        │ (called by requestAnimationFrame)
│  - Sort actors by Z │
│  - Filter loaded    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  updateActors()     │
│  - Update physics   │
│  - Update animations│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  bindForDraw()      │
│  - Clear canvas     │
│  - Update camera    │
│  - Clear text layer │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  drawActors()       │
│  - Render skeletons │
│  - Render debug     │
│  - Render text      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ postProcessActors() │
│  - Process animation│
│    state changes    │
└─────────────────────┘
```

**Actor Management**:

**Add/Update Actor**:
```typescript
addActorToUpdateQueue(actorName: string, config: SpineActorConfig) {
    // Queues actor for addition
    // Processed when asset loading completes
}

changeOrAddActor(actorName: string, config: SpineActorConfig) {
    if (this.actors.has(actorName)) {
        // Update existing actor
        let actor = this.actors.get(actorName)
        actor.ResetWithConfig(config)
        this.setupActor(actor)
    } else {
        // Create new actor
        let actor = new Actor(config, viewport, offscreenRender, mitigations)
        this.setupActor(actor)
        this.actors.set(actorName, actor)
    }
}
```

**Actor Queue System**:
- Actors are queued rather than added immediately
- Queue is consumed when asset manager is idle (every 100ms)
- Prevents blocking on asset loading

**Z-Ordering Algorithm**:
```typescript
// Sort actors for rendering (back to front)
actorsZOrder = Array.from(this.actors.keys()).sort((a, b) => {
    let a1 = this.actors.get(a)
    let a2 = this.actors.get(b)
    
    // Flashing actors always on top
    if (a1.ShouldFlashCharacter && !a2.ShouldFlashCharacter) return 1
    if (a2.ShouldFlashCharacter && !a1.ShouldFlashCharacter) return -1
    
    // Sort by Z position (depth)
    let r = a1.getPositionZ() - a2.getPositionZ()
    
    // Stable sort: older actors rendered first
    if (r == 0) return a1.lastUpdatedWhen - a2.lastUpdatedWhen
    return r
})
```

**Asset Loading**:
- **Spine Skeletons**: `.skel` (binary) or `.json` (text) + `.atlas` texture atlas
- **Spritesheets**: Custom JSON config + PNG textures
- **Lazy Loading**: Assets loaded on-demand when actor is added
- **Retry Logic**: Up to 10 retries per actor on load failure

**Camera System**:
- **Perspective camera** with configurable near/far planes
- **Origin**: Center-bottom of viewport (0,0 = bottom-center)
- **Z-axis**: Negative values = further back (depth)
- **FOV**: Fixed field of view with zoom support

**Code Pointers**:
- [Player.ts:171-220](src/player/Player.ts): `SpinePlayer` class definition
- [Player.ts:511-535](src/player/Player.ts): `addActorToUpdateQueue()` - Actor queueing
- [Player.ts:537-555](src/player/Player.ts): `changeOrAddActor()` - Add/update logic
- [Player.ts:771-900](src/player/Player.ts): `drawFrame()` - Main render loop
- [Player.ts:850-928](src/player/Player.ts): `loadActor()` - Asset loading

---

### 4. **Actor** ([player/Actor.ts](src/player/Actor.ts))

**Purpose**: Individual chibi character with physics, animation, and rendering state.

**Key Data Structure**:
```typescript
class Actor {
    // Core state
    public skeleton: Skeleton              // Spine skeleton data
    public spritesheet: SpritesheetActor   // OR spritesheet (alternative)
    public animationState: AnimationState  // Animation controller
    public loaded: boolean                 // Asset loading complete
    
    // Transform
    private position: Vector3              // World position (x, y, z)
    private velocity: Vector3              // Movement velocity (px/sec)
    private scale: Vector2                 // Sprite scale (x, y)
    private movementSpeed: Vector3         // Max speed (px/sec)
    private skeletonPositionOffset: Vector3 // Offset for sitting/flying
    
    // Visual
    public canvasBB: BoundingBox          // Rendering bounding box
    private speechBubble: SpeechBubble    // Name tag + chat renderer
    private messageQueue: ChatMessageQueue // Chat message queue
    
    // Action/Movement
    public currentAction: ActorAction     // Movement behavior
    public dirtyAnimation: boolean        // Needs animation update
    
    // Configuration
    public config: SpineActorConfig       // Actor settings
    
    // Mitigation flags
    private wanderActionWithPeriodicStoppingFlag: boolean
    private spreadNameTagsFlag: boolean
    private headerBandIndex: number       // Name tag height band
    public ShouldFlashCharacter: boolean  // Highlight effect
}
```

**Position System**:
- **World Coordinates**: Origin at (0, 0) = bottom-center of viewport
- **X-axis**: Negative = left, Positive = right
- **Y-axis**: 0 = bottom, height = top
- **Z-axis**: 0 = front, Negative = back (depth for sorting)

Example:
```typescript
// Start position from config (0-1 normalized)
config.startPosX = 0.5  // Center
config.startPosY = 0.0  // Bottom

// Converted to world coordinates
position.x = (0.5 * viewport.width) - (viewport.width / 2)  // = 0 (center)
position.y = 0.0 * viewport.height                          // = 0 (bottom)
position.z = 0                                              // Front
```

**Physics Update**:
```typescript
UpdatePhysics(player: SpinePlayer, deltaSecs: number) {
    // 1. Update chat message queue (show/hide)
    this.messageQueue.Update(deltaSecs)
    
    // 2. Update action (calculates velocity)
    this.currentAction.UpdatePhysics(this, deltaSecs, viewport, player)
    
    // 3. Apply velocity to position
    this.position.x += this.velocity.x
    this.position.y += this.velocity.y
    this.position.z += this.velocity.z
    
    // 4. Auto-flip sprite based on movement direction
    if (this.velocity.x > 0) {
        this.scale.x = this.config.scaleX   // Face right
    } else if (this.velocity.x < 0) {
        this.scale.x = -this.config.scaleX  // Face left
    }
    
    // 5. Update skeleton transform
    this.setSkeletonMovementData()
}
```

**Animation State Management**:

Actors use **deferred animation updates** to avoid GL context conflicts:

```typescript
// In UpdatePhysics: Mark animation as needing change
QueueAnimationStateChange() {
    this.dirtyAnimation = true
}

// After render frame: Apply animation change
ProcessAnimationStateChange() {
    if (this.dirtyAnimation) {
        this.dirtyAnimation = false
        this.InitAnimationState()  // Reset animation state
    }
}
```

**Bounding Box System**:

Two types of bounding boxes:

1. **Render Bounding Box** (`canvasBB`): Size of sprite on screen
2. **World Bounding Box** (`GetBoundingBox()`): World-space position

```typescript
GetBoundingBox(): BoundingBox {
    let renderBB = this.getRenderingBoundingBox()
    
    if (this.isFacingRight()) {
        return {
            x: this.position.x + renderBB.x,
            y: this.position.y,
            width: renderBB.width,
            height: renderBB.height
        }
    } else {
        // Flip bounding box for left-facing sprites
        return {
            x: this.position.x - renderBB.x - renderBB.width,
            y: this.position.y,
            width: renderBB.width,
            height: renderBB.height
        }
    }
}
```

**Rendering**:

```typescript
Draw(sceneRenderer: SceneRenderer) {
    if (this.isSpritesheetActor()) {
        // Spritesheet rendering
        let coords = this.spritesheet.GetUVFromFrame()
        if (this.isFacingLeft()) {
            // Flip UVs for left-facing
            [coords.U1, coords.U2] = [coords.U2, coords.U1]
        }
        const bb = this.GetBoundingBox()
        sceneRenderer.drawTextureUV(
            this.spritesheet.GetTexture(),
            bb.x, bb.y, this.getPositionZ(),
            bb.width, bb.height,
            coords.U1, coords.V1, coords.U2, coords.V2,
            this.spritesheet.highlightColor
        )
    } else {
        // Spine skeleton rendering
        sceneRenderer.drawSkeleton(
            this.skeleton, 
            this.config.premultipliedAlpha
        )
    }
}
```

**Chat Message Queue**:

Speech bubbles are shown above actors for 5 seconds each:

```typescript
EnqueueChatMessage(message: string) {
    this.messageQueue.AddMessage(message)
}

// In rendering:
DrawText(camera, context, showChatMessages) {
    let chatMessages = this.GetChatMessages()
    if (chatMessages && showChatMessages) {
        this.speechBubble.ChatText(
            viewport, camera, context,
            chatMessages.messages,
            this.getPositionX(),
            this.getPositionY() + this.GetUsernameHeaderHeight(),
            this.getPositionZ()
        )
    }
    
    // Always show name tag
    this.speechBubble.NameTag(
        viewport, camera, context,
        this.config.userDisplayName,
        this.getPositionX(),
        this.getPositionY() + this.GetUsernameHeaderHeight(),
        this.getPositionZ()
    )
}
```

**Excessive Chibi Mitigations**:

When `excessiveChibiMitigation` is enabled:
1. **Periodic Stopping**: Wandering actors pause periodically
2. **Spread Name Tags**: Name tags placed at random heights to reduce overlap

```typescript
// Name tag height distribution
GetUsernameHeaderHeight() {
    if (this.spreadNameTagsFlag && Actor.averageActorHeight > 0) {
        return Math.max(
            this.getRenderingBoundingBox().height + 10,
            Actor.averageActorHeight + 
            this.headerBandIndex * Actor.HEADER_BANDS_HEIGHT
        )
    } else {
        return this.getRenderingBoundingBox().height + 10
    }
}
```

**Code Pointers**:
- [Actor.ts:45-175](src/player/Actor.ts): `SpineActorConfig` interface
- [Actor.ts:177-240](src/player/Actor.ts): `Actor` class properties
- [Actor.ts:242-280](src/player/Actor.ts): Constructor
- [Actor.ts:421-450](src/player/Actor.ts): `ResetWithConfig()` - Actor update
- [Actor.ts:452-478](src/player/Actor.ts): `UpdatePhysics()` - Physics loop
- [Actor.ts:665-690](src/player/Actor.ts): `Draw()` - Rendering
- [Actor.ts:692-720](src/player/Actor.ts): `DrawText()` - Text overlay

---

### 5. **Actions System** ([player/Action.ts](src/player/Action.ts))

**Purpose**: Define movement behaviors for actors (AI).

**Action Interface**:
```typescript
interface ActorAction {
    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox): void
    GetAnimations(): string[]
    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox): void
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer): void
}
```

**Available Actions**:

#### 1. **PLAY_ANIMATION** - Static animation
```typescript
{
  "action": "PLAY_ANIMATION",
  "action_data": {
    "animations": ["Idle", "Sit", "Sleep"]
  }
}
```
- Plays animations in sequence
- If animation includes "Move", wanders randomly
- Otherwise, stays stationary

#### 2. **WALK** - Continuous walking
```typescript
{
  "action": "WALK",
  "action_data": {
    "walk_animation": "Move"
  }
}
```
- Randomly picks X position
- Walks to it
- Picks new position and repeats

#### 3. **WANDER** - Walk with periodic stops
```typescript
{
  "action": "WANDER",
  "action_data": {
    "wander_animation": "Move",
    "idle_animation": "Idle"
  }
}
```
- Walks to random position
- Stops and plays idle (10-30s)
- Repeats

States:
- `WANDER_WALK`: Moving to destination
- `WANDER_IDLE`: Stationary, playing idle animation

#### 4. **WALK_TO** - Walk to specific position
```typescript
{
  "action": "WALK_TO",
  "action_data": {
    "walk_animation": "Move",
    "idle_animation": "Idle",
    "target_x": 0.5,  // 0-1 normalized
    "target_y": 0.0
  }
}
```
- Walks to specific target
- Plays idle when arrived

#### 5. **PACE_AROUND** - Walk back and forth
```typescript
{
  "action": "PACE_AROUND",
  "action_data": {
    "walk_animation": "Move",
    "idle_animation": "Idle",
    "pace_distance_px": 200
  }
}
```
- Walks left/right within distance
- Turns around at boundaries
- Optional idle time at ends

#### 6. **FOLLOW** - Follow another actor
```typescript
{
  "action": "FOLLOW",
  "action_data": {
    "walk_animation": "Move",
    "idle_animation": "Idle",
    "target_username": "stymphalian2__",
    "follow_offset_x": 100,
    "follow_distance": 50
  }
}
```
- Follows target actor
- Maintains offset distance
- Plays idle when close enough

**Movement Algorithm**:

All actions use this velocity calculation:

```typescript
function updateVelocityFromDir(
    actor: Actor, 
    endPosition: Vector3, 
    dir: Vector3,  // Normalized direction
    deltaSecs: number
) {
    let actorPosition = actor.getPosition3()
    let step = actor.getMovementSpeed().scale(deltaSecs)
    
    // Don't overshoot target
    let stepDist = endPosition.subtract(actorPosition)
    if (step.x > Math.abs(stepDist.x)) step.x = Math.abs(stepDist.x)
    if (step.y > Math.abs(stepDist.y)) step.y = Math.abs(stepDist.y)
    if (step.z > Math.abs(stepDist.z)) step.z = Math.abs(stepDist.z)
    
    actor.setVelocity(
        dir.x * step.x,
        dir.y * step.y,
        dir.z * step.z
    )
}
```

**Animation Positioning**:

Different animations need different Y offsets:

```typescript
function setActorYPositionByAnimation(actor, animation, viewport) {
    const startPosYScaled = actor.config.startPosY * viewport.height
    actor.setPositionY(startPosYScaled)
    actor.setSkeletonPositionOffsetY(0)
    
    let isSitting = animation.toLowerCase().includes("sit")
    
    if (actor.canvasBB.y < 0) {
        // Sprite extends below origin (sitting)
        if (actor.IsEnemySprite() || isSitting) {
            actor.setSkeletonPositionOffsetY(-actor.canvasBB.y)
        }
    } else if (actor.canvasBB.y > 0) {
        // Sprite extends above origin (flying)
        if (actor.IsEnemySprite()) {
            actor.setSkeletonPositionOffsetY(-actor.canvasBB.y)
        }
    }
}
```

**Code Pointers**:
- [Action.ts:7-27](src/player/Action.ts): `ActorAction` interface
- [Action.ts:29-55](src/player/Action.ts): `ParseActionNameToAction()` - Factory
- [Action.ts:139-187](src/player/Action.ts): `PlayAnimationAction`
- [Action.ts:189-233](src/player/Action.ts): `WalkAction`
- [Action.ts:235-330](src/player/Action.ts): `WanderAction`
- [Action.ts:332-405](src/player/Action.ts): `WalkToAction`
- [Action.ts:407-530](src/player/Action.ts): `PaceAroundAction`
- [Action.ts:532-630](src/player/Action.ts): `FollowAction`

---

### 6. **Chat Messages** ([player/ChatMessages.ts](src/player/ChatMessages.ts))

**Purpose**: Queue and display chat messages as speech bubbles.

**Data Structures**:
```typescript
class MessageBlock {
    public messages: string[]  // Lines of text
}

class ChatMessageQueue {
    private messages: MessageBlock[]         // Queue of blocks
    private currentMessage: MessageBlock|null // Currently showing
    private showTimeRemaining: number        // Seconds left to show
    private showing: boolean
}
```

**Configuration**:
```typescript
const SHOW_DURATION_SECS = 5        // Show each block for 5s
const MAX_CHARS_PER_LINE = 30       // Line wrap at 30 chars
const MAX_LINES_PER_BLOCK = 3       // Max 3 lines per bubble
```

**Message Processing**:

```typescript
AddMessage(message: string) {
    // 1. Split by whitespace
    let words = message.split(/\s+/)
    
    // 2. Split long words
    words = words.flatMap(word => {
        if (word.length > MAX_CHARS_PER_LINE) {
            return word.match(/.{1,30}/g)  // Split into 30-char chunks
        }
        return word
    })
    
    // 3. Combine into lines
    let lines = []
    let currentLine = ""
    for (let word of words) {
        if (currentLine.length + word.length + 1 > MAX_CHARS_PER_LINE) {
            lines.push(currentLine)
            currentLine = word
        } else {
            currentLine += " " + word
        }
    }
    if (currentLine) lines.push(currentLine)
    
    // 4. Group lines into blocks (3 lines per block)
    for (let i = 0; i < lines.length; i += MAX_LINES_PER_BLOCK) {
        this.messages.push(new MessageBlock(
            lines.slice(i, i + MAX_LINES_PER_BLOCK)
        ))
    }
}
```

**Display Logic**:
```typescript
Update(deltaSecs: number) {
    if (this.currentMessage == null) {
        if (this.HasMessages()) {
            this.currentMessage = this.messages.shift()
            this.showTimeRemaining = SHOW_DURATION_SECS
            this.showing = true
        }
        return
    }
    
    if (this.showing) {
        this.showTimeRemaining -= deltaSecs
        if (this.showTimeRemaining <= 0) {
            if (this.messages.length > 0) {
                // Show next block
                this.currentMessage = this.messages.shift()
                this.showTimeRemaining = SHOW_DURATION_SECS
            } else {
                // No more messages
                this.showing = false
                this.currentMessage = null
            }
        }
    }
}
```

**Code Pointers**:
- [ChatMessages.ts:1-79](src/player/ChatMessages.ts): Full implementation

---

### 7. **Speech Bubbles** ([player/SpeechBubble.ts](src/player/SpeechBubble.ts))

**Purpose**: Render name tags and chat bubbles on 2D canvas overlay.

**Key Methods**:

```typescript
class SpeechBubble {
    nameTagSize: any = null  // Cached for performance
    
    // Render username above actor
    NameTag(viewport, camera, ctx, text, xpx, ypx, zpx)
    
    // Render chat message above actor
    ChatText(viewport, camera, ctx, texts[], xpx, ypx, zpx)
    
    // Internal rendering
    private _drawSpeechBubble(ctx, width, height)
}
```

**Rendering Process**:

1. **World to Screen Conversion**:
```typescript
// Convert 3D world coordinates to 2D screen coordinates
let tt = camera.worldToScreen(new Vector3(xpx, ypx, zpx))
let screenPos = new Vector2(tt.x, viewport.height - tt.y)
```

2. **Text Measurement**:
```typescript
// Measure text dimensions
let data = ctx.measureText(text)
let height = data.actualBoundingBoxAscent - data.actualBoundingBoxDescent
let width = data.width
```

3. **Bubble Drawing**:
```typescript
// Black rounded rectangle
ctx.fillStyle = "black"
ctx.strokeStyle = "#333"
ctx.roundRect(-width/2 - pad, pad, width + 2*pad, -height - 2*pad, 3)
ctx.stroke()
ctx.fill()

// Triangle pointer
ctx.moveTo(-5, pad)
ctx.lineTo(0, pad + 5)
ctx.lineTo(5, pad)
ctx.fill()
```

4. **Text Rendering**:
```typescript
// White text
ctx.fillStyle = "white"
ctx.fillText(text, -width/2, y)
```

**Optimization**: Name tag size is cached after first render to avoid repeated text measurement.

**Code Pointers**:
- [SpeechBubble.ts:1-130](src/player/SpeechBubble.ts): Full implementation

---

### 8. **Spritesheet System** ([player/Spritesheet.ts](src/player/Spritesheet.ts))

**Purpose**: Alternative to Spine skeletal animation - simple sprite sheet animation.

**Data Structures**:
```typescript
class SpritesheetAnimationConfig {
    filepath: string    // Path to PNG
    rows: number        // Grid rows
    cols: number        // Grid columns
    width: number       // Frame width (px)
    height: number      // Frame height (px)
    frames: number      // Total frames
    fps: number         // Playback speed
    scaleX: number      // X scale
    scaleY: number      // Y scale
}

class SpritesheetConfig {
    animations: Map<string, SpritesheetAnimationConfig>
}

class SpritesheetActor {
    config: SpritesheetConfig
    textures: Map<string, Texture>
    animationName: string
    currentFrame: number
    trackTime: number
    timeScale: number
}
```

**JSON Config Format**:
```json
{
  "animations": {
    "Idle": {
      "filepath": "idle_spritesheet.png",
      "rows": 4,
      "cols": 8,
      "width": 128,
      "height": 128,
      "frames": 32,
      "fps": 24,
      "scaleX": 1.0,
      "scaleY": 1.0
    },
    "Move": {
      "filepath": "walk_spritesheet.png",
      "rows": 2,
      "cols": 8,
      "frames": 16,
      "fps": 12
    }
  }
}
```

**UV Coordinate Calculation**:

Sprite sheets are grids of animation frames. To render frame N:

```typescript
GetUVFromFrame(): UVCoords {
    const frameIndex = this.currentFrame
    
    // Calculate grid position
    const row = Math.floor(frameIndex / this.animationConfig.cols)
    const col = frameIndex % this.animationConfig.cols
    
    // Convert to UV coordinates (0-1 range)
    const u1 = col / this.animationConfig.cols
    const v1 = (row + 1) / this.animationConfig.rows  // Flip V
    const u2 = (col + 1) / this.animationConfig.cols
    const v2 = row / this.animationConfig.rows
    
    return {U1: u1, V1: v1, U2: u2, V2: v2}
}
```

Example: 4x8 grid (32 frames), frame 17:
- row = 17 / 8 = 2
- col = 17 % 8 = 1
- U1 = 1/8 = 0.125, U2 = 2/8 = 0.25
- V1 = 3/4 = 0.75, V2 = 2/4 = 0.5

**Animation Update**:
```typescript
Update(delta: number) {
    delta *= this.timeScale
    this.trackTime += delta
    
    // Loop animation
    const trackTimeLocal = this.trackTime % this.totalAnimationDuration
    
    // Calculate current frame
    this.currentFrame = Math.floor(trackTimeLocal / this.durationPerFrame)
}
```

**Code Pointers**:
- [Spritesheet.ts:1-120](src/player/Spritesheet.ts): Full implementation

---

### 9. **Utilities** ([player/Utils.ts](src/player/Utils.ts), [stym/utils.ts](src/stym/utils.ts))

**Camera Configuration**:

```typescript
function configurePerspectiveCamera(cam: Camera, near: number, far: number, viewport: BoundingBox) {
    cam.near = near
    cam.far = far
    cam.zoom = 1
    cam.position.x = 0
    cam.position.y = viewport.height / 2
    cam.position.z = -getPerspectiveCameraZOffset(viewport, near, far, cam.fov)
    cam.direction = new Vector3(0, 0, -1)
    cam.update()
}
```

**Username Validation**:

```typescript
function isAlphanumeric(str: string): boolean {
    return /^[a-zA-Z0-9_-]{1,100}$/.test(str)
}

function isValidTwitchUserDisplayName(str: string): boolean {
    // Allows Unicode (Japanese/Chinese/Korean)
    // Blocks HTML/path characters
    return /^[^<>%&\\\/]{1,100}$/.test(str)
}
```

**Bounding Box**:
```typescript
interface BoundingBox {
    x: number      // Bottom-left X
    y: number      // Bottom-left Y
    width: number
    height: number
}
```

**Code Pointers**:
- [player/Utils.ts:1-170](src/player/Utils.ts): Camera, DOM, validation utils
- [stym/utils.ts:1-70](src/stym/utils.ts): URL parsing, config extraction

---

## Data Flow Examples

### Example 1: User Types `!chibi Amiya`

```
1. Twitch Chat → Go Backend
   - User: "!chibi Amiya"
   - Backend parses command
   - Looks up character "Amiya" → char_002_amiya

2. Go Backend → WebSocket Message
   {
     "type_name": "SET_OPERATOR",
     "user_name": "stymphalian2__",
     "operator_id": "char_002_amiya",
     "skel_file": "/assets/characters/char_002_amiya.skel",
     "atlas_file": "/assets/characters/char_002_amiya.atlas",
     "action": "WANDER",
     "action_data": {"wander_animation": "Move", "idle_animation": "Idle"}
   }

3. Runtime.messageHandler()
   - Receives message
   - Calls swapCharacter()
   - Creates SpineActorConfig
   - Calls spinePlayer.addActorToUpdateQueue()

4. SpinePlayer.consumeActorUpdateQueue()
   - Checks if assets loading complete
   - If new actor: creates Actor instance
   - If existing: calls actor.ResetWithConfig()
   - Calls setupActor() to load assets

5. SpinePlayer.loadActor()
   - AssetManager loads .skel and .atlas
   - Parses skeleton data
   - Creates AnimationState
   - Calls actor.config.success()

6. Actor.InitAnimationState()
   - Sets up AnimationStateData
   - Adds animation listener
   - Starts playing animations

7. Rendering Loop (every frame)
   - SpinePlayer.drawFrame()
   - actor.Update() → physics, animation
   - actor.Draw() → renders skeleton
   - actor.DrawText() → renders name tag
```

### Example 2: User Types in Chat

```
1. Twitch Chat → Go Backend
   - User: "Hello world!"
   - Backend creates SHOW_CHAT_MESSAGE

2. WebSocket → Runtime
   {
     "type_name": "SHOW_CHAT_MESSAGE",
     "user_name": "stymphalian2__",
     "message": "Hello world!"
   }

3. Runtime.messageHandler()
   - Calls showChatMessage()
   - Calls spinePlayer.showChatMessage()

4. Actor.EnqueueChatMessage()
   - Splits message into lines (30 chars/line)
   - Groups into blocks (3 lines/block)
   - Adds to messageQueue

5. Rendering Loop
   - messageQueue.Update() → shows for 5s
   - actor.DrawText() → renders speech bubble
   - speechBubble.ChatText() → draws on 2D canvas
```

### Example 3: Movement Physics (Wander Action)

```
Every Frame (60 FPS):

1. SpinePlayer.updateActors()
   - Calls actor.Update()

2. Actor.Update()
   - time.update() → calculate delta
   - UpdatePhysics(delta)

3. Actor.UpdatePhysics()
   - currentAction.UpdatePhysics()  [WanderAction]

4. WanderAction.UpdatePhysics()
   State: WANDER_WALK
   - Calculate direction to endPosition
   - updateVelocityFromDir()
   - If reached: set state = WANDER_IDLE

5. Actor.UpdatePhysics() (continued)
   - position += velocity
   - Auto-flip sprite based on velocity.x
   - setSkeletonMovementData()

6. Actor.Draw()
   - sceneRenderer.drawSkeleton()
   - Skeleton rendered at (position.x, position.y, position.z)
```

---

## Key Algorithms

### 1. **Z-Order Sorting (Depth Sorting)**

Purpose: Render actors back-to-front for proper occlusion.

```typescript
actorsZOrder.sort((a, b) => {
    let a1 = this.actors.get(a)
    let a2 = this.actors.get(b)
    
    // Priority 1: Flashing actors on top
    if (a1.ShouldFlashCharacter && !a2.ShouldFlashCharacter) return 1
    if (a2.ShouldFlashCharacter && !a1.ShouldFlashCharacter) return -1
    
    // Priority 2: Z position (depth)
    let r = a1.getPositionZ() - a2.getPositionZ()
    
    // Priority 3: Creation time (stable sort)
    if (r == 0) return a1.lastUpdatedWhen - a2.lastUpdatedWhen
    return r
})
```

**Complexity**: O(n log n)

### 2. **Velocity Clamping (Movement)**

Purpose: Prevent overshooting target position.

```typescript
function updateVelocityFromDir(actor, endPosition, dir, deltaSecs) {
    let actorPosition = actor.getPosition3()
    let step = actor.getMovementSpeed().scale(deltaSecs)
    let stepDist = endPosition.subtract(actorPosition)
    
    // Clamp velocity to not overshoot
    if (step.x > Math.abs(stepDist.x)) step.x = Math.abs(stepDist.x)
    if (step.y > Math.abs(stepDist.y)) step.y = Math.abs(stepDist.y)
    if (step.z > Math.abs(stepDist.z)) step.z = Math.abs(stepDist.z)
    
    actor.setVelocity(dir.x * step.x, dir.y * step.y, dir.z * step.z)
}
```

Without clamping, actors would oscillate around the target.

### 3. **Text Line Wrapping**

Purpose: Wrap long chat messages into multiple lines.

```typescript
AddMessage(message: string) {
    let words = message.split(/\s+/)
    
    // Split long words
    words = words.flatMap(word => {
        if (word.length > MAX_CHARS_PER_LINE) {
            return word.match(/.{1,30}/g)
        }
        return word
    })
    
    // Greedy line packing
    let lines = []
    let currentLine = ""
    for (let word of words) {
        if (currentLine.length + word.length + 1 > MAX_CHARS_PER_LINE) {
            lines.push(currentLine)
            currentLine = word
        } else {
            currentLine += " " + word
        }
    }
    if (currentLine) lines.push(currentLine)
}
```

**Complexity**: O(n) where n = number of words

### 4. **Exponential Backoff (WebSocket Reconnection)**

Purpose: Retry connection with increasing delays.

```typescript
this.backoffTimeMsec = 15000  // Start at 15s
this.backOffMaxtimeMsec = 5 * 60 * 1000  // Max 5 minutes

socket.addEventListener("close", (event) => {
    this.backoffTimeMsec *= 2  // Double on each failure
    if (this.backoffTimeMsec < this.backOffMaxtimeMsec) {
        setTimeout(() => this.openWebSocket(channelName), this.backoffTimeMsec)
    }
})

socket.addEventListener("open", (event) => {
    this.backoffTimeMsec = this.defaultBackoffTimeMsec  // Reset on success
})
```

Backoff sequence: 15s → 30s → 60s → 120s → 240s → 300s (capped)

### 5. **Perspective Camera Z-Offset Calculation**

Purpose: Position camera so viewport matches desired dimensions.

```typescript
function getPerspectiveCameraZOffset(viewport, near, far, fovY) {
    let cam = new PerspectiveCamera(viewport.width, viewport.height)
    cam.near = near
    cam.far = far
    cam.fov = fovY
    cam.update()
    
    // Extract projection matrix scale factor
    let a = cam.projectionView.values[M00]
    
    // Calculate required Z distance
    let w = -a * (viewport.width / 2)
    return w
}
```

This ensures the camera frustum exactly matches the viewport size.

---

## Performance Considerations

### 1. **Asset Loading**

**Problem**: Loading multiple actors simultaneously can freeze the browser.

**Solution**: Actor queue system
```typescript
addActorToUpdateQueue(actorName, config) {
    this.actorQueue.push({actorName, config})
    if (this.actorQueueIndex == null) {
        this.actorQueueIndex = setTimeout(() => this.consumeActorUpdateQueue(), 100)
    }
}
```

- Actors added to queue
- Processed one at a time
- 100ms delay between additions
- Only when AssetManager is idle

### 2. **Animation State Changes**

**Problem**: Changing animation state uses GL context, conflicts with rendering.

**Solution**: Deferred animation updates
```typescript
// During update: Mark dirty
QueueAnimationStateChange() {
    this.dirtyAnimation = true
}

// After rendering: Apply change
postProcessActors() {
    for (let actor of actors) {
        actor.ProcessAnimationStateChange()
    }
}
```

### 3. **Text Measurement Caching**

**Problem**: `measureText()` is expensive, called every frame for name tags.

**Solution**: Cache name tag dimensions
```typescript
if (this.nameTagSize == null) {
    let data = ctx.measureText(text)
    this.nameTagSize = {width: data.width, height: h}
}
```

Only measured once per actor, reused every frame.

### 4. **Z-Order Sorting**

**Problem**: Sorting all actors every frame is expensive.

**Solution**: 
- Stable sort using creation timestamp
- Only re-sort when actors added/removed
- Actors rarely change Z position

### 5. **Excessive Chibi Mitigations**

When `excessiveChibiMitigations` enabled:
- **Periodic stopping**: Reduces visual noise
- **Spread name tags**: Reduces overlap
- Helps performance by reducing rendering complexity

---

## Common Development Tasks

### **Adding a New Action Type**

1. Define action class in [Action.ts](src/player/Action.ts):
```typescript
export class MyNewAction implements ActorAction {
    constructor(actionData: any) {
        this.actionData = actionData
    }
    
    SetAnimation(actor: Actor, animation: string, viewport: BoundingBox) {
        setActorYPositionByAnimation(actor, animation, viewport)
    }
    
    GetAnimations(): string[] {
        return [this.actionData["animation"]]
    }
    
    UpdatePhysics(actor: Actor, deltaSecs: number, viewport: BoundingBox, player: SpinePlayer) {
        // Your movement logic here
    }
    
    DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox) {
        // Optional debug visualization
    }
}
```

2. Add to action factory:
```typescript
export function ParseActionNameToAction(actionName: string, actionData: any, flags: Map<string, any>): ActorAction {
    switch(actionName) {
        case "MY_NEW_ACTION":
            return new MyNewAction(actionData)
        // ... other cases
    }
}
```

3. Update Go backend to send new action type in WebSocket messages.

### **Adding a New WebSocket Message Type**

1. Add handler in [runtime.ts](src/stym/runtime.ts):
```typescript
messageHandler(event: MessageEvent) {
    let requestData = JSON.parse(event.data)
    if (requestData["type_name"] == "MY_NEW_MESSAGE") {
        this.handleMyNewMessage(requestData)
    }
    // ... other cases
}

handleMyNewMessage(requestData: any) {
    // Your logic here
}
```

2. Update Go backend to send new message type.

### **Customizing Actor Rendering**

Override `Draw()` method in [Actor.ts](src/player/Actor.ts):

```typescript
public Draw(sceneRenderer: SceneRenderer) {
    // Call parent rendering
    if (this.isSpritesheetActor()) {
        // ... spritesheet rendering
    } else {
        sceneRenderer.drawSkeleton(this.skeleton, this.config.premultipliedAlpha)
    }
    
    // Add custom rendering
    if (this.ShouldFlashCharacter) {
        let bb = this.GetBoundingBox()
        sceneRenderer.rect(
            false, // filled
            bb.x, bb.y, this.getPositionZ(),
            bb.width, bb.height,
            Color.YELLOW
        )
    }
}
```

### **Adding Debug Visualizations**

Implement `DrawDebug()` in action classes:

```typescript
DrawDebug(actor: Actor, renderer: SceneRenderer, viewport: BoundingBox) {
    // Draw target position
    renderer.circle(
        false, // filled
        this.endPosition.x,
        this.endPosition.y + viewport.height * 0.5,
        this.endPosition.z,
        10, // radius
        Color.RED
    )
    
    // Draw velocity vector
    let pos = actor.getPosition3()
    let vel = actor.getVelocity3()
    renderer.line(
        pos.x, pos.y, pos.x + vel.x * 10, pos.y + vel.y * 10,
        pos.z,
        Color.GREEN
    )
}
```

Enable with `?debug=true` URL parameter.

### **Modifying Chat Bubble Appearance**

Edit [SpeechBubble.ts](src/player/SpeechBubble.ts):

```typescript
private _drawSpeechBubble(ctx: CanvasRenderingContext2D, width: number, height: number) {
    // Change colors
    ctx.fillStyle = "darkblue"  // Background color
    ctx.strokeStyle = "cyan"    // Border color
    ctx.lineWidth = 3           // Border width
    
    // Change shape
    ctx.roundRect(-width/2 - pad, pad, width + 2*pad, -height - 2*pad, 10)  // Radius
    
    // ... rest of drawing
}
```

---

## Debugging Tips

### **Enable Debug Mode**

Add `?debug=true` to URL:
```
https://akchibibot.stymphalian.top/room/?channelName=mychannel&debug=true
```

Shows:
- Actor bounding boxes
- Movement targets
- Velocity vectors
- Z-order visualization

### **Access Runtime from Console**

```javascript
// Get runtime
window.SpineRuntime

// Get player
window.SpineRuntime.spinePlayer

// List actors
window.SpineRuntime.spinePlayer.getActorNames()

// Get specific actor
let actor = window.SpineRuntime.spinePlayer.getActor("stymphalian2__")

// Check actor state
actor.getPosition3()
actor.getVelocity3()
actor.loaded
actor.currentAction
```

### **Common Issues**

**Actor not appearing**:
1. Check `actor.loaded` - false means assets still loading
2. Check `actor.load_perma_failed` - true means load failed
3. Check browser console for asset errors
4. Verify asset URLs in WebSocket message

**Actor not moving**:
1. Check `actor.velocity` - should be non-zero
2. Check `actor.currentAction` - verify correct action type
3. Check `actor.paused` - should be false
4. Check `actor.speed` - should be > 0

**Chat messages not showing**:
1. Verify `showChatMessages` config is true
2. Check `actor.messageQueue.HasMessages()`
3. Check `actor.GetChatMessages()` - should return MessageBlock
4. Verify 2D canvas overlay is visible

**WebSocket disconnecting**:
1. Check backend logs for connection errors
2. Verify channel name is valid (alphanumeric only)
3. Check browser console for WebSocket errors
4. Monitor backoff timer in `Runtime.backoffTimeMsec`

---

## File Organization

```
static/spine/src/
├── index.ts                    # Entry point
├── main.css                    # Global styles
│
├── stym/                       # Runtime system
│   ├── runtime.ts              # WebSocket handler, message router
│   ├── utils.ts                # Config parsing, validation
│   ├── canvas_recorder.ts      # (Unused) Canvas recording
│   ├── control_cam.ts          # (Unused) Camera controls
│   └── gauss_jordan.ts         # (Unused) Matrix math
│
├── player/                     # Rendering system
│   ├── Player.ts               # Main orchestrator (SpinePlayer)
│   ├── Actor.ts                # Individual chibi character
│   ├── Action.ts               # Movement behaviors
│   ├── ChatMessages.ts         # Message queue system
│   ├── SpeechBubble.ts         # Text rendering
│   ├── Spritesheet.ts          # Spritesheet animation
│   ├── PerformancePanel.ts     # FPS and GPU timing display
│   ├── Utils.ts                # Camera, validation, DOM utils
│   ├── Flags.ts                # Feature flags
│   └── OffscreenRender.ts      # Offscreen rendering buffer
│
├── core/                       # Spine runtime core
│   ├── Skeleton.ts             # Skeleton data structure
│   ├── AnimationState.ts       # Animation controller
│   ├── AnimationStateData.ts   # Animation transitions
│   ├── SkeletonJson.ts         # JSON skeleton loader
│   ├── SkeletonBinary.ts       # Binary skeleton loader
│   ├── AtlasAttachmentLoader.ts # Texture atlas loader
│   └── ...                     # Other Spine classes
│
├── webgl/                      # WebGL rendering engine
│   ├── SceneRenderer.ts        # Main renderer
│   ├── Camera.ts               # Camera system
│   ├── AssetManager.ts         # Asset loading
│   ├── Input.ts                # Mouse/touch input
│   ├── LoadingScreen.ts        # Loading UI
│   ├── Vector3.ts              # 3D vector math
│   ├── Matrix4.ts              # 4x4 matrices
│   ├── WebGL.ts                # WebGL context management
│   └── ...                     # Other WebGL utilities
│
└── templates/                  # (Unused) HTML templates
```

---

## Dependencies

**NPM Packages** (from [package.json](package.json)):
```json
{
  "webpack": "^5.96.1",
  "webpack-cli": "^6.0.1",
  "typescript": "^5.7.2",
  "ts-loader": "^9.5.1",
  "css-loader": "^7.1.2",
  "style-loader": "^4.0.0"
}
```

**External Assets**:
- Spine Runtime 3.8 (embedded in `src/core`)
- Lato font (loaded from `/public/fonts/Lato/Lato-Black.ttf`)
- Spine assets (`.skel`, `.atlas`, `.png`) served by Go backend

---

## Build System

**Webpack Configuration** ([webpack.config.js](webpack.config.js)):

```javascript
module.exports = {
  entry: './src/index.ts',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist')
  },
  module: {
    rules: [
      {test: /\.tsx?$/, use: 'ts-loader'},
      {test: /\.css$/, use: ['style-loader', 'css-loader']}
    ]
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js']
  }
}
```

**Build Commands**:
```bash
# Development build (watch mode)
npm run watch

# Production build
npm run build

# Output: dist/bundle.js
```

**TypeScript Config** ([tsconfig.json](tsconfig.json)):
```json
{
  "compilerOptions": {
    "target": "ES6",
    "module": "commonjs",
    "lib": ["ES6", "DOM"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

---

## Integration with Backend

The Spine app connects to Go backend at:
```
WebSocket: wss://host/ws/?channelName={channelName}
Assets: https://host/assets/{type}/{filename}
```

**Asset URLs** in WebSocket messages:
```json
{
  "skel_file": "/assets/characters/char_002_amiya.skel",
  "atlas_file": "/assets/characters/char_002_amiya.atlas",
  "spritesheet_data_filepath": "/assets/custom/bubu_spritesheet.json"
}
```

The Go backend:
1. Serves static files from `static/spine/dist/`
2. Serves assets from `static/assets/`
3. Manages WebSocket connections per channel
4. Broadcasts messages to all connected clients for a channel

---

## Performance Monitoring

### **PerformancePanel Class**

The application includes an optional **PerformancePanel** that displays real-time performance metrics, including FPS and GPU frame timing.

**Implementation**: [PerformancePanel.ts](src/player/PerformancePanel.ts)

**Activation**: Add `?fps=1` to the URL:
```
https://akchibibot.stymphalian.top/room/?channelName=mychannel&fps=1
```

**Features**:
- **FPS (Frames Per Second)**: CPU frame rate averaged over last 60 frames
- **GPU Frame Time**: Actual GPU rendering time in milliseconds using `EXT_disjoint_timer_query_webgl2`
- **Color-coded display**:
  - 🟢 **Green**: Excellent performance (FPS ≥50, GPU ≤8ms)
  - 🟡 **Yellow**: Acceptable performance (FPS 30-49, GPU 8-16ms)
  - 🔴 **Red**: Poor performance (FPS <30, GPU >16ms)
- **Extensible**: Can display any custom metrics via `addMetric()` method
- **Semi-transparent background**: Ensures readability over any content
- **Configurable position**: top-left, top-right, bottom-left, bottom-right

---

### **Architecture**

**Class Structure**:
```typescript
export class PerformancePanel {
    // Configuration
    private config: PerformancePanelConfig
    private metrics: MetricConfig[]
    
    // FPS tracking
    private fpsFrameTimes: number[]
    private fpsCurrentFPS: number
    
    // GPU timing (WebGL2)
    private gl: WebGL2RenderingContext
    private timerExt: EXT_disjoint_timer_query_webgl2
    private queries: WebGLQuery[]
    private gpuFrameTime: number
    private gpuTimingEnabled: boolean
    
    // Methods
    initGPUTiming(gl): void
    beginGPUTiming(): void
    endGPUTiming(): void
    updateFPS(currentTime): void
    addMetric(metric: MetricConfig): void
    draw(ctx, width, height): void
}
```

**Metric Configuration**:
```typescript
interface MetricConfig {
    label: string                        // Display name (e.g., "FPS", "GPU")
    getValue: () => number               // Function to get current value
    format: (value: number) => string    // Format value for display
    getColor: (value: number) => string  // Color based on value
}
```

---

### **GPU Timing**

The panel uses **EXT_disjoint_timer_query_webgl2** to measure actual GPU rendering time.

**How it works**:

1. **Extension Check**:
```typescript
const ext = gl.getExtension('EXT_disjoint_timer_query_webgl2')
if (!ext) {
    console.warn('EXT_disjoint_timer_query_webgl2 not supported')
    return
}
```

2. **Query Creation** (double buffering):
```typescript
for (let i = 0; i < 2; i++) {
    const query = gl.createQuery()
    this.queries.push(query)
}
```

3. **Timing Measurement**:
```typescript
// At start of frame
gl.beginQuery(ext.TIME_ELAPSED_EXT, query)

// ... rendering happens ...

// At end of frame
gl.endQuery(ext.TIME_ELAPSED_EXT)

// Next frame: read result
const timeElapsed = gl.getQueryParameter(query, gl.QUERY_RESULT)
this.gpuFrameTime = timeElapsed / 1000000  // Convert ns to ms
```

4. **Integration in Player.ts**:
```typescript
// Initialize in setupDom()
if (this.playerConfig.showFPS) {
    this.performancePanel = new PerformancePanel()
    this.performancePanel.initGPUTiming(this.context.gl)
}

// In rendering loop
updateActors() {
    this.performancePanel?.beginGPUTiming()  // Start GPU timer
    // ... update actors ...
}

drawActors() {
    // ... draw actors ...
    this.performancePanel?.endGPUTiming()    // End GPU timer
}
```

**Browser Support**:
- Chrome/Edge: Full support
- Firefox: Full support
- Safari: May not be available (falls back to FPS only)

**Fallback**: If extension not available, only FPS is displayed.

---

### **Adding Custom Metrics**

The PerformancePanel is designed to be extensible:

```typescript
// Example: Add actor count metric
performancePanel.addMetric({
    label: 'Actors',
    getValue: () => this.actors.size,
    format: (value) => value.toString(),
    getColor: (value) => {
        if (value < 10) return '#00ff00'      // Green
        if (value < 50) return '#ffff00'      // Yellow
        return '#ff0000'                       // Red
    }
})

// Example: Add memory usage metric
performancePanel.addMetric({
    label: 'Memory',
    getValue: () => {
        if (performance.memory) {
            return performance.memory.usedJSHeapSize / 1048576  // MB
        }
        return 0
    },
    format: (value) => `${value.toFixed(1)}MB`,
    getColor: (value) => {
        if (value < 100) return '#00ff00'
        if (value < 200) return '#ffff00'
        return '#ff0000'
    }
})
```

---

### **Performance Impact**

The PerformancePanel has minimal overhead:

**FPS Tracking**: ~0.01ms per frame
- Simple array operations (push/shift)
- Basic arithmetic (averaging)

**GPU Timing**: ~0.02ms per frame
- Query creation is one-time cost
- Double buffering prevents stalls
- Reading results is asynchronous

**Drawing**: ~0.1-0.2ms per frame
- 2D canvas text rendering
- Only draws 2-3 lines of text
- Cached text measurements

**Total Overhead**: <0.5ms per frame (imperceptible at 60 FPS)

---

### **Code Pointers**

**PerformancePanel Class**:
- [PerformancePanel.ts:1-290](src/player/PerformancePanel.ts): Full implementation
- [PerformancePanel.ts:75-98](src/player/PerformancePanel.ts): `initGPUTiming()` - WebGL2 setup
- [PerformancePanel.ts:103-121](src/player/PerformancePanel.ts): `beginGPUTiming()` - Start measurement
- [PerformancePanel.ts:126-136](src/player/PerformancePanel.ts): `endGPUTiming()` - End measurement
- [PerformancePanel.ts:148-166](src/player/PerformancePanel.ts): `updateFPS()` - FPS calculation
- [PerformancePanel.ts:183-245](src/player/PerformancePanel.ts): `draw()` - Rendering

**Integration**:
- [Player.ts:214](src/player/Player.ts): PerformancePanel property
- [Player.ts:246](src/player/Player.ts): showFPS config validation
- [Player.ts:448-456](src/player/Player.ts): PerformancePanel initialization
- [Player.ts:705-709](src/player/Player.ts): GPU timing begin
- [Player.ts:743-747](src/player/Player.ts): GPU timing end
- [Player.ts:764-770](src/player/Player.ts): Panel drawing
- [Player.ts:836-839](src/player/Player.ts): FPS update

**Configuration**:
- [runtime.ts:6-14](src/stym/runtime.ts): RuntimeConfig with showFPS
- [runtime.ts:68](src/stym/runtime.ts): Pass showFPS to SpinePlayer
- [utils.ts:56](src/stym/utils.ts): Parse `?fps=1` URL parameter

---

## Summary

The Spine application is a high-performance WebGL renderer for animated chibi characters. Key architectural decisions:

1. **Separation of Concerns**: Runtime (WebSocket) → Player (Rendering) → Actor (Individual chibi)
2. **Deferred Updates**: Actor queue and animation state changes to avoid blocking
3. **Extensible Actions**: Pluggable movement behaviors via `ActorAction` interface
4. **Dual Rendering**: Spine skeletal + spritesheet support
5. **Performance Monitoring**: Optional PerformancePanel with CPU/GPU metrics
6. **Performance**: Asset caching, text measurement caching, Z-order sorting optimization
7. **Debug Support**: Console access to runtime, debug visualizations

**For Junior Engineers**:
- Start with [index.ts](src/index.ts) to understand entry point
- Read [runtime.ts](src/stym/runtime.ts) to understand WebSocket protocol
- Study [Actor.ts](src/player/Actor.ts) for individual chibi logic
- Explore [Action.ts](src/player/Action.ts) to add new movement behaviors
- Use `?debug=true` and `window.SpineRuntime` for debugging

**Common Changes**:
- Add new action → Edit [Action.ts](src/player/Action.ts)
- Customize rendering → Edit [Actor.ts](src/player/Actor.ts) `Draw()`
- Change chat bubbles → Edit [SpeechBubble.ts](src/player/SpeechBubble.ts)
- Add WebSocket message → Edit [runtime.ts](src/stym/runtime.ts) `messageHandler()`

The codebase is well-structured for extension. The action system, in particular, makes adding new behaviors straightforward without modifying core rendering logic.
