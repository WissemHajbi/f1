# OIDYSTS mobile

A deliberately small Expo SDK 54 + TypeScript client for the local Formula 1 API. SDK 54 is intentional because it matches the current Play Store release of Expo Go.

## Current scope

The validated MVP is intentionally narrow:

- Home: latest result and championship leaders
- Calendar: complete synchronized season schedule
- Standings: driver and constructor classification
- Race Details: circuit summary and final classification, opened from Calendar
- Drivers: searchable synchronized grid and simple driver profiles

Team radio, timelines, and track replay are intentionally deferred. They will be added one vertical feature at a time.

## Start

Run the API from the repository root:

```powershell
go run ./cmd/api
```

Then start Expo:

```powershell
cd mobile
npm install
npm start
```

Defaults:

- Android emulator: `http://10.0.2.2:8080`
- iOS simulator/web: `http://localhost:8080`
- Season: `2025`

For Expo Go on a physical device, copy `.env.example` to `.env`, set `EXPO_PUBLIC_API_URL` to the computer's LAN address, keep both devices on the same network, and allow port 8080 through the firewall.

## Checks

```powershell
npm run typecheck
npm run lint
npx expo-doctor
```

The app contacts only the Go backend. It never calls OpenF1 or Jolpica directly.
