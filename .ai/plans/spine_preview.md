# Spine Preview Page Implementation Plan

## Overview

Create a new page in the web app that allows users to drag and drop Spine skeleton files (.skel), texture files (.png), and atlas files (.atlas) to preview and test Spine animations in a SpinePlayer component.

## Goals

1. **File Upload Interface**: Drag-and-drop zone for Spine asset files
2. **Spine Integration**: Render uploaded assets using the existing Spine runtime
3. **Animation Controls**: Play/pause, animation selection, speed control
4. **Asset Management**: Handle multiple file uploads and validation
5. **Error Handling**: Clear feedback for invalid files or missing dependencies

---

## Technical Architecture

### File Structure
```
static/web_app/src/
├── pages/
│   └── SpinePreview.tsx          # New preview page
├── components/
│   ├── spine/                    # New spine-related components
│   │   ├── SpineViewer.tsx       # Main spine player component
│   │   ├── FileDropZone.tsx      # File upload interface
│   │   ├── AnimationControls.tsx # Play/pause/speed controls
│   │   └── AssetManager.tsx      # File management UI
│   └── ...
├── hooks/
│   ├── useSpineAssets.ts         # Asset management hook
│   └── useSpinePlayer.ts         # Spine player integration hook
├── utils/
│   └── spineUtils.ts             # File validation and helpers
└── types/
    └── spine.ts                  # Spine-related type definitions
```

---

## Implementation Plan

### Phase 1: Basic File Upload (1-2 days)

#### 1.1 Create File Drop Zone Component
```typescript
// components/spine/FileDropZone.tsx
interface FileDropZoneProps {
  onFilesUploaded: (files: SpineAssetFiles) => void;
  acceptedTypes: string[];
  maxFiles?: number;
}

interface SpineAssetFiles {
  skeleton?: File;  // .skel or .json
  atlas?: File;     // .atlas
  texture?: File;   // .png
}
```

**Features:**
- Drag-and-drop area with visual feedback
- File type validation (.skel, .json, .atlas, .png)
- Multiple file selection
- File size limits (e.g., 50MB total)
- Clear visual indicators for required vs optional files

#### 1.2 Asset Management Hook
```typescript
// hooks/useSpineAssets.ts
interface UseSpineAssetsReturn {
  assets: SpineAssetFiles;
  uploadFiles: (files: FileList) => void;
  clearAssets: () => void;
  isValidAssetSet: boolean;
  errors: string[];
}
```

**Responsibilities:**
- Store uploaded files in component state
- Validate file combinations (skeleton + atlas + texture)
- Convert files to URLs/ArrayBuffers for Spine player
- Handle file replacement logic

#### 1.3 File Validation Utilities
```typescript
// utils/spineUtils.ts
export function validateSpineFiles(files: SpineAssetFiles): ValidationResult;
export function getFileExtension(filename: string): string;
export function isValidSpineAsset(file: File): boolean;
export function createObjectURLs(files: SpineAssetFiles): SpineAssetURLs;
```

### Phase 2: Spine Player Integration (2-3 days)

#### 2.1 Spine Viewer Component
```typescript
// components/spine/SpineViewer.tsx
interface SpineViewerProps {
  assetUrls: SpineAssetURLs;
  width?: number;
  height?: number;
  onAnimationsLoaded?: (animations: string[]) => void;
  onError?: (error: string) => void;
}
```

**Integration Strategy:**
- Reuse existing SpinePlayer from `static/spine/src/player/Player.ts`
- Create wrapper component that manages player lifecycle
- Handle asset loading and error states
- Provide ref access for external controls

#### 2.2 Spine Player Hook
```typescript
// hooks/useSpinePlayer.ts
interface UseSpinePlayerReturn {
  playerRef: RefObject<SpinePlayer>;
  isLoading: boolean;
  animations: string[];
  currentAnimation: string;
  playAnimation: (name: string) => void;
  setAnimationSpeed: (speed: number) => void;
  play: () => void;
  pause: () => void;
}
```

#### 2.3 Handle Spine Runtime Dependencies
**Option A: Import from existing spine build**
```typescript
// Import from built spine.js bundle
import { SpinePlayer } from '/spine/dist/spine.js';
```

**Option B: Create shared spine module**
```typescript
// Create shared npm package or workspace
// Move spine classes to shared location
import { SpinePlayer } from '@shared/spine-runtime';
```

**Recommended: Option A** (simpler, faster implementation)

### Phase 3: Animation Controls (1 day)

#### 3.1 Animation Controls Component
```typescript
// components/spine/AnimationControls.tsx
interface AnimationControlsProps {
  animations: string[];
  currentAnimation: string;
  isPlaying: boolean;
  speed: number;
  onAnimationChange: (animation: string) => void;
  onPlayPause: () => void;
  onSpeedChange: (speed: number) => void;
  onRestart: () => void;
}
```

**Features:**
- Animation dropdown/list
- Play/pause button
- Speed slider (0.1x to 3.0x)
- Restart button
- Timeline scrubber (optional)

#### 3.2 Control Integration
- Connect controls to Spine player instance
- Real-time animation state updates
- Keyboard shortcuts (spacebar = play/pause)

### Phase 4: Main Page Integration (1 day)

#### 4.1 Create SpinePreview Page
```typescript
// pages/SpinePreview.tsx
export function SpinePreviewPage() {
  const [assets, setAssets] = useState<SpineAssetFiles>({});
  const spinePlayer = useSpinePlayer();
  
  return (
    <Container>
      <Row>
        <Col md={4}>
          <FileDropZone onFilesUploaded={setAssets} />
          <AssetManager assets={assets} />
        </Col>
        <Col md={8}>
          <SpineViewer assetUrls={createObjectURLs(assets)} />
          <AnimationControls {...spinePlayer} />
        </Col>
      </Row>
    </Container>
  );
}
```

#### 4.2 Add Route and Navigation
```typescript
// src/index.tsx - Add route
<Route path="/spine-preview" element={<SpinePreviewPage />} />

// components/TopNavBar.tsx - Add nav link
<NavLink className="nav-link" to="/spine-preview">Spine Preview</NavLink>
```

---

## Technical Challenges & Solutions

### 1. Spine Runtime Integration

**Challenge**: Web app and spine app are separate builds
**Solution**: 
- Load spine.js bundle as external script
- Use window.SpineRuntime or create bridge
- Alternative: Extract spine classes to shared module

### 2. File Loading and CORS

**Challenge**: Loading local files in browser with security restrictions
**Solution**:
- Use FileReader API to read files as ArrayBuffer
- Create object URLs for assets
- Handle cleanup of object URLs

### 3. Asset Dependencies

**Challenge**: Spine assets have dependencies (atlas references texture)
**Solution**:
- Parse atlas file to extract texture filename
- Validate texture filename matches uploaded PNG
- Provide clear error messages for missing dependencies

### 4. Memory Management

**Challenge**: Large texture files and potential memory leaks
**Solution**:
- Implement file size limits
- Clean up object URLs on component unmount
- Dispose spine resources properly

---

## User Experience Flow

### Happy Path
1. User navigates to `/spine-preview`
2. Sees empty drop zone with instructions
3. Drags 3 files (.skel, .atlas, .png) onto zone
4. Files validate successfully, preview loads
5. User selects animation from dropdown
6. Animation plays automatically
7. User adjusts speed, play/pause as needed

### Error Handling
1. **Missing files**: Show which files are still needed
2. **Invalid format**: Clear error message with expected formats
3. **File size too large**: Warning with size limits
4. **Loading errors**: Spine-specific error messages
5. **Dependency mismatch**: Atlas/texture filename conflicts

---

## File Validation Rules

### Required Files
- **Skeleton**: `.skel` (binary) or `.json` (text)
- **Atlas**: `.atlas` file
- **Texture**: `.png` file

### Validation Logic
```typescript
function validateAssets(files: SpineAssetFiles): ValidationResult {
  const errors: string[] = [];
  
  // Check required files
  if (!files.skeleton) errors.push("Skeleton file required (.skel or .json)");
  if (!files.atlas) errors.push("Atlas file required (.atlas)");
  if (!files.texture) errors.push("Texture file required (.png)");
  
  // Check file sizes
  if (files.texture && files.texture.size > 25 * 1024 * 1024) {
    errors.push("Texture file too large (max 25MB)");
  }
  
  // Parse atlas to validate texture reference
  if (files.atlas && files.texture) {
    const atlasText = await files.atlas.text();
    const textureRef = parseAtlasTextureName(atlasText);
    const textureName = files.texture.name;
    
    if (textureRef !== textureName) {
      errors.push(`Atlas references '${textureRef}' but uploaded '${textureName}'`);
    }
  }
  
  return { isValid: errors.length === 0, errors };
}
```

---

## UI/UX Design

### Layout
```
┌─────────────────────────────────────────────────────────┐
│                    Spine Preview                        │
├─────────────────┬───────────────────────────────────────┤
│ File Upload     │                                       │
│ ┌─────────────┐ │            Spine Canvas               │
│ │ Drop files  │ │         ┌─────────────────┐          │
│ │ here        │ │         │                 │          │
│ │             │ │         │   [Animation]   │          │
│ └─────────────┘ │         │                 │          │
│                 │         └─────────────────┘          │
│ Uploaded Files: │                                       │
│ ✓ amiya.skel    │         Animation Controls            │
│ ✓ amiya.atlas   │    [Idle ▼] [⏸️] [🔄] [──●───]       │
│ ✓ amiya.png     │    Speed: [1.0x]                     │
│                 │                                       │
│ [Clear All]     │                                       │
└─────────────────┴───────────────────────────────────────┘
```

### Visual Feedback
- **Drop zone**: Highlight border when dragging over
- **File status**: Checkmarks for uploaded, warnings for errors
- **Loading states**: Spinners during asset loading
- **Error states**: Red text with clear instructions

---

## Implementation Checklist

### Phase 1: File Upload
- [ ] Create `FileDropZone` component with drag-and-drop
- [ ] Implement file validation utilities
- [ ] Create `useSpineAssets` hook
- [ ] Add file size and type validation
- [ ] Handle multiple file upload scenarios

### Phase 2: Spine Integration
- [ ] Research spine.js bundle integration options
- [ ] Create `SpineViewer` component wrapper
- [ ] Implement asset loading from File objects
- [ ] Handle spine player lifecycle (create/destroy)
- [ ] Add error handling for spine loading failures

### Phase 3: Controls
- [ ] Create `AnimationControls` component
- [ ] Implement play/pause functionality
- [ ] Add animation selection dropdown
- [ ] Implement speed control slider
- [ ] Add keyboard shortcuts

### Phase 4: Page Integration
- [ ] Create `SpinePreviewPage` component
- [ ] Add route to router configuration
- [ ] Add navigation link to TopNavBar
- [ ] Style page layout with Bootstrap
- [ ] Add responsive design considerations

### Phase 5: Polish
- [ ] Add comprehensive error messages
- [ ] Implement memory cleanup
- [ ] Add loading states and transitions
- [ ] Write unit tests for file validation
- [ ] Add accessibility features (ARIA labels, keyboard nav)

---

## Security Considerations

1. **File Validation**: Strict file type checking
2. **Size Limits**: Prevent large file uploads
3. **Sanitization**: Validate atlas file content
4. **Memory Limits**: Clean up resources properly
5. **No Server Upload**: Files stay in browser only

---

## Testing Strategy

### Unit Tests
- File validation functions
- Asset management hook logic
- Spine utility functions

### Integration Tests
- File upload flow
- Spine player integration
- Animation control interactions

### Manual Testing
- Various file combinations
- Error scenarios
- Different browsers
- Mobile responsiveness

---

## Future Enhancements

### Phase 2 Features (Future)
1. **Asset Library**: Save/load asset combinations
2. **Animation Recording**: Export GIFs of animations
3. **Bone Inspector**: Show skeleton hierarchy
4. **Skin Switching**: Support multiple skins
5. **Background Options**: Custom backgrounds
6. **Share Links**: Generate shareable preview URLs
7. **Batch Upload**: ZIP file support with auto-detection

### Technical Improvements
1. **Web Workers**: Move file processing off main thread
2. **Caching**: LocalStorage for recently used assets
3. **Drag Sorting**: Reorder animations
4. **Timeline**: Frame-by-frame animation scrubbing
5. **Export Options**: Save modified animations

---

## Estimated Timeline

- **Phase 1 (File Upload)**: 2 days
- **Phase 2 (Spine Integration)**: 3 days  
- **Phase 3 (Controls)**: 1 day
- **Phase 4 (Page Integration)**: 1 day
- **Testing & Polish**: 1 day

**Total: ~8 days** (1.5-2 weeks for one developer)

---

## Success Criteria

1. **Functional**: Users can upload spine assets and see animations
2. **User-Friendly**: Clear instructions and error messages
3. **Performant**: Smooth animations, no memory leaks
4. **Robust**: Handles various file combinations and error cases
5. **Integrated**: Fits naturally into existing web app

This feature will significantly enhance the utility of the AK Chibi Bot by providing a testing ground for custom spine assets before deployment.