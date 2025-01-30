# Open Software License Catalogue (OSLC)

Welcome to the Open Software License Catalogue (OSLC), a comprehensive and robust API designed to streamline the process
of identifying and managing software licenses. OSLC is an essential tool for developers, legal teams, and organizations
committed to ensuring compliance with software licensing requirements.

## Example Usage

```bash
grpcurl -d '{"name":"requests","distributor":"pypi"}' localhost:8080 chainalysis_oss.oslc.v1alpha.OslcService.GetPackageInfo
{
  "name": "requests",
  "version": "2.32.3",
  "license": "Apache-2.0",
  "distributionPoints": [
    {
      "name": "requests",
      "url": "https://pypi.org/project/requests/",
      "distributor": "pypi"
    }
  ]
}
```

## About OSLC

In today's complex software ecosystem, understanding and adhering to various software licenses is crucial. OSLC
addresses this need by providing a stable and efficient API that accurately determines the license(s) applicable to any
given piece of software. Our tool is meticulously crafted to support a wide range of licenses, with a primary focus on
Open Source licenses, while also accommodating proprietary licenses.

## Key Features

- **Comprehensive License Detection**: OSLC leverages advanced algorithms to identify and categorize software licenses,
  ensuring thorough and precise results.
- **Broad License Support**: From popular Open Source licenses like MIT, GPL, and Apache, to various proprietary
  licenses, OSLC covers an extensive spectrum of licensing options.
- **User-Friendly API**: Our intuitive API is designed for seamless integration, making it easy for developers to
  incorporate license detection into their projects and workflows.
- **Compliance Assurance**: By providing accurate license information, OSLC helps organizations maintain compliance with
  legal and regulatory requirements, mitigating the risk of license violations.
- **Scalability and Performance**: Built with efficiency in mind, OSLC delivers fast and reliable performance, capable
  of handling large-scale projects and diverse software inventories.

## Why Choose OSLC?

OSLC is more than just a license detection tool; it is a cornerstone for projects and utilities focused on software
licensing compliance. Whether you are managing an open-source project, developing proprietary software, or overseeing a
complex software portfolio, OSLC provides the clarity and confidence needed to navigate the intricate landscape of
software licenses.

Join the growing community of developers and organizations who trust OSLC to simplify their software licensing
processes. Embrace the power of accurate license detection and ensure your projects are compliant, secure, and legally
sound.

## Get Started with OSLC

OSLC is currently in early alpha mode, and interactions with the tool require setting up a local instance. We are
actively working on a publicly available instance, website, and comprehensive usage documentation to make it easier for
you to integrate OSLC into your projects.

To get started with OSLC, you can set up a local instance by following the instructions provided in our repository.
The protobuf schema used when communicating with the application can be found here:
[OSLC Protobuf Schema](https://buf.build/chainalysis-oss/oslc/docs/main:chainalysis_oss.oslc.v1alpha).

Stay tuned for updates as we continue to develop and enhance OSLC, making it more accessible and user-friendly for the broader community.

## How to dev

A local version of the app can be run with the `mise run dev` command, which will
set up all the necessary dependencies via `docker-compose` and run the app.

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
