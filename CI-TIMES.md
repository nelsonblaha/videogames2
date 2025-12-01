# CI Run Times - videogames2 Project

**Last Updated:** 2025-12-01

This file tracks typical CI run durations for the videogames2 project to help Claude determine appropriate wait times before checking CI status.

## Typical Run Times

### Full Pipeline (Tests + Build)
- **Average:** ~5-7 minutes
- **Range:** 330-467 seconds (from recent runs)
- **Jobs:** test, cypress, build

## Job Breakdown

**Max Parallel Jobs: 2** (test + cypress run concurrently)

Individual job dependencies:
- **test (Go unit tests):** Runs first, ~2-3 min
- **cypress (E2E tests):** Runs in parallel with test, ~3-4 min
- **build (Docker image):** Runs after test completes, ~2-3 min

## Recommended Sleep Times

When waiting for CI after a push:
- First check: `sleep 60` (wait 1 min to let jobs start)
- Subsequent checks: `sleep 120` (check every 2 minutes)
- Full pipeline expected completion: ~6 minutes after push

## Notes

- Times based on self-hosted GitHub runners on NELNET
- Runners are on the same machine, so times are consistent
- Build step only runs after test job passes
- Cypress tests include full server startup and browser automation

## Update History

- 2025-12-01: Initial file created based on recent CI run analysis
