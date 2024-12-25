# Open Software License Catalogue

Open Software License Catalogue is a service that catalogues software in order to make the licensing information
discoverable via a well-defined API.

## How to dev

A local version of the app can be run with the `mise run dev` command, which will
set up all the necessary dependencies via `docker-compose` and run the app.

## TODO

- 100% test coverage
- basic metrics via OTEL
  - Go Runtime metrics
  - HTTP request metrics
- basic traces via OTEL
- settings / options integration
- HTTP Healthcheck
- SLSA3 compliant release process
- Submit SBOM, get license info
- Deep license scan
  - Get license info on package dependencies
- Acceptance tests

## License

See the [LICENSE](LICENSE) file for license rights and limitations.

### Licenses.json

Regarding the [licenses.json] file included in this project:
1. It is an unmodified copy of the JSON source file from the SPDX License List repository
  (https://github.com/spdx/license-list-data) licensed under the CC-BY-3.0
  (https://creativecommons.org/licenses/by/3.0/) license.
2. The source file is maintained by the SPDX Legal Team (https://spdx.dev/engage/participate/legal-team/).
3. It is the copyright of The Linux Foundation.
4. No warranties are provided.

[licenses.json]: sll/licenses.json
