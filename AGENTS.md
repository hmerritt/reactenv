# AGENTS

Short rules for working in this repo.

## Scope

- This is a Go CLI project to inject environment variables into **bundled** javascript files.
- All changes must be clean and testable.

## Setup

- See `magefile.go` for all available scripts

```sh
# setup
mage -v bootstrap
```

```sh
# build single debug binary (current platform), outputs to root directory
mage -v build:debug
```

```sh
# build all release binaries (cross platform), outputs to bin/
mage -v build:release
```

```sh
# bundles all binaries into zips, ready for release/distribution
mage -v release
```

## Structure

- `bin/` output directory for binaries
- `command/` package containing all CLI commands
- `npm/` npm release directories
- `reactenv/` core package for logic
- `tools/` package for build tools. ensures tool dependencies are kept in sync
- `ui/` package for terminal UI logic
- `version/` package just for the version

## Conventions

- Add brief code comments for tricky or non-obvious logic.

## Tests

- Tests are written in `Vitest`
- Everything should be tested, and be testable
- New behavior requires tests (unit or e2e).

### Test tips

A few tips to write better tests:

[Russ Cox - Go Testing By Example](https://www.youtube.com/watch?v=1-o-iJlL4ak)

- Make it easy to add new tests.
- Use test coverage to find untested code.
- Coverage is no substitute for thought.
- Write exhaustive tests.
- Separate test cases from test logic (i.e use test case tables, separate from logic).
- Look for special cases.
- If you didn't add a test, you didn't fix the bug.
- Test cases can be in testdata files.
- Compare against other implementations.
- Make test failures readable.
- If the answers can change, wtite coed to update them.
- Code quality is limited by test quality.
- Scripts make good test cases.
- Improve your tests over time.

## Commits

- Follow `CONTRIBUTING.md` prefix rules and lowercase messages.

## Agent-Specific Notes

- When answering questions, respond with high-confidence answers only: verify in code; do not guess.
