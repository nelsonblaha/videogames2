// Custom Cypress commands for videogames2 testing

// Mock Jitsi and YouTube before each test
beforeEach(() => {
  // Intercept external script loads
  cy.intercept('GET', '**/external_api.js', '').as('jitsiScript')
  cy.intercept('GET', '**/iframe_api', '').as('youtubeScript')

  cy.on('window:before:load', (win) => {
    // Mock JitsiMeetExternalAPI
    win.JitsiMeetExternalAPI = class {
      constructor() {
        this.disposed = false
      }
      dispose() {
        this.disposed = true
      }
    }

    // Mock YouTube API
    win.YT = {
      Player: class {
        constructor() {}
        playVideo() {}
        stopVideo() {}
        loadVideoById() {}
      },
      PlayerState: {
        ENDED: 0
      }
    }
  })
})

Cypress.Commands.add('joinGame', (playerName, groupName = 'test-game') => {
  cy.visit('/')
  cy.get('#player-name').clear().type(playerName)
  cy.get('#group-name').clear().type(groupName)
  cy.get('#join-button').click()

  // Wait for WebSocket connection
  cy.get('#game-state', { timeout: 5000 }).should('be.visible')
})

Cypress.Commands.add('clickNext', () => {
  cy.get('#next-button').click()
})

Cypress.Commands.add('getGameTitle', () => {
  return cy.get('#game-title')
})

Cypress.Commands.add('getGameInstructions', () => {
  return cy.get('#game-instructions')
})

Cypress.Commands.add('getRoundInstructions', () => {
  return cy.get('#round-instructions')
})

Cypress.Commands.add('getPlayerList', () => {
  return cy.get('#players-list')
})

Cypress.Commands.add('waitForState', (stateName, timeout = 5000) => {
  cy.window().its('lastGameState').should('exist')
  cy.window().its('lastGameState.game_title').should('include', stateName)
})

Cypress.Commands.add('submitWord', (word) => {
  cy.get('#word-input').clear().type(word)
  cy.get('#word-input').type('{enter}')
})

Cypress.Commands.add('submitVote', (playerIndex = 0) => {
  cy.get('#vote-select').select(playerIndex + 1) // +1 because first option is blank
  cy.get('#voting-area button').click()
})

Cypress.Commands.add('waitForTimer', (seconds) => {
  cy.get('#timer-display', { timeout: 1000 }).should('be.visible')
  cy.get('#timer-display').should('contain', seconds)
})

Cypress.Commands.add('waitForVoting', () => {
  cy.get('#voting-area', { timeout: 35000 }).should('be.visible')
  cy.get('#vote-select').should('be.visible')
})

Cypress.Commands.add('waitForYouTube', () => {
  cy.get('#youtube-container', { timeout: 5000 }).should('not.have.class', 'hidden')
})

Cypress.Commands.add('getGameState', () => {
  return cy.window().then((win) => win.lastGameState)
})
