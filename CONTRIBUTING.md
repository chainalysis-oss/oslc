# Contribution guide

Everyone can contribute with pull requests and bug reports.

## Development

### Commit Messages

This project follows [Conventional Commits] and asks that contributors follow the
[Conventional Commits Specification][Conventional Commits].

If you are not sure what [Conventional Commits] are, and isn't in the mood to read the specification, fear not. You
can still contribute to this project, just know that a maintainer will have to do some work to get your changes
merged.

#### For maintainers

In the event of a Pull Request without any commit messages that adhere to the [Conventional Commits] specification,
please ensure that changes are squashed into a single commit with a commit message that adheres to the
[Conventional Commits] specification.

### Branches

This project does not follow any specific Git branching model, and you are free to name your branches as you see fit.

We kindly ask that branch names do not contain spaces or forward slashes (`/`) because these makes working with git
tedious.

There are 3 reserved branches with special meaning:

- `main` - Is the main release branch. Conventional commits on this branch will trigger a production release.
- `beta` - Is the beta branch. Conventional commits on this branch will trigger a pre-release.
- `dev` - Is the development branch. Conventional commits on this branch will trigger a pre-release.

The `dev` branch is to be considered highly unstable.

The `beta` branch is to be considered somewhat stable. Some people might refer to releases from this branch as
release candidates, but we've opted to call it beta.

The `main` branch is to be considered stable and is considered a supported production release.

[Conventional Commits]: https://www.conventionalcommits.org/en/v1.0.0/