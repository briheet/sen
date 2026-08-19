// Embedded metrics shim. Sen reads this snapshot through the Node inspector.
const { performance, monitorEventLoopDelay } = require('node:perf_hooks');

const delay = monitorEventLoopDelay({ resolution: 20 });
let previousUtilization = performance.eventLoopUtilization();
delay.enable();

Object.defineProperty(globalThis, Symbol.for('sen.metrics'), {
  value() {
    const memory = process.memoryUsage();
    const utilization = performance.eventLoopUtilization(previousUtilization);
    previousUtilization = performance.eventLoopUtilization();
    const finite = value => Number.isFinite(value) ? value : 0;
    const result = {
      heapUsed: memory.heapUsed,
      heapTotal: memory.heapTotal,
      external: memory.external,
      arrayBuffers: memory.arrayBuffers,
      eventLoopUtilization: utilization.utilization,
      eventLoopDelayMean: finite(delay.mean),
      eventLoopDelayMax: finite(delay.max),
      eventLoopDelayP95: finite(delay.percentile(95)),
      eventLoopDelayP99: finite(delay.percentile(99)),
      activeResources: typeof process.getActiveResourcesInfo === 'function'
        ? process.getActiveResourcesInfo().length
        : 0,
    };
    delay.reset();
    return result;
  },
});
