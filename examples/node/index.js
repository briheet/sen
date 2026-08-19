// Senbon example: a small but real HTTP server to profile and visualize.
//
// Run it directly:     npm start
// Run it under senbon: npm i && senbon run node ./examples/node
const http = require('http');
const { route } = require('./lib/router');
const handlers = require('./lib/handlers');

const routes = [
  { method: 'GET', pattern: '/health', handler: handlers.health },
  { method: 'GET', pattern: '/fib', handler: handlers.fib },
  { method: 'GET', pattern: '/users', handler: handlers.listUsers },
  { method: 'GET', pattern: '/users/:id', handler: handlers.getUser },
  { method: 'POST', pattern: '/echo', handler: handlers.echo },
  { method: 'GET', pattern: '/stats', handler: handlers.stats },
];

function createServer() {
  return http.createServer((req, res) => route(routes, req, res));
}

function main() {
  const port = Number(process.env.PORT || 8080);
  const server = createServer();
  server.listen(port, () => {
    console.log(`example server listening on http://localhost:${port}`);
  });
}

main();
