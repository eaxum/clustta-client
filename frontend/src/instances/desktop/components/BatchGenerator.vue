<template>
    
    <div class="input-group">
      <div class="field">
        <label for="input1">First Item</label>
        <input class="input-short" :class="{ 'error' : invalidPattern}"
          id="input1" v-model="input1" type="text" placeholder="e.g., shot01" />
      </div>
      
      <div class="field">
        <label for="input2">Second Item</label>
        <input class="input-short" :class="{ 'error' : invalidPattern}"
          id="input2" v-model="input2" type="text" placeholder="e.g., shot02" />
      </div>
      
      <div class="field">
        <label for="count">Total Count</label>
        <input class="input-short"
          id="count" v-model="count" type="number" placeholder="e.g., 10" min="1" />
      </div>
    </div>

    <div class="result-display" :class="{ 'no-results' : !!!sequence.length }">
        {{ sequenceString }}
    </div> 

    
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'

const emit = defineEmits([ 'update-data'])

const props = defineProps({
  icon: String,
  buttonFunction: Function,
  showIcon: { type: Boolean, default: true },
});


const input1 = ref('0010')
const input2 = ref('0020')
const count = ref('5')
const sequence = ref([])

const extractPattern = (str1, str2) => {
  if (!str1 || !str2) return null
  
  // Find the numeric part and its position
  const match1 = str1.match(/(\d+)/)
  const match2 = str2.match(/(\d+)/)
  
  if (!match1 || !match2) return null
  
  const num1 = parseInt(match1[0])
  const num2 = parseInt(match2[0])
  const numIndex = match1.index
  
  // Extract prefix and suffix
  const prefix = str1.substring(0, numIndex)
  const suffix = str1.substring(numIndex + match1[0].length)
  
  // Check if the pattern is consistent
  const expectedStr2 = prefix + num2.toString().padStart(match1[0].length, '0') + suffix
  if (expectedStr2 !== str2) return null
  
  const increment = num2 - num1
  const padding = match1[0].length
  
  return {
    prefix,
    suffix,
    startNum: num1,
    increment,
    padding
  }
}

const patternValid = computed(() => {
  return extractPattern(input1.value, input2.value) !== null
})

const invalidPattern = computed(() => {
  return input1.value && input2.value && !patternValid.value
})

defineExpose({
  invalidPattern
})



const generateSequence = () => {
  const pattern = extractPattern(input1.value, input2.value)
  if (!pattern || !count.value || parseInt(count.value) <= 0) {
    sequence.value = []
    return
  }
  
  const totalCount = parseInt(count.value)
  const items = []
  
  for (let i = 0; i < totalCount; i++) {
    const currentNum = pattern.startNum + (i * pattern.increment)
    const paddedNum = currentNum.toString().padStart(pattern.padding, '0')
    items.push(pattern.prefix + paddedNum + pattern.suffix)
  }
  
  sequence.value = items;
  emit('update-data', sequence.value)
}

const sequenceString = computed(() => {
  if (sequence.value.length === 0) return ''
  
  // Take first 4 items for display
  const displayItems = sequence.value.slice(0, 4)
  let result = displayItems.join(', ')
  
  // Add ellipsis if there are more items
  if (sequence.value.length > 4) {
    result += ', ... ' + sequence.value[sequence.value.length - 1]
  }
  
  return result
})

// Watch for changes and regenerate sequence
watch([input1, input2, count], () => {
  generateSequence()
})

onMounted(() => {
  generateSequence();
})
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
}


label {
  display: block;
  font-size: 14px;
  color: #374151;
  color: var(--white);
  margin-bottom: 8px;
}

.input-short {
  flex: 1;
  width: 100%;
}

.input-short::-webkit-scrollbar {
	width: 8px;
}

.input-short::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.input-short::-webkit-scrollbar-track {
	border-radius: 10px;
}

.result-display{
    width: 100%;
    display: flex;
    padding: .5rem;
    box-sizing: border-box;
    color: var(--white);
}

.no-results{
  padding: 0px;
}

.results {
  margin-top: 24px;
}


.error, .error:hover, .error:focus {
    outline: 1px solid crimson;
}

.sequence-item {
  background: white;
  padding: 8px 12px;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
  font-size: 14px;
  font-family: 'Courier New', monospace;
}

.pattern-info {
  margin-top: 16px;
  padding: 12px;
  background: #dbeafe;
  border-radius: 6px;
}

.pattern-info p {
  font-size: 14px;
  color: #1e40af;
  margin: 0;
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: #fef2f2;
  border-radius: 6px;
}

.error-message p {
  font-size: 14px;
  color: #dc2626;
  margin: 0;
}

@media (min-width: 640px) {
  .input-group {
    flex-direction: row;
    gap: 16px;
  }
  
  .field {
    flex: 1;
  }
  
  .sequence-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  }
}

@media (min-width: 768px) {
  .sequence-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>

