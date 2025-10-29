<template>
  <div class="contribution-graph">
    <div class="graph-header">
      <span class="activity-count">{{ totalActivity }} contributions in the last year</span>
    </div>
    
    <div class="graph-container">
      <div class="months-labels">
        <span v-for="month in visibleMonths" :key="month" class="month-label">
          {{ month }}
        </span>
      </div>
      
      <div class="graph-grid">
        <div class="days-labels">
          <span class="day-label">Mon</span>
          <span class="day-label">Wed</span>
          <span class="day-label">Fri</span>
        </div>
        
        <div class="weeks-container">
          <div v-for="(week, weekIndex) in weeks" :key="weekIndex" class="week-column">
            <div
              v-for="(day, dayIndex) in week"
              :key="dayIndex"
              class="day-cell"
              :class="getActivityClass(day.count)"
              :title="getTooltip(day)"
            ></div>
          </div>
        </div>
      </div>
      
      <div class="legend">
        <span class="legend-label">Less</span>
        <div class="legend-cell level-0"></div>
        <div class="legend-cell level-1"></div>
        <div class="legend-cell level-2"></div>
        <div class="legend-cell level-3"></div>
        <span class="legend-label">More</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';

const weeks = ref([]);
const totalActivity = ref(0);

const visibleMonths = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

// Generate contribution data for the last year
const generateContributionData = () => {
  const weeksData = [];
  const today = new Date();
  const oneYearAgo = new Date(today.getFullYear() - 1, today.getMonth(), today.getDate());
  
  let currentDate = new Date(oneYearAgo);
  let week = [];
  let total = 0;
  
  // Start from the first day of the week
  const dayOfWeek = currentDate.getDay();
  const diff = dayOfWeek === 0 ? 6 : dayOfWeek - 1; // Adjust so Monday is 0
  currentDate.setDate(currentDate.getDate() - diff);
  
  while (currentDate <= today) {
    // Random activity count (0-10) - replace with real data
    const count = Math.floor(Math.random() * 11);
    total += count;
    
    week.push({
      date: new Date(currentDate),
      count: count
    });
    
    if (week.length === 7) {
      weeksData.push(week);
      week = [];
    }
    
    currentDate.setDate(currentDate.getDate() + 1);
  }
  
  // Add remaining days
  if (week.length > 0) {
    // Fill the rest with empty days
    while (week.length < 7) {
      week.push({ date: null, count: 0 });
    }
    weeksData.push(week);
  }
  
  weeks.value = weeksData;
  totalActivity.value = total;
};

const getActivityClass = (count) => {
  if (count === 0) return 'level-0';
  if (count <= 2) return 'level-1';
  if (count <= 5) return 'level-2';
  return 'level-3';
};

const getTooltip = (day) => {
  if (!day.date) return '';
  const dateStr = day.date.toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  });
  return `${day.count} contributions on ${dateStr}`;
};

onMounted(() => {
  generateContributionData();
});
</script>

<style scoped>
.contribution-graph {
  width: 100%;
  color: var(--white);
}

.graph-header {
  margin-bottom: 1rem;
}

.activity-count {
  font-size: 0.875rem;
  color: var(--white);
  opacity: 7;
}

.graph-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  overflow-x: auto;
  padding: 0.5rem 0;
}

.months-labels {
  display: flex;
  gap: 0.25rem;
  margin-left: 30px;
  margin-bottom: 0.25rem;
}

.month-label {
  font-size: 0.75rem;
  color: var(--white);
  opacity: .6;
  min-width: 60px;
  text-align: left;
}

.graph-grid {
  display: flex;
  gap: 0.5rem;
}

.days-labels {
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  gap: 4px;
  padding-right: 0.5rem;
}

.day-label {
  font-size: 0.75rem;
  color: var(--white);
  opacity: .6;
  height: 12px;
  display: flex;
  align-items: center;
}

.weeks-container {
  display: flex;
  gap: 4px;
  flex: 1;
}

.week-column {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.day-cell {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  cursor: pointer;
  transition: transform 0.1s;
}

.day-cell:hover {
  transform: scale(1.2);
}

.day-cell.level-0 {
  background-color: var(--light-steel);
}

.day-cell.level-1 {
  background-color: #9be9a8;
}

.day-cell.level-2 {
  background-color: #40c463;
}

.day-cell.level-3 {
  background-color: #216e39;
}

.legend {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 0.5rem;
  justify-content: flex-end;
}

.legend-label {
  font-size: 0.75rem;
  color: var(--white);
  opacity: .6;
}

.legend-cell {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

/* Scrollbar styling */
.graph-container::-webkit-scrollbar {
  height: 6px;
}

.graph-container::-webkit-scrollbar-thumb {
  background-color: var(--light-steel);
  border-radius: 3px;
}

.graph-container::-webkit-scrollbar-track {
  background-color: var(--dark-steel);
  border-radius: 3px;
}
</style>
