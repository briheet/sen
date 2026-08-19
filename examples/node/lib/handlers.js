// HTTP handlers that turn requests into service calls.
const { fib, listUsers, getUser, stats } = require('./services');

function sendJson(res, status, body) {
  res.statusCode = status;
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify(body));
}

function healthHandler(req, res) {
  sendJson(res, 200, { status: 'ok', uptime: process.uptime() });
}

function fibHandler(req, res) {
  const n = Number(new URL(req.url, 'http://localhost').searchParams.get('n') || 20);
  const start = Date.now();
  const value = fib(n);
  sendJson(res, 200, { n, value, ms: Date.now() - start });
}

function listUsersHandler(req, res) {
  sendJson(res, 200, { users: listUsers() });
}

function getUserHandler(req, res, params) {
  const user = getUser(params.id);
  if (!user) return sendJson(res, 404, { error: 'user not found' });
  sendJson(res, 200, user);
}

function echoHandler(req, res) {
  let body = '';
  req.on('data', (chunk) => { body += chunk; });
  req.on('end', () => sendJson(res, 200, { echoed: body }));
}

function statsHandler(req, res) {
  sendJson(res, 200, stats());
}

module.exports = {
  health: healthHandler,
  fib: fibHandler,
  listUsers: listUsersHandler,
  getUser: getUserHandler,
  echo: echoHandler,
  stats: statsHandler,
};
