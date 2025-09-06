export default {
  updateTooltip(el, { value, modifiers }) {
    if (typeof value === "string") {
      if (typeof value === "string") {
        el._tooltipContent = value;
      } else if (value && value.text) {
        el._tooltipContent = value.text;
      }

      const tooltipElement = el._tooltipElement;
      if (tooltipElement) {
        tooltipElement.innerText = el._tooltipContent;
      }
    }
  },
  // hooks
  mounted(el, { value, dir, modifiers }) {
    const tooltipElement = document.createElement("div");
    tooltipElement.className = "tooltip";
    document.body.appendChild(tooltipElement);

    el._tooltipElement = tooltipElement;
    let timeoutId;
    el._showTooltip = (event) => {
      timeoutId = setTimeout(() => {
        tooltipElement.innerText = el._tooltipContent;
        // console.log( el._tooltipContent);
        const rect = el.getBoundingClientRect();
        const { top, left } = calculateTooltipPosition(el, event);
        tooltipElement.style.top = `${top}px`;
        tooltipElement.style.left = `${left}px`;
        if(el._tooltipContent){
          tooltipElement.style.visibility = "visible";
          tooltipElement.style.opacity = 1;
        }
      }, 500);
    };

    el._hideTooltip = (event) => {
      clearTimeout(timeoutId);
      //   tooltipElement.style.display = "none";
      tooltipElement.style.visibility = "hidden";
      tooltipElement.style.opacity = 0;
    };

    el.addEventListener("mouseenter", el._showTooltip);
    el.addEventListener("mouseleave", el._hideTooltip);

    dir.updateTooltip(el, { value, modifiers });
  },
  updated(el, { value, dir, modifiers }) {
    // console.log("a");
    dir.updateTooltip(el, { value, modifiers });
  },
  unmounted(el) {
    if (el._tooltipElement) {
      el._tooltipElement.remove();
    }
    el.removeEventListener("mouseenter", el._showTooltip);
    el.removeEventListener("mouseleave", el._hideTooltip);
  },
};

function calculateTooltipPosition(el) {
  const { left, height } = el.getBoundingClientRect();
  // console.log(el)

  const top = el.getBoundingClientRect().top;
  let trayRoot;


  if (document.querySelector(".tray-root")) {
    trayRoot = document.querySelector(".tray-root").getBoundingClientRect();
  }
  
  const windowContext = trayRoot;
  const windowWidth = windowContext?.width;
  const windowHeight = windowContext?.height;
  const tooltipElement = el._tooltipElement;
  const tooltipWidth = tooltipElement.offsetWidth;
  const tooltipHeight = tooltipElement.offsetHeight;

  let tooltipTop;
  let tooltipLeft = left + (el.offsetWidth - tooltipWidth) / 2; // Center horizontally

  // Check for right-side overflow
  if (tooltipLeft + tooltipWidth > windowWidth) {
    tooltipLeft = windowWidth - tooltipWidth - 10; 
  }

  //  Check for left side overflow
  if (tooltipLeft < 0) {
    tooltipLeft = 10;
  }

  const spaceBelow =  windowHeight - (top + height) ;
  if (spaceBelow > tooltipHeight) {
    tooltipTop = top + height + 5; 
  } else {
    tooltipTop = top - tooltipHeight - 10;
  }
  return { top: tooltipTop, left: tooltipLeft };
}
