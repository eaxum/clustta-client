# Clustta Onboarding Flow Proposal

## Current Flow Summary

The current onboarding funnels **all users** through a single `Login.vue` page that tries to handle every user type with toggles and secondary links:

```
┌─────────────────────────────────────────────────┐
│              CURRENT LOGIN PAGE                  │
│                                                  │
│  ┌─── Clustta Logo ───┐                         │
│  │     "Log In"        │                         │
│  └─────────────────────┘                         │
│                                                  │
│  [Private Server toggle] ←── hidden by default   │
│    └─ Studio URL input (shown if toggled)        │
│                                                  │
│  [ Email / Username ]                            │
│  [ Password         ]                            │
│  "Forgot password?"                              │
│  [ Log In button    ]                            │
│                                                  │
│  "Don't have an account? Sign up"                │
│                                                  │
│  ──────── or ────────                            │
│                                                  │
│  "Use without an account" ←── offline mode       │
│                                                  │
└─────────────────────────────────────────────────┘
```

The **Sign Up page** mirrors the same pattern — one form with a "Private Server" toggle to switch between Clustta Cloud registration and self-hosted studio registration.

## Problems With the Current Flow

| Problem | Impact |
|---------|--------|
| **No user-type identification upfront** | A solo hobbyist sees the same screen as a studio admin. Nobody self-selects. |
| **"Private Server" toggle is cryptic** | The term means nothing to a new user. It's easy to miss and hard to understand. |
| **"Use without an account" is buried** | It's below a divider at the bottom — the simplest path (for the largest potential audience) is the least visible. |
| **One-size-fits-all Sign Up form** | Cloud sign-up (6 fields + email verification) is shown to studio users who don't need verification, and to offline users who don't need *any* of it. |
| **No mention of "collaboration" or "hosting"** | Users wanting collaboration or managed hosting can't tell from the login page that these options exist. |

## Target User Segments

1. **Individuals (Local/Offline)** — Want to use Clustta for personal projects locally, no account needed.
2. **Individuals (Personal + Collaborate)** — Want to work on personal projects and have others collaborate with them, without needing a server.
3. **Studios (Self-Hosted)** — Organisations that want to self-host Clustta.
4. **Studios (Managed Hosting)** — Organisations that want Clustta to host it for them.

## Proposed New Onboarding Flow

The core idea: **start with intent, not credentials.** Present a "Choose Your Path" screen that lets every user type self-identify in one click, then streamline each path independently.

### Flow Diagram

```
                    ┌─────────────────────────────────┐
                    │       🏠 WELCOME SCREEN          │
                    │          (New Page)               │
                    │                                   │
                    │   Clustta Logo                    │
                    │   "How do you want to use         │
                    │    Clustta?"                      │
                    │                                   │
                    │  ┌─────────────────────────────┐  │
                    │  │ 🖥️  Work Locally             │  │
                    │  │ Personal projects, no        │  │
                    │  │ account needed.              │  │
                    │  │              [Get Started →] │  │
                    │  └─────────────────────────────┘  │
                    │                                   │
                    │  ┌─────────────────────────────┐  │
                    │  │ 🌐  Personal + Collaborate   │  │
                    │  │ Work on your own projects    │  │
                    │  │ and invite others — no       │  │
                    │  │ server required.             │  │
                    │  │          [Create Account →]  │  │
                    │  └─────────────────────────────┘  │
                    │                                   │
                    │  ┌─────────────────────────────┐  │
                    │  │ 🏢  Studio / Organisation    │  │
                    │  │ Self-host or let us host     │  │
                    │  │ Clustta for your team.       │  │
                    │  │        [Set Up Studio →]     │  │
                    │  └─────────────────────────────┘  │
                    │                                   │
                    │  Already have an account?         │
                    │  [Sign In]                        │
                    └──────────┬──────┬──────┬──────────┘
                               │      │      │
              ┌────────────────┘      │      └────────────────┐
              ▼                       ▼                       ▼
   ┌──────────────────┐   ┌──────────────────┐   ┌───────────────────────┐
   │  PATH A: LOCAL   │   │ PATH B: PERSONAL │   │   PATH C: STUDIO     │
   │                  │   │  + COLLABORATE    │   │                      │
   │ Skip auth        │   │                  │   │  ┌─────────────────┐  │
   │ entirely.        │   │ Standard Cloud   │   │  │ How will your   │  │
   │                  │   │ Sign Up:         │   │  │ studio connect? │  │
   │ → enableOffline  │   │                  │   │  │                 │  │
   │   Mode()         │   │ First/Last Name  │   │  │ ○ Self-hosted   │  │
   │                  │   │ Username         │   │  │ ○ Clustta-      │  │
   │ → Set project    │   │ Email            │   │  │   hosted        │  │
   │   directory      │   │ Password ×2      │   │  └────┬──────┬────┘  │
   │                  │   │                  │   │       │      │       │
   │ → Land on        │   │ → Email verify   │   │       ▼      ▼       │
   │   home screen    │   │ → Login          │   │   ┌──────┐┌───────┐  │
   │                  │   │ → Home           │   │   │Self- ││Managed│  │
   │ 💡 Prompt:       │   │                  │   │   │Host  ││Host   │  │
   │ "Want to unlock  │   │ 💡 Prompt:       │   │   │Form  ││Form   │  │
   │  sync & collab?  │   │ "Create or join  │   │   │      ││       │  │
   │  Sign in         │   │  a studio to     │   │   │Studio││Contact│  │
   │  anytime."       │   │  scale up."      │   │   │URL + ││form / │  │
   │                  │   │                  │   │   │Admin ││waitlst│  │
   └──────────────────┘   └──────────────────┘   │   │signup││       │  │
                                                  │   └──────┘└───────┘  │
                                                  └──────────────────────┘

   ┌──────────────────────────────────────────────┐
   │         SIGN IN (existing users)              │
   │  (Reached from "Already have an account?")    │
   │                                               │
   │  [ Email / Username ]                         │
   │  [ Password         ]                         │
   │                                               │
   │  "Connecting to a studio server?"             │
   │  [ Enter Studio URL ] ←── inline, contextual  │
   │                                               │
   │  [ Sign In ]                                  │
   │  "Forgot password?"                           │
   └───────────────────────────────────────────────┘
```

## Page-Level Changes Needed

| Current File | What Changes |
|---|---|
| **New: `Welcome.vue`** | New landing page with the 3 path cards. Becomes the default `/auth` route. |
| **`Login.vue`** | Becomes the **Sign In** page only. Remove the "Use without account" link and the "Private Server" toggle from here. Instead, add a subtle "Connecting to a studio?" inline field that reveals the Studio URL input contextually. |
| **`SignUp.vue`** | Now only serves **Path B** (Personal + Collaborate). Remove the "Private Server" toggle entirely — studio users arrive through Path C instead. This greatly simplifies the form. |
| **New: `StudioSetup.vue`** | Dedicated page for **Path C**. Two sub-paths: self-hosted (Studio URL + admin registration form against that host) and managed hosting (contact/waitlist form or redirect). |
| **`router/index.js`** | Add `/auth/welcome` as the new default auth redirect. Add `/auth/studio-setup` route. Update the catch-all and `auth-default` redirect. |

## Router Change Sketch

```js
// New default auth landing
{
  path: '',
  name: 'auth-default',
  redirect: '/auth/welcome'   // was /auth/login
},
{
  path: 'welcome',
  name: 'welcome',
  component: () => import('@/instances/desktop/pages/Welcome.vue'),
  meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
},
{
  path: 'studio-setup',
  name: 'studio-setup',
  component: () => import('@/instances/desktop/pages/StudioSetup.vue'),
  meta: { requiresAuth: false, isPublic: true, isAuthPage: true }
},
// login & signup routes stay, but are reached contextually
```

## Key Design Principles

1. **Intent-first, credentials-second** — Users self-select before they see any form fields. This prevents cognitive overload.

2. **Shortest path to value for the biggest audience** — "Work Locally" is one click → offline mode → done. Zero friction for the user type that likely represents the most first-time users.

3. **No jargon on the first screen** — "Private Server" becomes "Studio / Organisation" with human-readable descriptions. "Self-host or let us host" covers both studio sub-types plainly.

4. **Progressive disclosure** — Studio users only see the Studio URL field *after* choosing that path. Cloud sign-up users never see it at all.

5. **Upgrade nudges built into each path** — After the offline user lands on the home screen, a gentle prompt ("Want to unlock sync &amp; collab? Sign in anytime.") shows the upgrade path without blocking them. Similarly, personal account users can be prompted to "Create or join a studio" later.

## Summary of UX Improvements

| Metric | Current | Proposed |
|---|---|---|
| Clicks to start working locally | 2 (find "Use without account" → click) | 1 (click "Work Locally → Get Started") |
| Concepts shown to a first-time user | ~6 (email, password, studio toggle, offline link, sign up link, forgot password) | 3 clear cards |
| Form fields for studio sign-up | Same form as cloud + a toggle | Dedicated page, tailored fields |
| Managed hosting discovery | Not surfaced at all | Explicit path in onboarding |
| Can a user self-identify their type? | No | Yes — first interaction |
