# Videogames2 - Actor-Based Multiplayer Game Platform

A modern multiplayer party games platform built with Go, WebSockets, and the Actor model for concurrency.

## Architecture

### Actor Model Implementation

This project demonstrates a clean actor-based architecture in Go:

```
┌─────────────────┐
│ GameCoordinator │  - Manages all game sessions
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼────┐
│ Game  │ │ Game  │  - Each game is an independent actor
│Actor 1│ │Actor 2│  - Processes messages sequentially
└───┬───┘ └──┬────┘  - Thread-safe state management
    │        │
  ┌─┴─┐    ┌┴──┐
  │ P │    │ P │    - Players connected via WebSockets
  └───┘    └───┘
```

### Components

**Actor (`actor.go`)**
- Base actor implementation with message inbox
- Sequential message processing
- Graceful lifecycle management (Start/Stop)

**GameActor (`game_actor.go`)**
- Manages a single game session
- Handles player join/leave
- State transitions: lobby → instructions → playing
- Broadcasts state updates to all players

**GameCoordinator (`coordinator.go`)**
- Creates and manages GameActors
- Automatic cleanup of empty games
- Thread-safe game lookup and creation

**Messages (`messages.go`)**
- Type-safe message definitions
- PlayerJoinMsg, PlayerLeaveMsg, NextGameMsg, etc.
- GetGameStateMsg for querying state

## Why Actors?

1. **Concurrency Safety**: Each game processes messages sequentially, eliminating race conditions
2. **Isolation**: Games are independent; bugs in one don't affect others
3. **Scalability**: Easy to distribute actors across machines
4. **Testability**: Actors can be tested in isolation with message injection
5. **Clean Architecture**: Clear separation of concerns and message contracts

## Running Tests

### Unit Tests (Go)

```bash
go test -v ./...
```

Tests cover:
- Actor message passing and lifecycle
- Game state transitions
- Player join/leave behavior
- Coordinator game management
- Concurrent access patterns

### End-to-End Tests (Cypress)

```bash
# Install dependencies
npm install

# Run tests headless
npm run test:cypress

# Open Cypress UI
npm run test:cypress:open
```

Cypress tests cover:
- Single and multiplayer game flow
- State transitions (lobby → instructions → playing)
- Homepage authentication integration
- WebSocket connection handling
- Edge cases (empty names, special characters)

## Development

### Local Development

```bash
# Build and run
go build -o videogames2 .
./videogames2

# Or with Docker
docker build -t videogames2 .
docker run -p 8080:8080 videogames2
```

### Project Structure

```
.
├── actor.go              # Base actor implementation
├── actor_test.go         # Actor unit tests
├── game_actor.go         # Game session actor
├── game_actor_test.go    # Game logic tests
├── coordinator.go        # Game coordinator
├── coordinator_test.go   # Coordinator tests
├── messages.go           # Message type definitions
├── main.go              # HTTP server and WebSocket handler
├── cypress/             # E2E tests
│   ├── e2e/
│   │   └── multiplayer.cy.js
│   └── support/
│       └── e2e.js
└── static/              # Frontend HTML/JS
    └── index.html
```

## Homepage Integration

The app integrates with the nelnet homepage for authentication:

1. Checks `/api/user` endpoint for X-Remote-User header
2. Auto-fills player name if authenticated
3. Falls back to manual entry for standalone mode
4. All sessions accessible through homepage iframe

## Deployment

Deployed via Docker Compose with:
- videogames2 Go server
- Dedicated Jitsi Meet stack (web, prosody, jicofo, jvb)
- nginx forward auth for homepage integration
- Automatic SSL via letsencrypt

```bash
cd /home/ben/stacks/videogames2
docker compose up -d --build
```

## CI/CD

GitHub Actions workflow runs:
1. Go unit tests (all actor tests)
2. Cypress E2E tests (multiplayer scenarios)
3. Docker build
4. Automatic deployment on master branch

## Future Enhancements

- [ ] Add more games (beyond Mad Libs)
- [ ] Persist game scores to database
- [ ] Add matchmaking for random games
- [ ] Implement game replay system
- [ ] Add spectator mode
- [ ] Metrics and monitoring with actor supervision

## License

MIT
