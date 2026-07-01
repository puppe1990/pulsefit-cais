# PulseFit (Cais)

Workout tracker built with [Cais](https://github.com/puppe1990/cais): server-rendered HTML, HTMX, Tailwind, and SQLite.

Go port of the React + Firebase [pulsefit](https://github.com/puppe1990/pulsefit) prototype — same dark theme, routines, live workout logging, and history.

Demo account (seeded on startup): `demo@pulsefit.local` / `demo`

Unauthenticated visits redirect to `/login`.

## Pages

| Route                   | Page                                           |
| ----------------------- | ---------------------------------------------- |
| `/login`                | Login (POST email + password)                  |
| `/register`             | Register (POST)                                |
| `POST /logout`          | End session                                    |
| `/`                     | Dashboard                                      |
| `/search`               | Browse categories                              |
| `/library`              | Exercise library                               |
| `/routine/new`          | Create routine (POST with exercise checkboxes) |
| `/routine/{id}`         | Routine detail + Start Workout                 |
| `/workout/{id}`         | Live workout logger (HTMX sets)                |
| `/workout/{id}/summary` | Post-workout summary                           |
| `/history`              | Workout history                                |

## Quick start

```bash
git clone https://github.com/puppe1990/pulsefit-cais.git
cd pulsefit-cais
go install github.com/puppe1990/cais/cmd/cais@v0.3.3
export PATH="$HOME/go/bin:$PATH"   # add to ~/.zshrc to persist
cais install   # npm install + go mod tidy
cais dev       # http://localhost:8080
cais test      # full test suite
```

## Stack

- **Go** + [Cais](https://github.com/puppe1990/cais) v0.3.3 (session auth, HTMX, Rails-style logs)
- **SQLite** via `modernc.org/sqlite`
- **Tailwind CSS** v3

## Schema

`users`, `exercises`, `routines`, `routine_exercises`, `workout_sessions`, `exercise_logs`, `set_logs`
