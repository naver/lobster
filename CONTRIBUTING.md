## Contributing

Contributions are always welcome!

All other things are managed later but the simple branch rules are exist.
The branches in this project are like below:

- `Your` branch
  - All is ok if the prefix of your branch name describes your purpose(e.g. `fix-`, `feat-`, etc...)
- `main` branch
  - protected: require a pull request before merging
  - reflect all changes
- `release-*` branch
  - version rule: `release-<major>.<minor>`
  - protected: require a pull request before merging
  - manage conflicts to process new release version
  - To sync the `main` branch, merge the `main` branch into the `release` branch first to resolve conflicts
    then merge the `release` branch into the `main` branch