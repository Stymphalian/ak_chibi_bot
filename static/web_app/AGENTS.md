# Web App Architecture Guide (AGENTS.md)

## Overview

The `static/web_app` directory contains a **React + TypeScript** single-page application (SPA) that serves as the control panel and documentation interface for the AK Chibi Bot. This web app allows streamers to:
- View getting started documentation
- Authenticate via Twitch OAuth
- Configure room settings (animation speed, sprite size, blacklists, etc.)
- Access admin tools for monitoring active rooms and connections

The actual chibi rendering and Twitch chat integration happens in a separate `static/spine` application (covered in the Spine Integration section).

---

## Technology Stack

- **Framework**: React 18.3 with TypeScript
- **Routing**: React Router v6 (client-side routing)
- **UI Framework**: React Bootstrap 5.3
- **Form Management**: React Hook Form with Yup validation
- **HTTP Client**: Axios + Fetch API
- **Build Tool**: Create React App (webpack under the hood)
- **State Management**: React Context API (for authentication)

---

## Project Structure

```
static/web_app/
├── public/                    # Static assets
│   ├── index.html            # HTML template (entry point)
│   ├── favicon.ico           # Site favicon
│   └── manifest.json         # PWA manifest
├── src/
│   ├── index.tsx             # Application entry point, router setup
│   ├── index.css             # Global styles
│   ├── error-page.tsx        # Error boundary page
│   │
│   ├── pages/                # Page-level components (route handlers)
│   │   ├── Layout.tsx        # Root layout with navbar and footer
│   │   ├── Home.tsx          # Landing page with quick start guide
│   │   ├── Docs.tsx          # Command documentation
│   │   ├── Settings.tsx      # Room settings page (authenticated)
│   │   ├── Admin.tsx         # Admin dashboard (admin only)
│   │   ├── Login.tsx         # Twitch login page
│   │   └── LoginCallback.tsx # OAuth callback handler
│   │
│   ├── components/           # Reusable UI components
│   │   ├── TopNavBar.tsx     # Navigation bar
│   │   ├── Footer.tsx        # Footer component
│   │   ├── Banner.tsx        # Header banner
│   │   ├── Code.tsx          # Inline code display component
│   │   ├── LoaderBlock.tsx   # Loading spinner wrapper
│   │   ├── RoomSettings.tsx  # Form for editing room settings
│   │   └── TwitchLoginButton.tsx  # Twitch OAuth button
│   │
│   ├── contexts/             # React Context providers
│   │   └── auth.tsx          # Authentication context and hooks
│   │
│   ├── api/                  # API client functions
│   │   ├── real.ts           # Room settings API calls
│   │   ├── admin.ts          # Admin API calls
│   │   └── utils.ts          # Validation utilities
│   │
│   ├── models/               # TypeScript type definitions
│   │   └── models.ts         # Data models (ChannelSettings, AdminInfo, etc.)
│   │
│   └── assets/               # Images and static resources
│       └── demo1.png         # Demo screenshot
│
├── package.json              # NPM dependencies and scripts
├── tsconfig.json            # TypeScript configuration
└── README.md                # Standard CRA readme
```

---

## Core Components

### 1. **Entry Point** (`src/index.tsx`)

Sets up the application router and wraps everything in the `AuthProvider` context.

**Key responsibilities:**
- Creates React Router with all route definitions
- Wraps app in `AuthProvider` for global auth state
- Defines protected routes using `RequireAuth` wrapper
- Renders root React tree into DOM

**Route structure:**
```
/ (Layout)
├── /              → HomePage
├── /docs          → DocsPage
├── /settings      → SettingsPage (requires auth)
├── /admin         → AdminPage (requires admin auth)
├── /login         → LoginPage
└── /login/callback → LoginCallbackPage
```

---

### 2. **Authentication System** (`src/contexts/auth.tsx`)

Provides global authentication state using React Context API.

**Components:**
- `AuthProvider`: Wraps the app and manages auth state
- `AuthContext`: Context object holding auth data
- `useAuth()`: Custom hook to access auth context
- `RequireAuth`: HOC to protect routes
- `AuthStatus`: UI component showing login/logout button

**Auth Flow:**
1. On mount, `AuthProvider` calls `/auth/check/` to verify session
2. User clicks "Login with Twitch" → redirects to `/auth/login/twitch/`
3. Twitch redirects to `/auth/callback` (handled by Go backend)
4. Backend sets session cookie and redirects to `/login/callback`
5. Frontend redirects user to originally requested page
6. Subsequent API calls use JWT tokens from `/auth/token/` endpoint

**Key methods:**
- `getAccessToken()`: Returns cached JWT or fetches new one
- `Login()`: Redirects to Twitch OAuth flow
- `Logout()`: Clears session and redirects
- `checkAuthenticated()`: Verifies current auth status

**State:**
```typescript
{
  isAuthenticated: boolean
  loading: boolean
  userName: string
  isAdmin: boolean
  accessToken: string (cached)
}
```

---

### 3. **Pages**

#### **Home.tsx**
- Landing page with quick start instructions
- Shows personalized OBS browser source URL if authenticated
- Displays demo image
- Includes disclaimers about asset ownership

#### **Docs.tsx**
- Comprehensive command reference table
- Lists all `!chibi` commands with descriptions
- Examples of command usage

#### **Settings.tsx**
- Protected route (requires authentication)
- Fetches user's channel settings via API
- Renders `RoomSettingsForm` component
- Shows loading spinner while fetching data

#### **Admin.tsx**
- Protected route (requires admin role)
- Real-time dashboard of all active rooms
- Shows:
  - General metrics (GC times, system info)
  - Per-room information (channel, chatters, websocket connections)
  - Average FPS per connection
  - Actions: kick users, remove rooms
- Auto-refreshes data periodically

#### **Layout.tsx**
- Root layout component for all pages
- Renders `TopNavBar`, `<Outlet/>` (child routes), and `Footer`
- Provides consistent structure across all pages

#### **Login.tsx & LoginCallback.tsx**
- `Login.tsx`: Shows Twitch login button
- `LoginCallback.tsx`: Handles OAuth redirect, then redirects to app

---

### 4. **API Layer**

All API calls go through typed functions in the `api/` directory.

#### **real.ts** - Room Settings API
```typescript
getUserChannelSettings(accessToken, channelName)
  GET /api/rooms/settings/?channel_name={channelName}
  → Returns ChannelSettings

updateUserChannelSettings(accessToken, channelName, updates)
  POST /api/rooms/settings/
  → Updates settings and returns new values
```

#### **admin.ts** - Admin API
```typescript
getAdminInfo(accessToken)
  GET /api/admin/info/
  → Returns AdminInfo (all rooms, metrics, etc.)

removeRoom(accessToken, channelName)
  POST /api/rooms/remove/
  → Deletes room

removeUserFromRoom(accessToken, channelName, userName)
  POST /api/rooms/users/remove/
  → Kicks user from room
```

#### **utils.ts** - Validation
```typescript
validateChannelName(channel: string): boolean
  // Validates alphanumeric + underscores/hyphens, max 100 chars

validateTwitchUserName(username: string): boolean
  // Same validation as channel names
```

---

### 5. **Components**

#### **RoomSettings.tsx**
- Complex form using `react-hook-form` + `yup` validation
- Allows editing of:
  - Min/max animation speed
  - Min/max movement velocity
  - Min/max sprite scale
  - Max sprite pixel size
  - Username blacklist (comma-separated)
- Shows validation errors inline
- Displays success/error alerts after submission

#### **TopNavBar.tsx**
- Navigation bar with links to Home, Docs, Settings
- Shows Admin link if user is admin
- Displays `AuthStatus` component (login/logout)
- Uses `react-router-dom`'s `NavLink` for active state

#### **LoaderBlock.tsx**
- Wrapper component that shows spinner while `loading` prop is true
- Renders children when loading completes

#### **Code.tsx**
- Styled inline code component
- Wraps text in `<code>` tag with custom CSS

#### **TwitchLoginButton.tsx**
- Button that redirects to Twitch OAuth flow
- Accepts `redirect_to` prop to return to original page after login

---

## Data Flow

### **Settings Update Flow**
```
User fills form
    ↓
RoomSettings.tsx validates with Yup
    ↓
updateUserChannelSettings(accessToken, channelName, data)
    ↓
POST /api/rooms/settings/ (Go backend)
    ↓
Backend updates Postgres DB
    ↓
Returns updated settings
    ↓
Form shows success alert
```

### **Authentication Flow**
```
Page loads
    ↓
AuthProvider calls /auth/check/
    ↓
Sets isAuthenticated, userName, isAdmin
    ↓
User navigates to /settings
    ↓
RequireAuth checks auth.isAuthenticated
    ↓
If false: redirect to /login
If true: render SettingsPage
    ↓
SettingsPage calls auth.getAccessToken()
    ↓
If token expired: fetch new from /auth/token/
    ↓
Use token in Authorization header for API calls
```

### **Admin Dashboard Flow**
```
AdminPage mounts
    ↓
Calls getAdminInfo(accessToken)
    ↓
GET /api/admin/info/ returns:
  - rooms: Array<AdminRoomInfo>
  - metrics: Map<string, any>
  - next_gc_time: string
    ↓
Render tables with room data
    ↓
User clicks "Kick" on a chatter
    ↓
removeUserFromRoom(accessToken, roomId, userName)
    ↓
POST /api/rooms/users/remove/
    ↓
Refresh data via getAdminInfo()
    ↓
UI updates
```

---

## Key Data Models

### **ChannelSettings**
```typescript
type ChannelSettings = {
  channelName: string
  minAnimationSpeed: number
  maxAnimationSpeed: number
  minVelocity: number
  maxVelocity: number
  minSpriteScale: number
  maxSpriteScale: number
  maxSpritePixelSize: number
  usernamesBlacklist?: string  // comma-separated
}
```

### **AdminInfo**
```typescript
type AdminInfo = {
  rooms: AdminRoomInfo[]
  next_gc_time: string
  metrics: Map<string, any>
}

type AdminRoomInfo = {
  channel_name: string
  created_at: string
  last_time_used: string
  chatters: AdminChatterInfo[]
  next_gc_time: string
  num_websocket_connections: number
  connection_average_fps: Map<string, number>
}

type AdminChatterInfo = {
  username: string
  operator: string
  last_chat_time: string
}
```

---

## Spine Integration

The web app is **separate** from the actual chibi rendering system. The rendering happens in `static/spine`, which is a completely different application.

### **Spine App Structure** (`static/spine/`)

```
static/spine/
├── src/
│   ├── index.ts              # Entry point, creates Runtime
│   ├── stym/
│   │   ├── runtime.ts        # Main runtime controller
│   │   └── utils.ts          # Utility functions
│   ├── player/
│   │   ├── Player.ts         # SpinePlayer (WebGL renderer)
│   │   ├── Actor.ts          # Individual chibi actor
│   │   ├── ChatMessages.ts   # Chat bubble system
│   │   └── Utils.ts          # Player utilities
│   ├── webgl/                # WebGL rendering engine
│   └── core/                 # Spine runtime core
└── webpack.config.js
```

### **How Spine Works**

1. **Entry Point** (`index.ts`):
   - Parses `?channelName=` from URL
   - Creates `Runtime` instance
   - Runtime opens WebSocket connection to Go backend

2. **WebSocket Protocol**:
   ```
   WS /ws/?channelName={channelName}
   ```
   
   **Message Types:**
   - `SET_OPERATOR`: Add/update chibi for user
   - `REMOVE_OPERATOR`: Remove user's chibi
   - `SHOW_CHAT_MESSAGE`: Display chat bubble
   - `FIND_OPERATOR`: Highlight user's chibi

3. **Rendering Flow**:
   ```
   Twitch chat: !chibi Amiya
        ↓
   Go backend processes command
        ↓
   Sends WebSocket message to room
        ↓
   Runtime.messageHandler() receives message
        ↓
   Runtime.swapCharacter() called
        ↓
   SpinePlayer creates/updates Actor
        ↓
   Actor loads Spine skeleton data
        ↓
   WebGL renders chibi on canvas
        ↓
   Animation loop updates all actors
   ```

4. **Key Classes**:
   - **Runtime**: WebSocket handler, message router
   - **SpinePlayer**: WebGL canvas manager, actor registry
   - **Actor**: Individual chibi (skeleton, animations, position)
   - **ChatMessages**: Speech bubble renderer

### **Integration Points**

- **Web App** (`/settings`): Updates room config in DB
- **Go Backend**: Reads config, enforces limits, sends to spine
- **Spine App** (`/room`): Receives WebSocket messages, renders chibis

**Example URL:**
```
https://akchibibot.stymphalian.top/room/?channelName=stymphalian2__
```

This loads the Spine app, which connects to:
```
wss://akchibibot.stymphalian.top/ws/?channelName=stymphalian2__
```

---

## Running the Web App

### **Development**

```bash
cd static/web_app

# Install dependencies
npm install

# Start dev server (runs on localhost:3000)
npm start

# Run in watch mode (auto-rebuild)
npm run watch
```

The dev server proxies API requests to the Go backend (configure in `package.json` if needed).

### **Production Build**

```bash
cd static/web_app

# Build optimized bundle
npm run build

# Output goes to static/web_app/build/
# Go backend serves these static files
```

### **Testing**

```bash
cd static/web_app

# Run tests
npm test
```

---

## API Endpoints Reference

### **Authentication**
- `GET /auth/check/` - Check if user is authenticated
- `GET /auth/login/twitch/` - Initiate Twitch OAuth flow
- `POST /auth/logout/` - Logout user
- `GET /auth/token/` - Get JWT access token

### **Room Settings**
- `GET /api/rooms/settings/?channel_name={name}` - Get settings
- `POST /api/rooms/settings/` - Update settings

### **Admin**
- `GET /api/admin/info/` - Get all rooms and metrics
- `POST /api/rooms/remove/` - Delete a room
- `POST /api/rooms/users/remove/` - Kick user from room

---

## Common Development Tasks

### **Adding a New Page**

1. Create component in `src/pages/NewPage.tsx`
2. Add route in `src/index.tsx`:
   ```tsx
   <Route path="/newpage" element={<NewPage />} />
   ```
3. Add nav link in `src/components/TopNavBar.tsx`

### **Adding a New API Call**

1. Define TypeScript types in `src/models/models.ts`
2. Create API function in `src/api/real.ts`:
   ```typescript
   export async function getNewData(accessToken: string) {
     const response = await fetch('/api/new-endpoint', {
       headers: { 'Authorization': `Bearer ${accessToken}` }
     });
     return await response.json();
   }
   ```
3. Use in component with `useAuth()`:
   ```tsx
   const auth = useAuth();
   const token = await auth.getAccessToken();
   const data = await getNewData(token);
   ```

### **Adding Form Validation**

Use Yup schema in `RoomSettings.tsx` as example:
```typescript
const schema = yup.object({
  fieldName: yup.string().required(),
  numberField: yup.number().min(0).max(100).required()
});

const { register, handleSubmit, formState } = useForm({
  resolver: yupResolver(schema)
});
```

---

## Security Considerations

1. **Authentication**: Session cookies + JWT tokens
2. **Authorization**: Backend validates admin status
3. **CSRF Protection**: Handled by Go backend
4. **Input Validation**: 
   - Client-side: Yup schemas
   - Server-side: Go validation (primary defense)
5. **Channel Name Validation**: Alphanumeric only, max 100 chars
6. **WebSocket Auth**: Channel name in URL (not sensitive)

---

## Performance Notes

- **Code Splitting**: React Router handles lazy loading
- **Build Optimization**: CRA uses webpack with minification
- **API Caching**: JWT tokens cached to reduce `/auth/token/` calls
- **Admin Dashboard**: Auto-refresh can be expensive (consider rate limiting)

---

## Troubleshooting

### **"401 Unauthorized" errors**
- Check if user is logged in: `auth.isAuthenticated`
- Verify JWT token: `auth.getAccessToken()`
- Check backend logs for session issues

### **API calls fail in development**
- Ensure Go backend is running
- Check proxy configuration in `package.json`
- Verify CORS settings in Go backend

### **Form validation not working**
- Check Yup schema matches form fields
- Look for `formState.errors` in console
- Ensure `resolver: yupResolver(schema)` is set

---

## Additional Resources

- **React Router Docs**: https://reactrouter.com/
- **React Hook Form**: https://react-hook-form.com/
- **Yup Validation**: https://github.com/jquense/yup
- **React Bootstrap**: https://react-bootstrap.github.io/

---

## Summary

The `static/web_app` is a modern React SPA that provides:
- **User Interface** for streamers to configure their bot
- **Authentication** via Twitch OAuth
- **API Integration** with Go backend
- **Admin Tools** for monitoring active rooms

The actual chibi rendering happens in `static/spine`, which is a separate TypeScript WebGL application that receives commands via WebSocket from the Go backend.

**Key Architectural Decisions:**
1. Separate web app from rendering engine (clean separation of concerns)
2. React Context for global auth state (simple, no Redux needed)
3. Form validation with Yup (type-safe, declarative)
4. JWT tokens for API auth (stateless, scalable)
5. Protected routes with `RequireAuth` HOC (security boundary)

This architecture allows independent development of the control panel and the rendering system, while maintaining a clean API boundary between them.
