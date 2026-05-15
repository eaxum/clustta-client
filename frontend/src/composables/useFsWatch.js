import { onBeforeUnmount, watch } from 'vue';
import { useFsWatchStore } from '@/stores/fsWatch';

let nextId = 1;

// Subscribes the calling component to fs-change events for a folder.
// Subscription is moved when pathRef changes and cleaned up on unmount.
export function useFsWatch(pathRef, callback, { debounce = 0 } = {}) {
  const store = useFsWatchStore();
  store.init();

  const id = `fsw-${nextId++}`;
  let timer = null;
  let activePath = null;

  const fire = (changedPath) => {
    if (debounce > 0) {
      clearTimeout(timer);
      timer = setTimeout(() => callback(changedPath), debounce);
    } else {
      callback(changedPath);
    }
  };

  watch(
    () => (typeof pathRef === 'function' ? pathRef() : pathRef?.value),
    async (next) => {
      if (next === activePath) return;
      const prev = activePath;
      activePath = next || null;
      if (prev) await store.unsubscribe(prev, id);
      if (activePath) await store.subscribe(activePath, id, fire);
    },
    { immediate: true }
  );

  onBeforeUnmount(async () => {
    clearTimeout(timer);
    if (activePath) await store.unsubscribe(activePath, id);
  });
}
