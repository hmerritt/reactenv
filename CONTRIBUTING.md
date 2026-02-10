# Contributing

#1 Rule. Try to write good code; no worries if not.

## Commit rules

### Commit message

A good commit message should describe what changed and why.

It should:

- contain a short description of the change (preferably 50 characters or less)
- be entirely in lowercase with the exception of proper nouns, acronyms, and the words that refer to code, like function/variable names
- be prefixed with one of the following word:
    - `fix` : bug fix
    - `hotfix` : urgent bug fix
    - `feat` : new or updated feature
    - `refactor` : code refactoring (no functional change)
    - `test` : tests updates
    - `ci` : ci and build updates
    - `chore` : miscellaneous housekeeping updates
    - `Merge branch` : when merging branch
    - `Merge pull request` : when merging PR

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
