// src/directives.js

import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";

export const stopPropagation = {
  beforeMount(el, binding) {
    const eventType = binding.arg || "click";
    el.addEventListener(eventType, (event) => event.stopPropagation());
  },
};

export const rightClick = {
  beforeMount(el, binding) {
    el.addEventListener('contextmenu', (event) => {
      event.preventDefault();
      event.stopPropagation();
      if (binding.value) {
        binding.value(event);
      }
    });
  },
  onBeforeUnmount(el, binding) {
    el.removeEventListener('contextmenu', binding.value);
  }
};

const getHighestZIndexElement = () => {
  const selector = '.checkpoint-item, .list-tags, .tag-parent, .tray-overlay-mask, .pop-up-container, .data-name-input, .dash-board-root, .tray-home, .combo-new-chip, .task-item-menu-container, .task-list-container, .task-list-root, .modal-mask';
  const elements = document.querySelectorAll(selector);
  let highestZElement = null;
  let highestZIndex = -Infinity;

  elements.forEach((element) => {
    const zIndex = window.getComputedStyle(element).zIndex;
    if (!isNaN(zIndex) && parseInt(zIndex) > highestZIndex) {
      highestZIndex = parseInt(zIndex);
      highestZElement = element;
    }
  });

  return highestZElement;
};

export const escDirective = {
  beforeMount(el, binding) {
    el.handleEscKey = (event) => {
      // el.classList.add() = "escape";
      // console.log(el.classList);
      if (event.key === "Escape") {
        const highestZIndexElement = getHighestZIndexElement();
          if (highestZIndexElement === el) {
            binding.value(event);
          }
      }
    };

    window.addEventListener("keydown", el.handleEscKey);
  },
  beforeUnmount(el) {
    window.removeEventListener("keydown", el.handleEscKey);
  },
};

export const returnDirective = {
  beforeMount(el, binding) {
    el.handleReturnKey = (event) => {
      if (event.key === "Enter") {
        binding.value(event);
        event.stopPropagation();
      }
      // if (event.key === "Enter") {
      //   const highestZIndexElement = getHighestZIndexElement();
      //     if (highestZIndexElement === el) {
      //       binding.value(event);
      //     }
      // }
    };
    window.addEventListener("keydown", el.handleReturnKey);
  },
  beforeUnmount(el) {
    window.removeEventListener("keydown", el.handleReturnKey);
  },
};

export const focusDirective = {
  beforeMount(el) {
    // Use nextTick to ensure the element is fully in the DOM before focusing
    nextTick(() => {
      if ((el && el.tagName === "INPUT") || "TEXTAREA") {
        el.focus();
      }
    });
  },
};

export const tooltip = {
  mounted(el, binding) {
    const tooltipText = binding.arg;
    let timeoutId;

    const startTooltipTimer = (event) => {
      timeoutId = setTimeout(() => {
        const tooltip = el.querySelector(".tooltip");
        if (tooltip) {
          const { top, left } = calculateTooltipPosition(el, event);
          tooltip.style.top = `${top}px`;
          tooltip.style.left = `${left}px`;
          tooltip.classList.add("visible");
        }
      }, 500); // Adjust time to 2000 milliseconds (2 seconds)
    };

    const clearTooltipTimer = () => {
      clearTimeout(timeoutId);
      const tooltip = el.querySelector(".tooltip");
      if (tooltip) {
        tooltip.classList.remove("visible");
      }
    };

    el.addEventListener("mouseenter", startTooltipTimer);
    el.addEventListener("mouseleave", clearTooltipTimer);

    // onBeforeUnmount(() => {
    //   el.removeEventListener('mouseenter', startTooltipTimer);
    //   el.removeEventListener('mouseleave', clearTooltipTimer);
    // });
  },
};

import { useTrayStates } from "@/stores/TrayStates";

function calculateTooltipPosition(el) {
  const { left, height } = el.getBoundingClientRect();

  const trayStates = useTrayStates();
  const isPopupOpen = trayStates.isPopupOpen;
  const top = isPopupOpen ? el.offsetTop : el.getBoundingClientRect().top;
  let popUp;
  let trayRoot;
  console.log('top ' + top)

  if (document.querySelector(".pop-up-container")) {
    popUp = document.querySelector(".pop-up-container").getBoundingClientRect();
  }

  if (document.querySelector(".tray-root")) {
    trayRoot = document.querySelector(".tray-root").getBoundingClientRect();
  }
  const windowContext = isPopupOpen ? popUp : trayRoot;
  const windowWidth = windowContext.width;
  const windowHeight = windowContext.height;

  const tooltipWidth = el.querySelector(".tooltip").offsetWidth;
  const tooltipHeight = el.querySelector(".tooltip").offsetHeight;

  let tooltipTop;
  let tooltipLeft = left + (el.offsetWidth - tooltipWidth) / 2; // Center horizontally

  // Check for right-side overflow
  if (tooltipLeft + tooltipWidth > windowWidth) {
    tooltipLeft = windowWidth - tooltipWidth - 10; // Position to the left
  }

  //  Check for left side overflow
  if (tooltipLeft < 0) {
    tooltipLeft = 10;
  }

  // Check if there's enough space below
  const spaceBelow = windowHeight - (top + height);
  if (spaceBelow > tooltipHeight) {
    tooltipTop = top + height + 5; // Position below
  } else {
    // Not enough space below, position above
    tooltipTop = top - tooltipHeight - 10;
  }

  return { top: tooltipTop, left: tooltipLeft };
}
