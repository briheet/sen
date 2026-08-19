// Business logic behind the example server's handlers.
const USERS = [
  { id: '1', name: 'Ada Lovelace' },
  { id: '2', name: 'Grace Hopper' },
  { id: '3', name: 'Linus Torvalds' },
];

const counters = { requests: 0, fibCalls: 0 };

function fib(n) {
  counters.fibCalls++;
  if (n < 2) return n;
  return fib(n - 1) + fib(n - 2);
}

function listUsers() {
  counters.requests++;
  return USERS.map((user) => user.id);
}

function getUser(id) {
  counters.requests++;
  return USERS.find((user) => user.id === id) || null;
}

function stats() {
  const memory = process.memoryUsage();
  const active = typeof process.getActiveResourcesInfo === 'function'
    ? process.getActiveResourcesInfo().length
    : 0;
  return {
    requests: counters.requests,
    fibCalls: counters.fibCalls,
    activeResources: active,
    heapUsed: memory.heapUsed,
    rss: memory.rss,
  };
}

module.exports = { fib, listUsers, getUser, stats };
