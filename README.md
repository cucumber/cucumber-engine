# Cucumber Engine

[![CircleCI](https://circleci.com/gh/cucumber/cucumber-engine/tree/master.svg?style=svg)](https://circleci.com/gh/cucumber/cucumber-engine/tree/master)

Shared go binary that can be used by all language implementations.

It takes care of loading the features, filtering the pickles, and orchestrating the test run. It defers running the hooks / steps to the caller. Its primary output is events that conform to the event protocol.

#### Links

* [Usage](./docs/usage.md)
* [Contributing](./CONTRIBUTING.md)
