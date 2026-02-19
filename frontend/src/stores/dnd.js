import { defineStore } from "pinia";
import { useStageStore } from "@/stores/stages";
import { useCommonStore } from '@/stores/common';
import { useAssetStore } from "@/stores/assets";
import { useCollectionStore } from "@/stores/collections";
import { useProjectStore } from '@/stores/projects';
import { DragService } from '@/services';

export const useDndStore = defineStore("dnd", {
  state: () => ({
    domUpdateTrigger: 0,
    dragDelay: 200,
    draggedItemId: null,
    isDropHovering: false,
    userCanDrag: true,
    importEditMode: false,
    lockUI: true,
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
    nativeDragInitiated: false,
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
      const commonStore = useCommonStore();

      if (this.lockUI || !this.userCanDrag || (!commonStore.useGrid && id in stage.expandedEntities)) {
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
      console.log('[DragDebug] Drag started, draggedItemId:', id, 'filePath:', this.getDraggedFilePaths());

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
      // DEBUG: Verify onDrag is called (will spam console during drag)
      if (this.draggedItemId !== null) {
        // Only log once per drag session
        if (!this._dragLoggedOnce) {
          console.log('[DragDebug] onDrag active, draggedItemId:', this.draggedItemId);
          this._dragLoggedOnce = true;
        }
      }
      if (this.draggedItemId === null) return;
      this.mousePos.x = e.pageX;
      this.mousePos.y = e.pageY;
      
      // Check if cursor is at/beyond window boundary for native drag-out
      this.checkWindowBoundary(e);
    },
    
    // Detects when drag exits window and initiates native OS drag.
    async checkWindowBoundary(e) {
      if (this.nativeDragInitiated || this.draggedItemId === null) return;
      
      const edgeThreshold = 20; // pixels from edge to trigger (needs buffer before mouse leaves webview)
      const debugThreshold = 50; // log when within this distance of edge
      const { clientX, clientY } = e;
      const windowWidth = window.innerWidth;
      const windowHeight = window.innerHeight;
      
      // Calculate distances to each edge
      const distLeft = clientX;
      const distTop = clientY;
      const distRight = windowWidth - clientX;
      const distBottom = windowHeight - clientY;
      const minDist = Math.min(distLeft, distTop, distRight, distBottom);
      
      // DEBUG: Only log when approaching edge (within 50px)
      if (minDist <= debugThreshold) {
        console.log('[DragDebug] Near edge - distances:', { left: distLeft, top: distTop, right: distRight, bottom: distBottom }, 'min:', minDist);
      }
      
      const isAtEdge = (
        clientX <= edgeThreshold ||
        clientY <= edgeThreshold ||
        clientX >= windowWidth - edgeThreshold ||
        clientY >= windowHeight - edgeThreshold
      );
      
      if (!isAtEdge) return;
      
      // Get file paths of all selected/dragged items
      const filePaths = this.getDraggedFilePaths();
      console.log('[DragDebug] AT EDGE! filePaths:', filePaths);
      
      if (filePaths.length === 0) {
        console.log('[DragDebug] No file paths found, aborting');
        return;
      }
      
      // Mark as initiated to prevent multiple calls
      this.nativeDragInitiated = true;
      console.log('[DragDebug] Starting native drag with paths:', filePaths);
      
      try {
        // Cancel internal drag visuals
        document.documentElement.style.cursor = 'default';
        this.ghostCardStyle.leaving = true;
        
        // Start native OS drag - this will capture the mouse
        const result = await DragService.StartNativeDrag(filePaths);
        console.log('[DragDebug] Native drag completed, result:', result);
      } catch (err) {
        console.error('[DragDebug] Native drag failed:', err);
      } finally {
        // Reset drag state after native drag completes/cancels
        this.resetValues();
      }
    },
    
    // Returns file paths for all currently dragged items.
    getDraggedFilePaths() {
      const stage = useStageStore();
      const allItems = this.allViewItems || [];
      
      // Get marked items or just the dragged item
      const draggedIds = stage.markedItems?.length > 0 
        ? stage.markedItems 
        : [this.draggedItemId];
      
      const filePaths = [];
      for (const id of draggedIds) {
        const item = allItems.find(i => i.id === id);
        if (item?.file_path) {
          filePaths.push(item.file_path);
        }
      }
      return filePaths;
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
      this.nativeDragInitiated = false;
      this._dragLoggedOnce = false; // DEBUG

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

    setGhostCardStyle(isDragstart, rotate=false, centerOnCursor=true) {
      let dragX = this.mousePos.x,
        dragY = this.mousePos.y;
      let transform = [];
      const angle = rotate ? 4 : 0;
      if (isDragstart)
        // transform.push(`scale(1.05)`);

      transform.push(`rotate(${angle}deg)`);
      this.ghostCardStyle.transform = transform.join(" ");
      
      if (centerOnCursor) {
        this.ghostCardStyle.pos.x = dragX;
        this.ghostCardStyle.pos.y = dragY;
      } else {
        this.ghostCardStyle.pos.x = dragX - this.ghostCardStyle.cursorDistance.x;
        this.ghostCardStyle.pos.y = dragY - this.ghostCardStyle.cursorDistance.y;
      }
    },
  },
});

