# PulseFit (Cais)

Workout tracker built with [Cais](https://github.com/puppe1990/cais): server-rendered HTML, HTMX, Tailwind, and SQLite.

Go port of the React + Firebase [pulsefit](https://github.com/puppe1990/pulsefit) prototype — same dark theme, routines, live workout logging, and history.

Demo account (seeded on startup): `demo@pulsefit.local` / `demo`

Unauthenticated visits redirect to `/login`.

## Pages

| Route | Page |
|-------|------|
| `/login` | Login (POST email + password) |
| `/register` | Register (POST) |
| `POST /logout` | End session |
| `/` | Dashboard |
| `/search` | Browse categories |
| `/library` | Exercise library |
| `/routine/new` | Create routine (POST with exercise checkboxes) |
| `/routine/{id}` | Routine detail + Start Workout |
| `/workout/{id}` | Live workout logger (HTMX sets) |
| `/workout/{id}/summary` | Post-workout summary |
| `/history` | Workout history |

## Quick start

```bash
npm install
make css       # build Tailwind
make dev       # http://localhost:8080
make test      # full test suite
```

## Stack

- **Go** + [Cais](https://github.com/puppe1990/cais) v0.3.0 (session auth, HTMX, renderer)
- **SQLite** via `modernc.org/sqlite`
- **Tailwind CSS** v3

## Schema

`users`, `exercises`, `routines`, `routine_exercises`, `workout_sessions`, `exercise_logs`, `set_logs`