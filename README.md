# astra

A GitHub Action which helps automate the management of GitHub pull requests for your organization

[![.github/workflows/gotest.yml](https://github.com/champ-oss/astra/actions/workflows/gotest.yml/badge.svg?branch=main)](https://github.com/champ-oss/astra/actions/workflows/gotest.yml)
[![.github/workflows/golint.yml](https://github.com/champ-oss/astra/actions/workflows/golint.yml/badge.svg?branch=main)](https://github.com/champ-oss/astra/actions/workflows/golint.yml)
[![.github/workflows/release.yml](https://github.com/champ-oss/astra/actions/workflows/release.yml/badge.svg)](https://github.com/champ-oss/astra/actions/workflows/release.yml)
[![.github/workflows/sonar.yml](https://github.com/champ-oss/astra/actions/workflows/sonar.yml/badge.svg)](https://github.com/champ-oss/astra/actions/workflows/sonar.yml)

[![SonarCloud](https://sonarcloud.io/images/project_badges/sonarcloud-black.svg)](https://sonarcloud.io/summary/new_code?id=astra_champ-oss)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=astra_champ-oss&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=astra_champ-oss)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=astra_champ-oss&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=astra_champ-oss)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=astra_champ-oss&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=astra_champ-oss)

## Features
- Automatically reruns failed workflows for pull requests
- Enables auto-merge for pull requests
- Supports processing pull requests only opened by certain actors (ex: Dependabot)

## Example Usage

```yaml
jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          token: ${{ secrets.GITHUB_TOKEN }}

      - uses: champ-oss/astra
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
```

## Token
By default the `GITHUB_TOKEN` should be passed to the `actions/checkout` step as well as this action (see example usage). This is necessary for the action to be allowed to push changes to a branch as well as open a pull request.

*Important:*

If you are syncing workflow files (`.github/workflows`) then you will need to generate and use a Personal Access Token (PAT) with `repo` and `workflow` permissions. 


## Parameters
| Parameter | Required | Description |
| --- | --- | --- |
| token | false | GitHub Token or PAT |
| repo | true | Source GitHub repo |
| files | true | List of files to sync |
| target-branch | false | Target branch for pull request |
| pull-request-branch | false | Branch to push changes |
| user | false | Git username |
| email | false | Git email |
| commit-message | false | Updated by astra |

## Contributing

