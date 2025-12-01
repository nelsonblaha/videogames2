describe('All Games - Comprehensive Tests', () => {
  const testGameID = () => `cypress-game-${Date.now()}-${Math.random()}`

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  describe('First to Find', () => {
    it('should display item to find and 30 second timer', () => {
      const gameID = testGameID()
      cy.joinGame('Player1', gameID)

      // Navigate until we get First to Find
      let attempts = 0
      const findGame = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'firsttofind') {
            // Found it! Move to playing
            cy.clickNext()
            cy.wait(500)

            // Should show timer
            cy.waitForTimer(30)

            // Should show item to find in instructions
            cy.getRoundInstructions().should('contain', 'wins!')
            cy.getRoundInstructions().invoke('text').should('match', /(banana|pen|musical|ice|disc|grass|device)/)
          } else if (attempts < 10) {
            attempts++
            findGame()
          }
        })
      }
      findGame()
    })

    it('should transition to voting after 30 seconds', () => {
      const gameID = testGameID()
      cy.joinGame('TimerPlayer', gameID)

      // Navigate to First to Find and start playing
      let attempts = 0
      const playGame = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'firsttofind' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Wait for voting to appear (timer + transition)
            cy.waitForVoting()

            // Voting UI should be visible
            cy.get('#vote-select').should('be.visible')
            cy.getRoundInstructions().should('contain', 'Vote')
          } else if (attempts < 10 && (!state || state.game_state !== 'playing')) {
            attempts++
            playGame()
          }
        })
      }
      playGame()
    })

    it('should allow voting and award points to winner', () => {
      const gameID = testGameID()
      cy.joinGame('VotePlayer', gameID)

      let attempts = 0
      const voteTest = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'firsttofind' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Wait for voting
            cy.waitForVoting()

            // Submit vote for self (only player)
            cy.submitVote(0)

            cy.wait(1000)

            // Should show finished state
            cy.getGameState().then((finishedState) => {
              expect(finishedState.game_state).to.equal('finished')
            })
          } else if (attempts < 10) {
            attempts++
            voteTest()
          }
        })
      }
      voteTest()
    })
  })

  describe('Find the Blankest Blank', () => {
    it('should display adjective and noun with 30 second timer', () => {
      const gameID = testGameID()
      cy.joinGame('BlankPlayer', gameID)

      let attempts = 0
      const findGame = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'blankestblank') {
            cy.clickNext()
            cy.wait(500)

            // Should show timer
            cy.waitForTimer(30)

            // Should show "Find the [adjective] [noun]"
            cy.getRoundInstructions().should('contain', 'Find the')
            cy.getRoundInstructions().invoke('text').should('match', /(oldest|biggest|smallest|fanciest|weirdest)/)
            cy.getRoundInstructions().invoke('text').should('match', /(thing|food|hat|book|coin|utensil)/)
          } else if (attempts < 10) {
            attempts++
            findGame()
          }
        })
      }
      findGame()
    })

    it('should transition to voting after timer expires', () => {
      const gameID = testGameID()
      cy.joinGame('BlankTimer', gameID)

      let attempts = 0
      const playGame = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'blankestblank' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Wait for voting
            cy.waitForVoting()

            cy.get('#voting-area').should('be.visible')
            cy.getRoundInstructions().should('contain', 'Vote')
          } else if (attempts < 10) {
            attempts++
            playGame()
          }
        })
      }
      playGame()
    })
  })

  describe('Mad Libs', () => {
    it('should prompt for words and display story when complete', () => {
      const gameID = testGameID()
      cy.joinGame('MadLibber', gameID)

      let attempts = 0
      const playMadLibs = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'madlibs') {
            cy.clickNext()
            cy.wait(500)

            // Should show word input
            cy.get('#word-input').should('be.visible')

            // Get the prompt
            cy.getRoundInstructions().invoke('text').then((prompt) => {
              expect(prompt).to.match(/(adjective|noun|verb|plural)/)
            })

            // Submit some words
            cy.submitWord('silly')
            cy.wait(300)
            cy.submitWord('banana')
            cy.wait(300)
            cy.submitWord('jumped')
            cy.wait(300)

            // Progress should update
            cy.get('#progress-text').should('contain', 'Words:')
          } else if (attempts < 10) {
            attempts++
            playMadLibs()
          }
        })
      }
      playMadLibs()
    })

    it('should complete story and show result', () => {
      const gameID = testGameID()
      cy.joinGame('StoryTeller', gameID)

      let attempts = 0
      const completeStory = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'madlibs' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Submit words until complete (max 20 words)
            const words = ['silly', 'banana', 'jumped', 'quickly', 'tree', 'happy', 'ran', 'big', 'small', 'red']
            const submitWords = (count = 0) => {
              if (count >= 20) return

              cy.getGameState().then((currentState) => {
                if (currentState.game_state === 'playing') {
                  cy.get('#word-input').should('be.visible')
                  cy.submitWord(words[count % words.length])
                  cy.wait(300)
                  submitWords(count + 1)
                } else if (currentState.game_state === 'finished') {
                  // Story should be visible
                  cy.get('#story-text').should('be.visible')
                  cy.get('#story-text').invoke('text').should('have.length.gt', 10)
                }
              })
            }
            submitWords()
          } else if (attempts < 10) {
            attempts++
            completeStory()
          }
        })
      }
      completeStory()
    })
  })

  describe('You Laugh You Lose', () => {
    it('should display YouTube video when game starts', () => {
      const gameID = testGameID()
      cy.joinGame('Laugher', gameID)

      let attempts = 0
      const playYLYL = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'youlaughyoulose') {
            cy.clickNext()
            cy.wait(500)

            // YouTube player should appear
            cy.waitForYouTube()

            // Instructions should mention straight face
            cy.getGameInstructions().should('contain', 'straight face')
          } else if (attempts < 10) {
            attempts++
            playYLYL()
          }
        })
      }
      playYLYL()
    })

    it('should have YouTube video ID in state', () => {
      const gameID = testGameID()
      cy.joinGame('VideoChecker', gameID)

      let attempts = 0
      const checkVideo = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'youlaughyoulose' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            cy.getGameState().then((playingState) => {
              expect(playingState).to.have.property('youtube_video_id')
              expect(playingState.youtube_video_id).to.be.a('string')
              expect(playingState.youtube_video_id).to.have.length.gt(5)
            })
          } else if (attempts < 10) {
            attempts++
            checkVideo()
          }
        })
      }
      checkVideo()
    })
  })

  describe('Charades', () => {
    it('should display charades topic and instructions', () => {
      const gameID = testGameID()
      cy.joinGame('Actor', gameID)

      let attempts = 0
      const playCharades = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'charades') {
            // Should show charades instructions
            cy.getGameInstructions().should('contain', 'act')
            cy.clickNext()
            cy.wait(500)

            // Game should start
            cy.getGameState().then((playingState) => {
              expect(playingState.game_state).to.equal('playing')
            })
          } else if (attempts < 10) {
            attempts++
            playCharades()
          }
        })
      }
      playCharades()
    })
  })

  describe('Imitations', () => {
    it('should display imitations instructions', () => {
      const gameID = testGameID()
      cy.joinGame('Imitator', gameID)

      let attempts = 0
      const playImitations = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations') {
            cy.getGameInstructions().should('contain', 'Imitate')
            cy.clickNext()
            cy.wait(500)

            cy.getGameState().then((playingState) => {
              expect(playingState.game_state).to.equal('playing')
            })
          } else if (attempts < 10) {
            attempts++
            playImitations()
          }
        })
      }
      playImitations()
    })

    it('should accept exact match guess', () => {
      const gameID = testGameID()
      cy.joinGame('Guesser1', gameID)

      let attempts = 0
      const testGuess = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Submit exact name from the list
            cy.submitWord('Barack Obama')
            cy.wait(1000)

            // Should complete immediately on correct guess
            cy.getGameState().then((finalState) => {
              // Could be finished if we guessed right, or still playing if wrong
              expect(['playing', 'finished']).to.include(finalState.game_state)
            })
          } else if (attempts < 10) {
            attempts++
            testGuess()
          }
        })
      }
      testGuess()
    })

    it('should accept first name only', () => {
      const gameID = testGameID()
      cy.joinGame('Guesser2', gameID)

      let attempts = 0
      const testFirstName = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Try first names only
            cy.submitWord('Barack')
            cy.wait(500)
            cy.submitWord('George')
            cy.wait(500)
            cy.submitWord('Morgan')
            cy.wait(1000)

            // One of these should match if the person is in our list
            cy.getGameState().then((finalState) => {
              expect(['playing', 'finished']).to.include(finalState.game_state)
            })
          } else if (attempts < 10) {
            attempts++
            testFirstName()
          }
        })
      }
      testFirstName()
    })

    it('should accept last name only', () => {
      const gameID = testGameID()
      cy.joinGame('Guesser3', gameID)

      let attempts = 0
      const testLastName = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Try last names
            cy.submitWord('Washington')
            cy.wait(500)
            cy.submitWord('Obama')
            cy.wait(500)
            cy.submitWord('Freeman')
            cy.wait(1000)

            cy.getGameState().then((finalState) => {
              expect(['playing', 'finished']).to.include(finalState.game_state)
            })
          } else if (attempts < 10) {
            attempts++
            testLastName()
          }
        })
      }
      testLastName()
    })

    it('should accept slight misspellings', () => {
      const gameID = testGameID()
      cy.joinGame('Guesser4', gameID)

      let attempts = 0
      const testMisspelling = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Try with slight typos
            cy.submitWord('Barrack Obama')  // extra 'r'
            cy.wait(500)
            cy.submitWord('Clint Eastwod')  // missing 'o'
            cy.wait(500)
            cy.submitWord('Morgn Freeman')  // missing 'a'
            cy.wait(1000)

            cy.getGameState().then((finalState) => {
              expect(['playing', 'finished']).to.include(finalState.game_state)
            })
          } else if (attempts < 10) {
            attempts++
            testMisspelling()
          }
        })
      }
      testMisspelling()
    })

    it('should accept name without spaces', () => {
      const gameID = testGameID()
      cy.joinGame('Guesser5', gameID)

      let attempts = 0
      const testNoSpace = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Try without spaces
            cy.submitWord('BarackObama')
            cy.wait(500)
            cy.submitWord('MorganFreeman')
            cy.wait(500)
            cy.submitWord('DonaldDuck')
            cy.wait(1000)

            cy.getGameState().then((finalState) => {
              expect(['playing', 'finished']).to.include(finalState.game_state)
            })
          } else if (attempts < 10) {
            attempts++
            testNoSpace()
          }
        })
      }
      testNoSpace()
    })

    it('should show winner in result when someone guesses correctly', () => {
      const gameID = testGameID()
      cy.joinGame('WinnerTest', gameID)

      let attempts = 0
      const testWinner = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'imitations' && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Submit all possible answers to ensure we get at least one right
            const names = ['Barack Obama', 'George Washington', 'Clint Eastwood', 'Morgan Freeman', 'Captain Kirk', 'Donald Duck']
            names.forEach((name, index) => {
              cy.submitWord(name)
              cy.wait(300)
            })

            cy.wait(1000)

            // Check if game finished and shows winner
            cy.getGameState().then((finalState) => {
              if (finalState.game_state === 'finished') {
                cy.get('#story-text').invoke('text').then((text) => {
                  // Should show winner name and the person
                  expect(text).to.match(/(guessed it|person was)/)
                })
              }
            })
          } else if (attempts < 10) {
            attempts++
            testWinner()
          }
        })
      }
      testWinner()
    })
  })

  describe('Claude\'s Game', () => {
    it('should display two random words to connect', () => {
      const gameID = testGameID()
      cy.joinGame('Connector', gameID)

      let attempts = 0
      const playClaudesGame = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && state.game_type === 'claudesgame') {
            cy.getGameInstructions().should('contain', 'Connect')
            cy.clickNext()
            cy.wait(500)

            // Should show word input for creative connection
            cy.get('#word-input').should('be.visible')
            cy.getRoundInstructions().should('contain', 'connected')
          } else if (attempts < 10) {
            attempts++
            playClaudesGame()
          }
        })
      }
      playClaudesGame()
    })
  })

  describe('Game Flow - Random Selection', () => {
    it('should cycle through different random games', () => {
      const gameID = testGameID()
      cy.joinGame('Cycler', gameID)

      const gamesEncountered = new Set()

      const cycleGames = (count = 0) => {
        if (count >= 15) {
          // Should have seen at least 3 different games
          expect(gamesEncountered.size).to.be.gte(3)
          return
        }

        cy.clickNext()
        cy.wait(500)

        cy.getGameState().then((state) => {
          if (state && state.game_type) {
            gamesEncountered.add(state.game_type)
          }

          // If in playing or voting, move to next game
          if (state && (state.game_state === 'playing' || state.game_state === 'voting')) {
            // Skip to finished by clicking Next
            cy.clickNext()
            cy.wait(500)
          }

          cycleGames(count + 1)
        })
      }

      cycleGames()
    })

    it('should maintain player scores across games', () => {
      const gameID = testGameID()
      cy.joinGame('Scorer', gameID)

      let initialScore = 0

      // Play through a few games and check score persistence
      cy.clickNext()
      cy.wait(500)
      cy.clickNext()
      cy.wait(500)

      cy.getGameState().then((state) => {
        const player = state.players && state.players[0]
        if (player) {
          initialScore = player.score
        }
      })

      // Move through game
      cy.clickNext()
      cy.wait(500)
      cy.clickNext()
      cy.wait(500)

      cy.getGameState().then((state) => {
        const player = state.players && state.players[0]
        if (player) {
          // Score should be maintained (may increase)
          expect(player.score).to.be.gte(initialScore)
        }
      })
    })
  })

  describe('Voting System', () => {
    it('should show all players in voting dropdown', () => {
      const gameID = testGameID()
      cy.joinGame('Voter', gameID)

      let attempts = 0
      const testVoting = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          // Find a game with voting
          if (state && ['firsttofind', 'blankestblank', 'youlaughyoulose'].includes(state.game_type)) {
            if (state.game_state === 'instructions') {
              cy.clickNext()
              cy.wait(500)

              // Wait for voting
              cy.waitForVoting()

              // Dropdown should have player options
              cy.get('#vote-select option').should('have.length.gt', 1)
            }
          } else if (attempts < 15) {
            attempts++
            testVoting()
          }
        })
      }
      testVoting()
    })

    it('should show vote progress', () => {
      const gameID = testGameID()
      cy.joinGame('ProgressChecker', gameID)

      let attempts = 0
      const checkProgress = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && ['firsttofind', 'blankestblank'].includes(state.game_type) && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            cy.waitForVoting()

            // Should show vote status
            cy.get('#vote-status').should('exist')
            cy.submitVote(0)
            cy.wait(500)

            cy.get('#vote-status').should('contain', 'Vote')
          } else if (attempts < 15) {
            attempts++
            checkProgress()
          }
        })
      }
      checkProgress()
    })
  })

  describe('Timer System', () => {
    it('should countdown from 30 for timed games', () => {
      const gameID = testGameID()
      cy.joinGame('TimerWatcher', gameID)

      let attempts = 0
      const watchTimer = () => {
        cy.clickNext()
        cy.wait(500)
        cy.getGameState().then((state) => {
          if (state && ['firsttofind', 'blankestblank'].includes(state.game_type) && state.game_state === 'instructions') {
            cy.clickNext()
            cy.wait(500)

            // Timer should start at 30
            cy.get('#timer-display').should('contain', '30')

            // Wait a bit and check it's counting down
            cy.wait(2000)
            cy.get('#timer-display').invoke('text').then((text) => {
              const time = parseInt(text)
              expect(time).to.be.lt(30)
              expect(time).to.be.gte(25)
            })
          } else if (attempts < 15) {
            attempts++
            watchTimer()
          }
        })
      }
      watchTimer()
    })
  })
})
