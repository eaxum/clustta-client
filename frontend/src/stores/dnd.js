import { defineStore } from "pinia";
import { useStageStore } from "@/stores/stages";
import { useAssetStore } from "@/stores/assets";
import { useEntityStore } from "@/stores/entity";
import { useProjectStore } from '@/stores/projects';

export const useDndStore = defineStore("dnd", {
  state: () => ({
    domUpdateTrigger: 0,
    dragDelay: 200,
    draggedItemId: null,
    isDropHovering: false,
    userCanDrag: true,
    importEditMode: false,
    altKeyActive: false,
    itemOverlappedId: null,
    isOverlapping: false,
    targetItemId: null,
    targetItemPath: null,
    targetItem: null,
    draggedItem: null,
    dropType: null,
    hoverTargetId: null,
    dropPromptMessage: "",
    droppedFiles: [],
    untrackedParents: [],
    trackedParents: [],
    droppedFolders: [],
    itemRefs: {},
    visibleItemRefs: {},
    intersectingItemIds: [],
    previewData: {},
    previewDataActiveItem: null,
    previewDataSelectedItems: {},
    selectedPreviewItemIds: [],
    mousePos: {
      x: -1000,
      y: -1000,
    },
    isDragging: false,
    dragPosition: {
      x: 100,
      y: 100,
    },
    ghostCardStyle: {
      leaving: false,
      pos: { x: 0, y: 0 },
      width: 0,
      cursorDistance: { x: 0, y: 0 },
      transform: "",
    },
  }),
  
  getters: {
    allElements: (state) => {
      state.domUpdateTrigger
      
      const item_class = 'virtua-item-header'
      return Array.from(document.querySelectorAll(`.${item_class}`))
    },
    allViewItems() {
      return this.allElements?.map((item) => ({
        ...JSON.parse(item.dataset.other)
      }));
    },
    // itemRefs(state) {
    //   return this.allElements?.reduce((acc, element) => {
    //     if (element.id) {
    //       acc[element.id] = element;
    //     }
    //     return acc;
    //   }, {}) || {};
    // },
    getSelectedPreviewItemIds: (state) => {
      return state.selectedPreviewItemIds;
    },
  },
  actions: {
    triggerDomUpdate() {
      this.domUpdateTrigger++
    },

    addRef(id, element) {
      this.itemRefs[id] = element;
    },

    removeRef(id) {
      delete this.itemRefs[id];
    },

    addVisibleItemsRef(id, element) {
      this.visibleItemRefs[id] = element;
    },

    removeVisibleItemsRef(id) {
      delete this.visibleItemRefs[id];
    },

    clearRefs() {
      this.itemRefs = {};
      this.visibleItemRefs = {};
    },

    onDragStart(e, id) {
      const stage = useStageStore();

      if (!this.userCanDrag || id in stage.expandedEntities) {
        return;
      }

      let selectedCard = this.itemRefs[id];
      if (!selectedCard) {
        return;
      }

      let cardRect = selectedCard.getBoundingClientRect();

      document.documentElement.style.cursor = "grabbing";

      let paddingLeft = parseFloat(getComputedStyle(selectedCard).paddingLeft);
      let paddingRight = parseFloat(
        getComputedStyle(selectedCard).paddingRight
      );

      this.mousePos.x = e.pageX;
      this.mousePos.y = e.pageY;

      this.draggedItemId = id;

      this.ghostCardStyle.width =
        selectedCard.clientWidth - paddingLeft - paddingRight;
      this.ghostCardStyle.cursorDistance.x = e.pageX - cardRect.x;
      this.ghostCardStyle.cursorDistance.y = e.pageY - cardRect.y;

      this.setGhostCardStyle(e);
      this.updateUI();
    },

    updateUI() {
      const stage = useStageStore();

      const allTargets = this.allViewItems

      this.draggedItem = allTargets.find((item) => item.id === this.draggedItemId );

      const filteredTargets = allTargets.filter((item) => item.id !== this.draggedItemId);
      
      let dragX = this.mousePos.x,
      dragY = this.mousePos.y;
      
      if (this.draggedItemId === null || this.ghostCardStyle.leaving) return;
      
      if (!dragX && !dragY) {
        return requestAnimationFrame(this.updateUI);
      }
      
      this.setGhostCardStyle(true);
      
      for (let target of filteredTargets) {
        let targetEl = this.itemRefs[target.id];
        
        if (!targetEl) {
          continue;
        }
        
        let dropZone = targetEl.querySelector(".drop-zone");

        if (!dropZone) {
          continue;
        }

        this.isOverlapping = this.checkOverlap(
          { x: dragX, y: dragY },
          dropZone.getBoundingClientRect()
        );

        if (this.isOverlapping && this.targetItemId === target.id){
          return requestAnimationFrame(this.updateUI);
        } else if (this.isOverlapping) {
          
          this.targetItemId = target.id;
          this.targetItem = target;

          // if this is an untracked item, attempt to perform a move
          if (this.targetItem.item_type) {
            this.dropType = "untracked";
          }

          // if this is an entity make the selected items children
          if (this.targetItem.entity_type_id) {
            this.dropType = "child";
          }

          // if this is a task make the selected items dependencies
          if (this.targetItem.task_type_id) {
            this.dropType = "dependency";
          }

          // else console.log('cant drop here')

          break;
        }
      }

      if (!this.isOverlapping) {
        // console.log('not overlapping')
        return requestAnimationFrame(this.updateUI);
      }

      return requestAnimationFrame(this.updateUI);
    },

    onDrag(e) {
      e = e || window.event;
      if (this.draggedItemId === null) return;
      this.mousePos.x = e.pageX;
      this.mousePos.y = e.pageY;
    },

    onDragStop(cardEl) {
      if (this.draggedItemId === null) return;
      document.documentElement.style.cursor = "default";

      let cardRect = cardEl.getBoundingClientRect();

      setTimeout(() => {
        this.resetValues();
      }, 100);

      this.ghostCardStyle.leaving = true;
      let xOffset = cardRect.x - this.ghostCardStyle.pos.x;
      let yOffset = cardRect.y - this.ghostCardStyle.pos.y;
      this.ghostCardStyle.transform = `scale(1) translate(${xOffset}px, ${yOffset}px)`;
    },

    resetValues() {
      this.droppedFolders = [];
      this.previewData = {};
      this.droppedFiles = [];
      this.targetItemId = null;
      this.targetItemPath = null;
      this.trackedParents = [];
      this.untrackedParents = [];

      this.altKeyActive = false;
      this.isOverlapping = false;
      this.dropType = null;

      this.draggedItem = null;
      this.draggedItemId = null;
      this.targetItem = null;

      this.itemOverlappedId = null;
      this.dropPromptMessage = "";
      this.ghostCardStyle.x = -1000;
      this.ghostCardStyle.y = -1000;
      this.ghostCardStyle.width = 0;
      this.ghostCardStyle.cursorDistance.x = 0;
      this.ghostCardStyle.cursorDistance.y = 0;
      this.ghostCardStyle.transform = "";
      this.ghostCardStyle.leaving = false;
    },

    checkOverlap(drag, rect) {
      if (drag.x < rect.x || drag.x > rect.x + rect.width) {
        this.targetItem = null;
        this.targetItemId = null;
        return false;
      }
      if (drag.y < rect.y || drag.y > rect.y + rect.height) {
        this.targetItem = null;
        this.targetItemId = null;
        return false;
      }
      return true;
    },

    putCardInColumn() {
      let card = cards.value.find((card) => card.id === draggedItemId.value);
      if (card) {
        card.status_id = columnOverlappedId.value;
      }
    },

    setGhostCardStyle(isDragstart, rotate=false) {
      let dragX = this.mousePos.x,
        dragY = this.mousePos.y;
      let transform = [];
      const angle = rotate ? 4 : 0;
      if (isDragstart)
        // transform.push(`scale(1.05)`);

      transform.push(`rotate(${angle}deg)`);
      this.ghostCardStyle.transform = transform.join(" ");
      this.ghostCardStyle.pos.x = dragX - this.ghostCardStyle.cursorDistance.x;
      this.ghostCardStyle.pos.y = dragY - this.ghostCardStyle.cursorDistance.y;
    },
  },
});
