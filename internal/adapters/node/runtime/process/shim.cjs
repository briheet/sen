// Embedded metrics shim. Run via node --require.
const fs = require('fs');

const file = process.env.SEN_METRICS_FILE;
const intervalMs = Number(process.env.SEN_METRICS_INTERVAL_MS || 100);

if (!file) return;

function sample() {
  const mem = process.memoryUsage();
  const cpu = process.cpuUsage();
  fs.appendFileSync(file, JSON.stringify({
    heapUsed: mem.heapUsed,
    heapTotal: mem.heapTotal,
    rss: mem.rss,
    user: cpu.user,
    system: cpu.system,
  }) + '\n');
}

sample();
setInterval(sample, intervalMs);
process.on('exit', sample);
