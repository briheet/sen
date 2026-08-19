// A tiny dependency-free router used by the example server.
function match(pattern, pathname) {
  const parts = pattern.split('/').filter(Boolean);
  const segments = pathname.split('/').filter(Boolean);
  if (parts.length !== segments.length) return null;
  const params = {};
  for (let index = 0; index < parts.length; index++) {
    if (parts[index].startsWith(':')) {
      params[parts[index].slice(1)] = decodeURIComponent(segments[index]);
    } else if (parts[index] !== segments[index]) {
      return null;
    }
  }
  return params;
}

function notFound(res) {
  res.statusCode = 404;
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify({ error: 'not found' }));
}

function route(routes, req, res) {
  const { pathname } = new URL(req.url, 'http://localhost');
  for (const entry of routes) {
    if (entry.method !== req.method) continue;
    const params = match(entry.pattern, pathname);
    if (!params) continue;
    return entry.handler(req, res, params);
  }
  return notFound(res);
}

module.exports = { route, match };
