## v0.27.0

#### 2026-10-09

### Download

- https://github.com/restuhaqza/octant/releases/tag/v0.27.0

### All Changes

- Modernized the web client toolchain: Angular 12 → Angular 22, Clarity 18, TypeScript 6.0, ESLint 9, and Node.js 24, and upgraded Electron to 44. The Clarity v18 migration removed the legacy `@cds/core` and `@clr/icons` packages and the `@webcomponents/custom-elements` polyfill.
- Fixed views that went stale or hung on "Loading..." after the framework upgrade, caused by Angular 22 defaulting to `OnPush` change detection; components fed by websocket and service events are now marked for check, and dynamically created views are updated through `setInput`.
- Replaced Clarity icon shapes that no longer exist in v18 (which rendered as animated loading dots) with valid shapes, and registered the Clarity v18 icon collections.
- Fixed subscription leaks in the view container, tabs, and preferences components, and dropped the unused `@angular/elements` dependency.
- Webhooks and other asynchronous view updates now render on the first push instead of waiting for a user interaction.
- Added object status for Kubernetes `Node` (readiness, memory/disk/PID pressure, cordoned state, roles, capacity), `HorizontalPodAutoscaler` in `autoscaling/v2` (scaling conditions plus target and replica properties), `GatewayClass` and `GRPCRoute` (completing the Gateway API status coverage), and `MutatingWebhookConfiguration` / `ValidatingWebhookConfiguration` (webhook count with a warning when `failurePolicy=Fail` can reject requests).
- The datagrid now shows its loading indicator and a proper empty-state placeholder instead of a blank grid, and content pages with no views render a "Page not found" view with a "Go back" action instead of an endless spinner.
- Added a notification center: concurrent warnings, errors, and successes stack as dismissible toasts (with "Dismiss all") and are kept in a history panel opened from a bell in the header with an unread-style badge.
- Added a Playwright end-to-end test harness that runs the production web build against a mock backend (no cluster required), covering the app shell, datagrid rendering, empty states, and notifications, and wired it into CI.
- Repaired the CI pipeline, which had never run successfully: frontend installs now use legacy peer resolution (recorded in `web/.npmrc`), the Karma job installs Go, a `pkg/dash` runner test no longer panics and cancels unrelated jobs, and the artifact actions were upgraded to v4.
