# Octant web end-to-end tests

A [Playwright](https://playwright.dev/) harness that runs the **production web
build** against a **mock Octant backend**. No Kubernetes cluster, no real
Octant process and no network access are required.

## How it works

- `mock-backend.mjs` serves `web/dist/octant` (SPA fallback to `index.html`)
  and accepts the app websocket on `/api/v1/stream`.
- On every websocket connection it pushes a deterministic fixture sequence
  (`build-info`, `kube-config`, `kube-config-path`, `navigation`,
  `namespaces`, `filters`) followed by a `content` message, all wrapped as
  `{ "type": ..., "data": ... }` exactly like the real backend.
- The pushed content is chosen per connection by the `octant-e2e-scenario`
  cookie, so specs can run in parallel without shared server state.

### Scenarios

| Scenario    | Content pushed                                            |
| ----------- | --------------------------------------------------------- |
| `table`     | datagrid with 3 rows (default)                            |
| `empty`     | datagrid with no rows and a blank `emptyContent`          |
| `alert`     | `table` content + an `event.octant.dev/alert` after ~250ms |
| `overview`  | overview page (list of cards)                             |

## Running locally

```bash
cd web

# 1. install deps (peer conflicts with storybook/rxjs require this flag)
npm ci --legacy-peer-deps

# 2. install the Playwright browser
npx playwright install chromium

# 3. build the app that the mock server serves
npm run build

# 4. run the tests (starts the mock server automatically)
npx playwright test
```

Useful variants:

```bash
npx playwright test --headed            # watch it run
npx playwright test --ui                # Playwright UI mode
npx playwright test e2e/app-shell.spec.ts
npx playwright show-report
```

`npm run e2e` is a shortcut for `npx playwright test`, and
`npm run e2e:install` installs the Chromium browser.

### Using a system Chrome

Some sandboxes block Playwright's browser download. Point the harness at an
existing Chromium/Chrome binary instead:

```bash
PLAYWRIGHT_CHROMIUM_EXECUTABLE=/path/to/chrome npx playwright test
```

In the `octant-web` Docker Sandbox the wrapper at `/home/agent/chrome.sh`
(which execs `/home/agent/chrome/chrome-linux-arm64/chrome`) can be used:

```bash
PLAYWRIGHT_CHROMIUM_EXECUTABLE=/home/agent/chrome.sh npx playwright test
```

## Fixtures

`e2e/fixtures/*.json` were captured from a real running Octant by wrapping
`window.WebSocket` with Chrome DevTools Protocol and recording every
server->client frame (a scratch capture script lives outside the repo under
`plans/_ws-capture.mjs`; it is not part of the committed harness). They are
trimmed to keep the harness fast:

- `content-table.json` keeps the `kube-system` **Pods** datagrid reduced to
  6 columns and 3 rows.
- `content-empty.json` reuses a real `table` view with `rows: []` and
  `emptyContent: ""` to exercise the empty state.
- `navigation.json` keeps all 24 real navigation sections.

## CI

`.github/workflows/e2e.yaml` runs on PRs (and pushes to `master` / `release-*`):
Node 24 + `npm ci` + `npx playwright install --with-deps chromium` +
`npm run build` + `npx playwright test`.

## Known limitation

The spec `e2e/empty-datagrid.spec.ts` checks that an empty datagrid renders
Clarity's empty placeholder region with no data rows (passing). A second,
`test.fixme`-marked assertion for the literal fallback text
`"No items to display."` is intentionally skipped: at this commit the fork's
`datagrid.component.html` replaced the `#emptyPlaceholder` template body with an
HTML comment, so a blank `emptyContent` renders **no** text at all. Change
`test.fixme` back to `test` once the fallback text is restored.
