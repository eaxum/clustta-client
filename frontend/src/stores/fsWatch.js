import { defineStore } from 'pinia';
import { ref } from 'vue';
import { Events } from '@wailsio/runtime';
import { FSService } from '@/services';

export const useFsWatchStore = defineStore('fsWatch', () => {
  const watchers = ref(new Map());
  const pendingOps = new Map();
  let initialized = false;

  // Returns the parent folder of an absolute path.
  const dirname = (p) => {
    if (!p) return '';
    const i = Math.max(p.lastIndexOf('/'), p.lastIndexOf('\\'));
    return i >= 0 ? p.slice(0, i) : '';
  };

  // Normalizes path separators and strips trailing slashes so keys match
  // regardless of source style.
  const normalize = (p) => {
    if (!p) return '';
    return p.replace(/\\/g, '/').replace(/\/+$/, '');
  };

  // Routes one fs-change event to subscribers whose folder is the changed
  // file's parent.
  const dispatch = (message) => {
    const changedPath = typeof message?.data === 'string'
      ? message.data
      : (typeof message === 'string' ? message : null);

    if (!changedPath) {
      for (const entry of watchers.value.values()) {
        for (const cb of entry.subscribers.values()) cb(null);
      }
      return;
    }

    const parent = normalize(dirname(changedPath));
    const entry = watchers.value.get(parent);
    if (!entry) return;
    for (const cb of entry.subscribers.values()) cb(changedPath);
  };

  // Queues an async op for a path so subscribe and unsubscribe can't race.
  const serialize = (path, fn) => {
    const prev = pendingOps.get(path) ?? Promise.resolve();
    const next = prev.catch(() => {}).then(fn);
    pendingOps.set(path, next);
    next.finally(() => {
      if (pendingOps.get(path) === next) pendingOps.delete(path);
    });
    return next;
  };

  // Registers the single global fs-change listener.
  const init = () => {
    if (initialized) return;
    Events.On('fs-change', dispatch);
    initialized = true;
  };

  // Adds a subscriber for a folder. The Go watcher is added on first subscribe.
  const subscribe = async (path, id, callback) => {
    if (!path || !id || typeof callback !== 'function') return;
    const key = normalize(path);

    let entry = watchers.value.get(key);
    if (!entry) {
      entry = { refCount: 0, subscribers: new Map() };
      watchers.value.set(key, entry);
    }
    if (entry.subscribers.has(id)) return;

    entry.subscribers.set(id, callback);
    entry.refCount++;

    if (entry.refCount === 1) {
      await serialize(key, async () => {
        try {
          const exists = await FSService.DirExists(path);
          if (exists) await FSService.AddWatcherFolder(path);
        } catch (err) {
          console.error('fsWatch: AddWatcherFolder failed', path, err);
        }
      });
    }
  };

  // Removes a subscriber. The Go watcher is released when the last one leaves.
  const unsubscribe = async (path, id) => {
    if (!path || !id) return;
    const key = normalize(path);
    const entry = watchers.value.get(key);
    if (!entry) return;
    if (!entry.subscribers.delete(id)) return;

    entry.refCount--;
    if (entry.refCount <= 0) {
      watchers.value.delete(key);
      await serialize(key, async () => {
        try {
          await FSService.RemoveWatcherFolder(path);
        } catch (err) {
          // ignore: folder may already be gone
        }
      });
    }
  };

  return { init, subscribe, unsubscribe };
});
