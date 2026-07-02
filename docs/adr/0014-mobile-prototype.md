# ADR 0014: Mobile Client Prototype Strategy

## Status

Accepted (design; native mobile implementation deferred to extension phase)

## Context

EchoLine's current frontend is a Vite/React web application. Mobile access is currently via a mobile browser, which has significant limitations:

1. **Push notifications**: Web Push is not reliable on iOS; background notifications require a native app.
2. **Background sync**: Browser tabs are suspended; native apps can maintain persistent WS or use APNs/FCM for wake-up.
3. **Media access**: Camera, microphone, and file picker are restricted or inconsistent in mobile browsers.
4. **Performance**: React SPA on mobile browsers has higher memory and battery cost than native rendering.
5. **App store distribution**: No discoverability via App Store / Google Play.

The question is: **React Native vs Flutter vs native iOS/Android**, and how to reuse the existing API surface.

## Decision

Use **React Native** for the mobile prototype, with a shared business logic layer.

### Rationale

| Factor | React Native | Flutter | Native (Swift/Kotlin) |
|--------|-------------|---------|----------------------|
| Code reuse with web | High (shared hooks, API layer) | Low (Dart) | Zero |
| Performance | Good (JS bridge; 60fps UI achievable) | Excellent (Skia renderer) | Best |
| Team skill | High (TypeScript/React already used) | Requires Dart learning | Requires separate iOS+Android teams |
| Time to prototype | Fast | Medium | Slowest |
| WS support | react-native-websocket (mature) | dart:io WebSocket | Platform native |

Given the existing TypeScript/React codebase, React Native maximizes code reuse and minimizes time to a working prototype. Flutter would be chosen if we prioritized rendering performance over code reuse.

### Shared Code Architecture

```
packages/
  shared/
    api/          ← REST API client (fetch-based, works in browser + RN)
    hooks/        ← useConversations, useMessages, useWebSocket
    types/        ← TypeScript interfaces (Message, Conversation, User)
  web/            ← Vite React (imports from shared/)
  mobile/         ← React Native (imports from shared/)
```

Use a **monorepo** (npm workspaces or Turborepo) to share the `api` and `hooks` layers between web and mobile.

### Mobile-Specific Additions

- **Push notifications**: Use `@react-native-firebase/messaging` (FCM) for Android; `react-native-push-notification` + APNs for iOS.
- **Background sync**: On app-to-background, disconnect WS; register FCM/APNs token. On foreground return, reconnect WS and sync via `/api/sync`.
- **Offline cache**: Use `@react-native-async-storage/async-storage` for last 50 messages per conversation.
- **Media picker**: `react-native-image-picker` for camera/gallery access.

### API Changes Required

The EchoLine API is already mobile-ready (REST + JWT + WS). The only addition needed:

```
POST /api/devices/push-token
{ "platform": "fcm" | "apns", "token": "<device push token>" }
```

Stored in the existing `devices` table with a new `push_token` and `push_platform` column.

## Implementation Files

- `frontend/` — current web app (to be refactored into monorepo)
- `mobile/` _(planned)_ — React Native app
- `packages/shared/` _(planned)_ — shared API + hooks
- `backend/migrations/` — add `push_token`, `push_platform` to `devices`
- `backend/internal/api/devices.go` _(planned)_ — push token registration API
- `backend/internal/notification/` _(planned)_ — FCM/APNs dispatch

## Consequences

**Positive:**
- Shared API and hook layers reduce duplication.
- React Native team can use existing TypeScript knowledge.
- Push notification support enables background delivery.

**Negative:**
- React Native's JS bridge adds overhead vs Flutter's Skia renderer.
- Metro bundler and RN ecosystem have historically required more maintenance than web tooling.
- Monorepo setup adds build system complexity.

## Interview Talking Points

- **Why not PWA?** "PWA push on iOS Safari is limited; background sync is unreliable. For a messaging app where background notifications are a core feature, a native wrapper is necessary."
- **Code sharing trade-off**: "We share the API client and business logic hooks, but not the UI components. Mobile UI patterns (gesture navigation, platform-native lists) are different enough that sharing UI would produce a worse product."
- **Push token lifecycle**: "Push tokens expire and rotate. We update the token on every app launch and handle APNs/FCM 'invalid token' errors by removing the stale token from the `devices` table."
