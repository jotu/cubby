# Contributing to Cubby

Thanks for your interest in contributing.

## How to contribute

1. Fork the repository
2. Create a feature branch
3. Make focused changes with tests
4. Run verification locally
5. Open a pull request with a clear description

## Development setup

```sh
mise run dev
```

## Quality checks

Before opening a PR, run:

```sh
mise run verify
```

This project follows a TDD workflow (Red → Green → Tidy) and prefers small, reviewable pull requests.

## Style expectations

- Match existing project structure and naming
- Keep changes minimal and focused
- Add or update tests for behavior changes
- Avoid unrelated refactors in the same PR

## Commit messages

Conventional Commits are preferred (e.g. `feat:`, `fix:`, `docs:`).

## Questions

Open an issue for questions, proposals, or discussions before large changes.
