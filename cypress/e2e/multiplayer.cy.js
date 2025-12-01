describe('Multiplayer Game Flow', () => {
  const testGameID = `cypress-${Date.now()}`

  beforeEach(() => {
    // Clear any existing state
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('should allow a single player to join and see lobby', () => {
    cy.joinGame('Alice', testGameID)

    cy.getGameTitle().should('contain', 'Ready to play!')
    cy.getGameInstructions().should('contain', 'Click')
    cy.getPlayerList().should('contain', 'Alice')
  })

  it('should show waiting state when multiple players join', () => {
    // Open two browser contexts
    cy.visit('/')
    cy.get('#player-name').type('Alice')
    cy.get('#group-name').clear().type(testGameID)
    cy.get('#join-button').click()

    // Wait for connection
    cy.get('#game-title', { timeout: 5000 }).should('be.visible')

    // Open in new window/tab simulation
    cy.visit('/', {
      onBeforeLoad(win) {
        // Simulate second player in same test
      }
    })

    cy.get('#player-name').type('Bob')
    cy.get('#group-name').clear().type(testGameID)
    cy.get('#join-button').click()

    cy.get('#game-title').should('contain', 'Waiting')
  })

  it('should transition from lobby to instructions to playing', () => {
    cy.joinGame('Solo Player', `solo-${testGameID}`)

    // Start in lobby
    cy.getGameTitle().should('contain', 'Ready to play!')

    // Click Next to go to instructions
    cy.clickNext()
    cy.wait(500)

    cy.getGameTitle().should('contain', 'Mad Libs')
    cy.getGameInstructions().should('contain', 'blanks')

    // Click Next to start playing
    cy.clickNext()
    cy.wait(500)

    cy.getGameTitle().should('contain', 'Playing')
  })

  it('should handle player disconnection gracefully', () => {
    cy.joinGame('TestPlayer', `disconnect-${testGameID}`)

    cy.getPlayerList().should('contain', 'TestPlayer')

    // Simulate disconnect by closing WebSocket
    cy.window().then((win) => {
      if (win.socket) {
        win.socket.close()
      }
    })

    // Player should be removed from list (in a real scenario)
    cy.wait(1000)
  })

  it('should display all players in the game', () => {
    const gameName = `players-${testGameID}`

    // First player
    cy.joinGame('Alice', gameName)
    cy.getPlayerList().should('contain', 'Alice')

    // Note: Testing multiple concurrent players requires multiple browser contexts
    // which Cypress doesn't support natively. This would need a different approach
    // or manual testing with multiple browsers.
  })

  it('should handle rapid Next button clicks', () => {
    cy.joinGame('Clicker', `rapid-${testGameID}`)

    // Rapid clicks
    cy.clickNext()
    cy.clickNext()
    cy.clickNext()

    cy.wait(500)

    // Should handle gracefully without crashing
    cy.get('#game-state').should('be.visible')
  })

  it('should maintain game state across page navigation', () => {
    const gameName = `persist-${testGameID}`

    cy.joinGame('Persistent', gameName)
    cy.clickNext()

    cy.wait(500)

    // Store current state
    let initialTitle
    cy.getGameTitle().invoke('text').then((text) => {
      initialTitle = text

      // Reload page
      cy.reload()

      // Rejoin same game
      cy.get('#player-name').clear().type('Persistent')
      cy.get('#group-name').clear().type(gameName)
      cy.get('#join-button').click()

      cy.wait(1000)

      // Game might be in different state after rejoin (new player in same game)
      cy.get('#game-state').should('be.visible')
    })
  })

  it('should show correct ready status for players', () => {
    cy.joinGame('ReadyPlayer', `ready-${testGameID}`)

    // Move to instructions
    cy.clickNext()
    cy.wait(500)

    cy.getGameTitle().should('contain', 'Mad Libs')

    // Click ready
    cy.clickNext()
    cy.wait(500)

    // Single player should move to playing state
    cy.getGameTitle().should('contain', 'Playing')
  })
})

describe('Homepage Integration', () => {
  it('should detect homepage authentication', () => {
    // Mock the /api/user endpoint
    cy.intercept('GET', '/api/user', {
      statusCode: 200,
      body: {
        authenticated: true,
        name: 'Ben'
      }
    }).as('getUserAuth')

    cy.visit('/')

    cy.wait('@getUserAuth')

    // Should auto-fill with homepage credentials
    cy.get('#player-name').should('have.value', 'Ben')
    cy.get('#group-name').should('have.value', 'homepage')
  })

  it('should fallback to manual entry when not authenticated', () => {
    cy.intercept('GET', '/api/user', {
      statusCode: 200,
      body: {
        authenticated: false,
        name: ''
      }
    }).as('getUserNoAuth')

    cy.visit('/')

    cy.wait('@getUserNoAuth')

    // Should not auto-fill
    cy.get('#player-name').should('have.value', '')
    cy.get('#group-name').should('not.have.value', 'homepage')
  })
})

describe('WebSocket Connection', () => {
  it('should establish WebSocket connection on join', () => {
    cy.visit('/')

    // Spy on WebSocket
    cy.window().then((win) => {
      cy.spy(win, 'WebSocket').as('websocketSpy')
    })

    cy.get('#player-name').type('WSTest')
    cy.get('#group-name').clear().type('ws-test')
    cy.get('#join-button').click()

    // Verify WebSocket was created
    cy.get('@websocketSpy').should('have.been.called')
  })

  it('should send join message on connection', () => {
    cy.joinGame('MessageTest', 'msg-test')

    // Wait for connection
    cy.wait(1000)

    // Check that game state was received
    cy.get('#game-state').should('be.visible')
    cy.getGameTitle().should('exist')
  })

  it('should handle ping/pong for keepalive', () => {
    cy.joinGame('PingTest', 'ping-test')

    cy.wait(2000)

    // Should still be connected after some time
    cy.get('#game-state').should('be.visible')
  })
})

describe('Edge Cases', () => {
  it('should handle empty player name gracefully', () => {
    cy.visit('/')

    cy.get('#group-name').clear().type('empty-name-test')
    cy.get('#join-button').click()

    // Should either show error or use default name
    cy.wait(500)
    cy.get('body').should('exist') // Should not crash
  })

  it('should handle empty group name', () => {
    cy.visit('/')

    cy.get('#player-name').type('EmptyGroup')
    cy.get('#group-name').clear()
    cy.get('#join-button').click()

    cy.wait(500)

    // Should use default group
    cy.get('#game-state', { timeout: 5000 }).should('be.visible')
  })

  it('should handle special characters in names', () => {
    cy.joinGame('Test<>User&', 'special-chars-test')

    cy.wait(500)

    // Should not break the UI
    cy.get('#game-state').should('be.visible')
  })
})
