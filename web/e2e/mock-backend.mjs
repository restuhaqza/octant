// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

/**
 * Mock Octant backend for the Playwright e2e harness.
 *
 * Serves the production web build (web/dist/octant) with SPA fallback and
 * accepts the app's websocket on /api/v1/stream. On every connection it pushes
 * a deterministic sequence of fixtures and never talks to a real cluster.
 *
 * The pushed content varies by "scenario", selected per-connection from the
 * `octant-e2e-scenario` cookie (so tests can run in parallel) or from the
 * `--scenario` CLI flag / OCTANT_E2E_SCENARIO env var (manual runs).
 *
 * Scenarios: table (default), empty, alert, overview.
 *
 * Usage:
 *   node e2e/mock-backend.mjs [--port 4321] [--dist dist/octant] [--scenario table]
 */
import { createServer } from 'node:http';
import { existsSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import * as wsNamespace from 'ws';

// ws 8 exposes `WebSocketServer`, ws 7 exposes `Server` (also as a property on
// the default export). Support both so the harness works with either version.
const WebSocketServer =
  wsNamespace.WebSocketServer ||
  wsNamespace.Server ||
  (wsNamespace.default && wsNamespace.default.Server);

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function argValue(flag) {
  const i = process.argv.indexOf(flag);
  return i !== -1 ? process.argv[i + 1] : undefined;
}

const PORT = Number(
  argValue('--port') || process.env.OCTANT_E2E_PORT || 4321
);
const DIST = path.resolve(
  argValue('--dist') ||
    process.env.OCTANT_E2E_DIST ||
    path.join(process.cwd(), 'dist', 'octant')
);
const FIXTURES_DIR = path.join(__dirname, 'fixtures');
const WS_PATH = '/api/v1/stream';
const COOKIE = 'octant-e2e-scenario';

let defaultScenario =
  argValue('--scenario') || process.env.OCTANT_E2E_SCENARIO || 'table';

const readJSON = name =>
  JSON.parse(readFileSync(path.join(FIXTURES_DIR, name), 'utf8'));

// ---- fixtures -------------------------------------------------------------
const BUILD_INFO = readJSON('build-info.json');
const KUBE_CONFIG = readJSON('kube-config.json');
const KUBE_CONFIG_PATH = readJSON('kube-config-path.json');
const NAVIGATION = readJSON('navigation.json');
const NAMESPACES = readJSON('namespaces.json');
const FILTERS = readJSON('filters.json');
const CONTENT_TABLE = readJSON('content-table.json');
const CONTENT_EMPTY = readJSON('content-empty.json');
const CONTENT_OVERVIEW = readJSON('content-overview.json');
const ALERT = readJSON('alert.json');

// ---- pushing ---------------------------------------------------------------
function send(socket, type, data) {
  if (socket.readyState === socket.OPEN) {
    socket.send(JSON.stringify({ type, data }));
  }
}

function pushBase(socket) {
  send(socket, 'event.octant.dev/kubeConfig', KUBE_CONFIG);
  send(socket, 'event.octant.dev/buildInfo', BUILD_INFO);
  send(socket, 'event.octant.dev/kubeConfigPath', KUBE_CONFIG_PATH);
  send(socket, 'event.octant.dev/navigation', NAVIGATION);
  send(socket, 'event.octant.dev/namespaces', NAMESPACES);
  send(socket, 'event.octant.dev/filters', FILTERS);
}

function scenarioContent(scenario) {
  switch (scenario) {
    case 'empty':
      return CONTENT_EMPTY;
    case 'overview':
      return CONTENT_OVERVIEW;
    case 'alert':
    case 'table':
    default:
      return CONTENT_TABLE;
  }
}

// ---- static file serving ---------------------------------------------------
const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.map': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.txt': 'text/plain; charset=utf-8',
};

function serveFile(res, filePath) {
  const body = readFileSync(filePath);
  res.writeHead(200, {
    'content-type': MIME[path.extname(filePath)] || 'application/octet-stream',
    'cache-control': 'no-store',
  });
  res.end(body);
}

function serveStatic(pathname, res) {
  const indexFile = path.join(DIST, 'index.html');
  if (!existsSync(indexFile)) {
    res.writeHead(500, { 'content-type': 'text/plain; charset=utf-8' });
    res.end(`web build not found at ${DIST}; run "npm run build" first`);
    return;
  }

  const rel = pathname === '/' ? 'index.html' : pathname.replace(/^\/+/, '');
  const candidate = path.resolve(DIST, rel);

  // never escape DIST
  if (
    candidate.startsWith(DIST) &&
    existsSync(candidate) &&
    statSync(candidate).isFile()
  ) {
    serveFile(res, candidate);
    return;
  }

  // SPA fallback (hash routing: everything non-asset resolves to index.html)
  serveFile(res, indexFile);
}

// ---- servers ---------------------------------------------------------------
const server = createServer((req, res) => {
  const url = new URL(req.url, `http://${req.headers.host || 'localhost'}`);

  if (url.pathname === '/healthz') {
    res.writeHead(200, { 'content-type': 'text/plain; charset=utf-8' });
    res.end('ok');
    return;
  }

  if (url.pathname === '/__scenario') {
    const name = url.searchParams.get('name');
    if (name) {
      defaultScenario = name;
    }
    res.writeHead(200, { 'content-type': 'application/json' });
    res.end(JSON.stringify({ scenario: defaultScenario }));
    return;
  }

  serveStatic(url.pathname, res);
});

const wss = new WebSocketServer({ noServer: true });

server.on('upgrade', (req, socket, head) => {
  let pathname = '/';
  try {
    pathname = new URL(req.url, `http://${req.headers.host || 'localhost'}`)
      .pathname;
  } catch {
    socket.destroy();
    return;
  }
  if (pathname !== WS_PATH) {
    socket.destroy();
    return;
  }
  wss.handleUpgrade(req, socket, head, ws => wss.emit('connection', ws, req));
});

function scenarioFor(req) {
  const cookie = req.headers.cookie || '';
  const match = cookie.match(
    new RegExp(`(?:^|;\\s*)${COOKIE}=([^;]+)`)
  );
  return match ? decodeURIComponent(match[1]) : defaultScenario;
}

wss.on('connection', (ws, req) => {
  const scenario = scenarioFor(req);
  const content = scenarioContent(scenario);
  ws.scenario = scenario;

  // A tiny delay lets the rxjs WebSocketSubject subscription settle.
  setTimeout(() => {
    if (ws.readyState !== ws.OPEN) {
      return;
    }
    pushBase(ws);
    send(ws, 'event.octant.dev/content', content);
    if (scenario === 'alert') {
      setTimeout(() => send(ws, 'event.octant.dev/alert', ALERT), 250);
    }
  }, 30);

  ws.on('message', raw => {
    let message;
    try {
      message = JSON.parse(raw.toString());
    } catch {
      return;
    }
    // Re-issue content for the requested path so in-app navigation also works.
    if (
      message.type === 'action.octant.dev/setContentPath' &&
      message.payload &&
      message.payload.contentPath
    ) {
      send(ws, 'event.octant.dev/content', {
        ...content,
        contentPath: message.payload.contentPath,
      });
    }
  });
});

server.listen(PORT, '127.0.0.1', () => {
  console.log(
    `[mock-backend] http://127.0.0.1:${PORT} dist=${DIST} scenario=${defaultScenario}`
  );
});
