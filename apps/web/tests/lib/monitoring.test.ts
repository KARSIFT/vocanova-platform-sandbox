import assert from "node:assert/strict";
import { test } from "node:test";

import {
  createMonitor,
  MAX_EARLY_ERRORS,
  MAX_LOAD_ATTEMPTS,
  type MonitoringSdk,
} from "../../src/lib/monitoring";

const COOLDOWN = 10_000;

function fakeSdk(initFailures = 0) {
  const captured: unknown[] = [];
  const transitions: string[] = [];
  let inits = 0;
  const sdk: MonitoringSdk = {
    init() {
      inits += 1;
      if (inits <= initFailures) throw new Error("init failed");
    },
    captureException(error) {
      captured.push(error);
    },
    captureRouterTransitionStart(href) {
      transitions.push(href);
    },
  };
  return { sdk, captured, transitions, inits: () => inits };
}

/** A manual clock and timer queue, so cooldowns are tested without sleeping. */
function fakeTime() {
  let current = 0;
  let nextId = 0;
  const timers = new Map<number, { at: number; callback: () => void }>();
  const flush = async () => {
    for (let index = 0; index < 5; index += 1)
      await new Promise((resolve) => setImmediate(resolve));
  };
  return {
    now: () => current,
    setTimer: (callback: () => void, delay: number) => {
      nextId += 1;
      timers.set(nextId, { at: current + delay, callback });
      return nextId as unknown as ReturnType<typeof setTimeout>;
    },
    clearTimer: (timer: ReturnType<typeof setTimeout>) => {
      timers.delete(timer as unknown as number);
    },
    scheduled: () => [...timers.values()].map((timer) => timer.at),
    async advance(ms: number) {
      const until = current + ms;
      for (;;) {
        const due = [...timers.entries()]
          .filter(([, timer]) => timer.at <= until)
          .sort((a, b) => a[1].at - b[1].at)[0];
        if (!due) break;
        timers.delete(due[0]);
        current = due[1].at;
        due[1].callback();
        await flush();
      }
      current = until;
      await flush();
    },
    flush,
  };
}

const errorEvent = (error: unknown) =>
  Object.assign(new Event("error"), { error });
const rejectionEvent = (reason: unknown) =>
  Object.assign(new Event("unhandledrejection"), { reason });

function trackListeners(target: EventTarget) {
  const counts = new Map<string, number>();
  const add = target.addEventListener.bind(target);
  const remove = target.removeEventListener.bind(target);
  target.addEventListener = (type, listener, options) => {
    counts.set(type, (counts.get(type) ?? 0) + 1);
    add(type, listener, options);
  };
  target.removeEventListener = (type, listener, options) => {
    counts.set(type, (counts.get(type) ?? 0) - 1);
    remove(type, listener, options);
  };
  return counts;
}

test("coming back online during the cooldown schedules one retry at the earliest allowed time", async () => {
  const time = fakeTime();
  const events = new EventTarget();
  const listeners = trackListeners(events);
  const { sdk, captured, transitions } = fakeSdk();
  let online = false;
  let loads = 0;
  const monitor = createMonitor({
    load: async () => {
      loads += 1;
      if (!online) throw new Error("offline");
      return sdk;
    },
    options: {},
    events,
    retryIntervalMs: COOLDOWN,
    now: time.now,
    setTimer: time.setTimer,
    clearTimer: time.clearTimer,
    isOnline: () => online,
  });
  monitor.listen();
  const early = new Error("before the monitor started");
  events.dispatchEvent(errorEvent(early)); // t=0: first attempt, offline
  await time.flush();
  assert.equal(loads, 1);
  // Offline: no timer spends attempts while the browser cannot fetch.
  assert.deepEqual(time.scheduled(), []);

  await time.advance(2_000);
  online = true;
  events.dispatchEvent(new Event("online")); // t=2s, inside the cooldown
  events.dispatchEvent(new Event("online")); // duplicates are coalesced
  await time.flush();
  assert.equal(loads, 1);
  assert.deepEqual(time.scheduled(), [COOLDOWN]);

  // Nothing else happens on the page; the scheduled retry still runs.
  await time.advance(COOLDOWN);
  assert.equal(loads, 2);
  assert.deepEqual(captured, [early]);
  assert.equal(monitor.pending(), 0);
  assert.deepEqual(time.scheduled(), []);
  assert.deepEqual(Object.fromEntries(listeners), {
    error: 0,
    unhandledrejection: 0,
    online: 0,
  });
  // The same SDK now takes reports and router transitions.
  const later = new Error("later");
  monitor.capture(later);
  monitor.routerTransitionStart("/progress", "push");
  assert.deepEqual(captured, [early, later]);
  assert.deepEqual(transitions, ["/progress"]);
  events.dispatchEvent(errorEvent(new Error("handled by the SDK itself")));
  assert.equal(captured.length, 2);
});

test("a signal during a failing attempt queues one retry for after the cooldown", async () => {
  const time = fakeTime();
  const events = new EventTarget();
  const { sdk, captured } = fakeSdk();
  let finishFirst: (value: MonitoringSdk) => void = () => {};
  let failFirst: (reason: unknown) => void = () => {};
  let loads = 0;
  const monitor = createMonitor({
    load: () => {
      loads += 1;
      if (loads === 1)
        return new Promise<MonitoringSdk>((resolve, reject) => {
          finishFirst = resolve;
          failFirst = reject;
        });
      return Promise.resolve(sdk);
    },
    options: {},
    events,
    retryIntervalMs: COOLDOWN,
    now: time.now,
    setTimer: time.setTimer,
    clearTimer: time.clearTimer,
    isOnline: () => true,
  });
  monitor.listen();
  void monitor.ready(); // t=0: slow first attempt
  await time.advance(1_000);
  events.dispatchEvent(new Event("online")); // recovery while it is running
  failFirst(new Error("download failed"));
  void finishFirst;
  await time.flush();
  assert.deepEqual(time.scheduled(), [COOLDOWN]);
  await time.advance(COOLDOWN);
  assert.equal(loads, 2);
  assert.equal(await monitor.ready(), sdk);
  assert.deepEqual(captured, []);
});

test("an early error after a failed start is delivered by the scheduled retry", async () => {
  const time = fakeTime();
  const events = new EventTarget();
  const { sdk, captured } = fakeSdk();
  let loads = 0;
  const monitor = createMonitor({
    load: async () => {
      loads += 1;
      if (loads === 1) throw new Error("chunk failed");
      return sdk;
    },
    options: {},
    events,
    retryIntervalMs: COOLDOWN,
    now: time.now,
    setTimer: time.setTimer,
    clearTimer: time.clearTimer,
    isOnline: () => true,
  });
  monitor.listen();
  await monitor.ready(); // the idle start fails with nothing buffered
  assert.deepEqual(time.scheduled(), []);
  await time.advance(3_000);
  const rejection = new Error("unhandled after the failed start");
  events.dispatchEvent(rejectionEvent(rejection)); // inside the cooldown
  await time.flush();
  assert.equal(loads, 1);
  assert.deepEqual(time.scheduled(), [COOLDOWN]);
  await time.advance(COOLDOWN);
  assert.equal(loads, 2);
  assert.deepEqual(captured, [rejection]);
});

test("an initialisation exception never becomes an unhandled rejection and is retried", async () => {
  const unhandled: unknown[] = [];
  const onUnhandled = (reason: unknown) => unhandled.push(reason);
  process.on("unhandledRejection", onUnhandled);
  try {
    const time = fakeTime();
    const { sdk, captured, inits } = fakeSdk(1);
    const monitor = createMonitor({
      load: async () => sdk,
      options: {},
      retryIntervalMs: COOLDOWN,
      now: time.now,
      setTimer: time.setTimer,
      clearTimer: time.clearTimer,
      isOnline: () => true,
    });
    const reported = new Error("boundary error");
    monitor.capture(reported);
    assert.equal(await monitor.ready(), null);
    await time.flush();
    // The waiting report is a reason to retry once the cooldown allows.
    assert.deepEqual(time.scheduled(), [COOLDOWN]);
    await time.advance(COOLDOWN);
    assert.equal(inits(), 2);
    assert.deepEqual(captured, [reported]);
    assert.equal(await monitor.ready(), sdk);
    await new Promise((resolve) => setImmediate(resolve));
    assert.deepEqual(unhandled, []);
  } finally {
    process.off("unhandledRejection", onUnhandled);
  }
});

test("buffering, retries and timers are bounded", async () => {
  const time = fakeTime();
  let loads = 0;
  const monitor = createMonitor({
    load: async () => {
      loads += 1;
      throw new Error("always unavailable");
    },
    options: {},
    retryIntervalMs: COOLDOWN,
    now: time.now,
    setTimer: time.setTimer,
    clearTimer: time.clearTimer,
    isOnline: () => true,
  });
  for (let index = 0; index < MAX_EARLY_ERRORS + 5; index += 1)
    monitor.capture(new Error(`error ${index}`));
  assert.equal(monitor.pending(), MAX_EARLY_ERRORS);
  await time.flush();
  // A burst inside the cooldown does not spend attempts or add timers.
  for (let index = 0; index < 5; index += 1) monitor.capture(new Error("burst"));
  await time.flush();
  assert.equal(loads, 1);
  assert.ok(time.scheduled().length <= 1);
  await time.advance(COOLDOWN * 10);
  assert.equal(loads, MAX_LOAD_ATTEMPTS);
  assert.deepEqual(time.scheduled(), []);
  monitor.capture(new Error("after giving up"));
  await time.advance(COOLDOWN * 2);
  assert.equal(loads, MAX_LOAD_ATTEMPTS);
  assert.equal(monitor.pending(), MAX_EARLY_ERRORS);
});

test("concurrent callers share one load", async () => {
  const { sdk } = fakeSdk();
  let loads = 0;
  const time = fakeTime();
  const monitor = createMonitor({
    load: async () => {
      loads += 1;
      return sdk;
    },
    options: {},
    now: time.now,
    setTimer: time.setTimer,
    clearTimer: time.clearTimer,
  });
  const [first, second] = await Promise.all([monitor.ready(), monitor.ready()]);
  assert.equal(first, sdk);
  assert.equal(second, sdk);
  assert.equal(loads, 1);
  assert.deepEqual(time.scheduled(), []);
});
