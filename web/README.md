# Contributing to Octant's UI

This document describes how to setup your development environment to contribute to Octant's UI.

**Note**: Most the code related to this development is organized under the `web/` directory.

## Getting started

### Dependencies

Our web UI is built on Node.js 24 (see `.nvmrc`) and npm 11. It is generated with [Angular CLI](https://github.com/angular/angular-cli) v22, so you'll want to install and get familiar with that tool to understand how some of our npm scripts work. Here are some of the major libraries we use:

- [Angular v21](http://angular.io)
- [TypeScript v4.3](https://www.typescriptlang.org/)
- [Clarity v12](https://clarity.design/)
- [Lodash v4](https://lodash.com/)

There are different ways to installing these dependencies:

The most thorough way that verifies that you have the needed dependencies to build/develop both the Octant's UI and go server component is run `go run build.go ci` from the root directory. This is the command used to build a final distributable.

To install just the UI dependencies from the root directory, we have the command `go run build.go web-deps` for just npm installation.

You can also run `npm install` yourself, but this will only work if you are within the `web/` directory.

Additionally, run `go run build.go web-build` to generate web assets.

### Running development mode

Once you have the necessary dependencies installed, you can start the backend and the frontend server with the following command:

    go run build.go serve

This will run both server processes in parallel.  Point your browser to `[http://localhost:4200/](http://localhost:4200/)` to connect to the frontend server.

## Directory structure

The entry file into our application is `src/main.ts` but most of our UI logic is written under `src/app/modules` with our base module being `OverviewModule`.

Here is a summary of the `app/` directory structure:

    ➜ tree ./web/src/app -L 1
    ├── components     // components living outside any modules
    ├── models         // typescript definitions of backend payloads
    ├── modules        // NgModules (https://angular.io/guide/architecture-modules)
    ├── services       // globally available service classes
    ├── testing        // testing mocks and stubs
    └── util           // app-wide reused functions

## Testing & Linting

For testing we use [Karma's Test Runner](https://karma-runner.github.io/latest/index.html) with [Jasmine](https://jasmine.github.io). To lint the codebase we rely on [TSLint](https://palantir.github.io/tslint/) & [Prettier](https://prettier.io/) to keep our codebase formatted.

There are 3 commands that can help keep PRs tested and linted properly:

- `npm test` uses Karma's Chrome launcher to open up an instance of Chrome and run our test suite against that environment. This is primarily the tool our team uses.
- `npm run test:headless` helps run our test suite against a headless version of Chrome
- `npm run lint` runs our static analysis tools against our TypeScript code

## Building production

To build a production version of the UI, you can run `npm run build` yourself which will build the productionized assets to `dist/`.

To build a full production binary (including the backend server), you can run `go run build.go ci` from the root directory.

## Building the desktop app

The desktop shell is packaged with Electron. From the repository root, this builds the Go sidecar, the renderer, and the installers:

    go run build.go dist-electron

On macOS it emits **two independent artifacts** so users can pick the smaller native build or the one that also runs on Intel:

    Octant-<version>-arm64.dmg      # Apple Silicon only (native arm64 sidecar)
    Octant-<version>-universal.dmg  # Apple Silicon + Intel (fat sidecar via lipo)

Elsewhere it emits the platform's normal installer (`.nsis` on Windows, `.AppImage` on Linux).

For a quick single-architecture build during development, `build-electron` packages one app using the Go sidecar for the host (or a `--target-arch`/`--target-os` you pass):

    go run build.go build-electron --target-arch=arm64

`--target-arch` accepts `amd64`, `arm64`, or `universal` (macOS only), and `--target-os` overrides GOOS. The sidecar has to match the packaged app, which is why `dist-electron` builds each one explicitly.

### Code signing (macOS)

electron-builder auto-discovers a signing identity from the login keychain. To pin a specific one (for example your personal Apple Development certificate), export `CSC_NAME` before building:

    CSC_NAME='Apple Development: Your Name (TEAMID)' go run build.go dist-electron

CI does not sign desktop builds; provide signing secrets there if you need distribution-signed artifacts.
