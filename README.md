# route-protection

[API Specs](https://bbdocs.monstercode.net)

# Contributors

[![Contributors](https://github.com/terra-discover/bbcrs-route-protection-lib/-/jobs/artifacts/master/raw/contributors.svg?job=report)](https://github.com/terra-discover/bbcrs-route-protection-lib/-/graphs/master)

# Developer Guide

## Do

- This library must be independent.
- Some connection to database, router, or context, must be parameterize.

## May

- Import [helper](https://lab.tog.co.id/bb/helper) directly.
- Import [migration](https://github.com/terra-discover/bbcrs-migration-lib) directly.

## Don't

- Using [viper](https://github.com/spf13/viper) directly.
