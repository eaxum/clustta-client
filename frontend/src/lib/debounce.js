import { ref } from 'vue'

export function useDebounce(func, delay) {
  const timeoutId = ref(null)
  
  return (...args) => {
    clearTimeout(timeoutId.value)
    timeoutId.value = setTimeout(() => {
      func(...args)
    }, delay)
  }
}