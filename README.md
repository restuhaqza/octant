# Octant

![Logo][octant-logo]

[![Build Status](https://github.com/restuhaqza/octant/actions/workflows/preflight-checks.yaml/badge.svg)](https://github.com/restuhaqza/octant/actions/workflows/preflight-checks.yaml)
![GitHub release](https://img.shields.io/github/release/restuhaqza/octant.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> A highly extensible platform for developers to better understand the complexity of Kubernetes clusters.

`restuhaqza/octant` is a **community-maintained fork** of Octant. The original project was developed by
VMware and is now archived at [vmware-tanzu/octant](https://github.com/vmware-tanzu/octant) (final
upstream release: v0.25.1). This fork continues development from that codebase. It keeps the original
Apache-2.0 license and credits the original Octant authors — see
[Relationship to upstream](#relationship-to-upstream).

Octant is a tool for developers to understand how applications run on a Kubernetes cluster. It aims to be part of the developer's toolkit for gaining insight and approaching complexity found in Kubernetes. Octant offers a combination of introspective tooling, cluster navigation, and object management along with a plugin system to further extend its capabilities.

## Features

* **Resource Viewer**

    Graphically visualize relationships between objects in a Kubernetes cluster. The status of individual objects are represented by color to show workload performance.

* **Summary View**

    Consolidated status and configuration information in a single page aggregated from output typically found using multiple kubectl commands.

* **Port Forward**

    Forward a local port to a running pod with a single button for debugging applications and even port forward multiple pods across namespaces.

* **Log Stream**

    View log streams of pod and container activity for troubleshooting or monitoring without holding multiple terminals open.

* **Label Filter**

    Organize workloads with label filtering for inspecting clusters with a high volume of objects in a namespace.

* **Cluster Navigation**

   Easily change between namespaces or contexts across different clusters. Multiple kubeconfig files are also supported.

 * **Plugin System**

   Highly extensible plugin system for users to provide additional functionality through gRPC. Plugin authors can add components on top of existing views.

## Usage

![Octant demo](web/src/assets/octant-demo.gif)

## Installation

### Download a Pre-built Binary (Linux, macOS, Windows)

Pre-built binaries for this fork are published on the
[releases page](https://github.com/restuhaqza/octant/releases). Archives are provided for
`Linux`, `macOS`, and `Windows` on `64bit`, `ARM`, and `ARM64`.

1. Download the archive for your platform from the
   [latest release](https://github.com/restuhaqza/octant/releases/latest).

2. Extract the tarball or zip, where `X.Y` is the release version:

    ```sh
    $ tar -xzvf ~/Downloads/octant_0.X.Y_macOS-64bit.tar.gz
    octant_0.X.Y_macOS-64bit/README.md
    octant_0.X.Y_macOS-64bit/octant
    ```

3. Verify it runs:

    ```sh
    $ ./octant_0.X.Y_macOS-64bit/octant version
    ```

### Build from Source

Requirements: [Go 1.24+](https://golang.org/dl/) and [Node.js](https://nodejs.org/en/).

```sh
git clone https://github.com/restuhaqza/octant.git
cd octant
go run build.go go-install   # install Go dependencies
go run build.go ci-quick     # build UI, generate UI files, and create the octant binary
./build/octant               # run the binary you just built
```

See the [hacking guide](HACKING.md) for the full development setup.

> **Package managers:** `brew install octant`, `choco install octant`, and `scoop install octant`
> still resolve to the archived upstream project and will install the old v0.25.x build. Use the
> release binaries above (or build from source) to run this fork.

## Nightly Builds

Releases of this fork are published as versioned binaries on the
[releases page](https://github.com/restuhaqza/octant/releases). The upstream project previously
published nightly builds to a Google Cloud Storage bucket; that pipeline depended on upstream
infrastructure and is no longer maintained, so nightly builds are not currently produced here.
Use the latest tagged release instead.

## Getting Started

Before starting Octant, make sure you have access to a healthy cluster. If kubectl is installed, test using `kubectl cluster-info`.

Start running Octant:

`$ octant`

Octant should immediately launch your default web browser on `127.0.0.1:7777`.

Octant uses the default web browser on the system to act as the UI client. In the future Octant will ship with a UI.

For setting extra configuration such as what kubeconfig or context to use at startup, refer to the upstream [documentation](https://reference.octant.dev/) (still valid for the plugin API).

## Supported Versions

Octant versions follow [Semantic Versioning](https://semver.org/) where a given version number represents `MAJOR.MINOR.PATCH`.

Patch releases address bug fixes, regressions, and small enhancements.

Minor releases contain security fixes, API changes, and significant enhancements such as UI changes or new components.

Major releases contain breaking changes that are not guaranteed to be backwards compatible. Octant versions before 1.0 should not be considered stable and API may change between minor releases.

### Supported Version Skew

Version of Octant are compiled against a version of client-go.

Octant follows an `n±1` policy for versions of Kubernetes similar to kubectl. For example, Octant `0.26.0` uses the Kubernetes 1.34 client. So version `0.26.0` can be used with Kubernetes 1.33, 1.34, and 1.35.

## Plugins

Plugins are a core part of Octant in the Kubernetes ecosystem. A plugin can read objects and allows users to add components to Octant's views.

An example plugin can be found in [`cmd/octant-sample-plugin`](cmd/octant-sample-plugin) and installed to the default plugin path with `go run build.go install-test-plugin`.

Some plugins can be found on GitHub in the [`#octant-plugin`](https://github.com/topics/octant-plugin) topic (tag).

Documentation for plugin components can be found in the [Plugins section](https://reference.octant.dev/?path=/docs/docs-plugins-1-getting-started--page) of the documentation.

## Discussion

Feature requests, bug reports, and enhancements for this fork are welcome.

 - [GitHub issues](https://github.com/restuhaqza/octant/issues) — report bugs or request features here
 - [GitHub discussions](https://github.com/restuhaqza/octant/discussions) — for questions and longer conversations

The upstream project also used Kubernetes Slack ([#octant](https://kubernetes.slack.com/app_redirect?channel=CM37M9FCG)), a Google group, and Twitter ([@projectoctant](https://twitter.com/projectoctant)); those channels are associated with the archived upstream project and may be inactive.

## Relationship to upstream

This repository is a fork of [vmware-tanzu/octant](https://github.com/vmware-tanzu/octant). VMware
[archived the original project](https://github.com/vmware-tanzu/octant) after the v0.25.1 release;
this fork continues from that codebase under the same [Apache-2.0 license](LICENSE) and retains the
original copyright and NOTICE attribution.

What that means in practice:

* Development and releases happen here, at [restuhaqza/octant](https://github.com/restuhaqza/octant).
* The Go module path remains `github.com/vmware-tanzu/octant` for compatibility with existing plugin
  imports; this is intentional and not an indication that upstream is active.
* Historical changelogs under [`changelogs/`](changelogs/) and blog posts under [`site/`](site/)
  reference the original project and are preserved as-is for historical accuracy.
* Plugin API documentation at [reference.octant.dev](https://reference.octant.dev/) is the upstream
  resource and remains the reference for the plugin API this fork implements.

## Contributing

Contributors will need to sign a DCO (Developer Certificate of Origin) with all changes. We also ask that a changelog entry is included with your pull request. Details are described in our [contributing](CONTRIBUTING.md) documentation.

See our [hacking](HACKING.md) guide for getting your development environment setup.

See our [roadmap](ROADMAP.md) for tentative features in a 1.0 release.

## License

Octant is available under the [Apache License, Version 2.0](LICENSE)

[octant-logo]: /web/src/assets/octant-logo.png
