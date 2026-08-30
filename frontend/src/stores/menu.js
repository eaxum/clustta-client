import { defineStore } from "pinia";
import { nextTick } from "vue";

export const useMenu = defineStore("useMenu", {
  state: () => ({
    menuStates: {
      typeFilterMenu: false,
      dependencySearchFilterMenu: false,
      assetTypeFilterMenu: false,
      resourceTypeFilterMenu: false,
      collectionTypeFilterMenu: false,
      statusFilterMenu: false,
      stateFilterMenu: false,
      extensionFilterMenu: false,
      exportColumnsMenu: false,
      tagsFilterMenu: false,
      assigneeFilterMenu: false,

      projectMenu: false,
      projectItemMenu: false,
      collectionMenu: false,
      untrackedItemMenu: false,
      assetMenu: false,
      selectionMenu: false,
      resourceItemMenu: false,
      assignMenu: false,
      manageTagsMenu: false,
      checkpointTagMenu: false,
      accountMenu: false,
      copyToProjectSubMenu: false,
      moveToCollectionSubMenu: false,
      sortMenu: false,
      viewMenu: false,
      compactEditMenu: false,
    },

    // Sub-menu navigation state
    subMenuState: {
      active: false,
      sourceMenu: null, // The menu that triggered the sub-menu
      navigationStack: [], // Stack of { type: 'projects' | 'collections', projectUri?: string, parentId?: string, title: string }
      selectedProject: null, // The project being navigated into
      selectedAssetIds: [], // Assets being moved (for move-to-collection)
      startingCollectionId: '', // Starting collection for navigation (for move-to-collection)
      slideDirection: 'left', // 'left' for entering deeper, 'right' for going back
    },

    activeMenu: null,
    nonFilterMenus: [
      'projectMenu',
      'projectItemMenu',
      'collectionMenu', 
      'assetMenu', 
      'selectionMenu',
      'untrackedItemMenu',
      'resourceItemMenu', 
      'assignMenu',
      'manageTagsMenu',
      'checkpointTagMenu',
      'accountMenu',
      'sortMenu',
      'viewMenu',
      'compactEditMenu'
    ],

    compactEditMenuData: {
      key: '',
      title: '',
      loading: false,
      options: [],
      selectedId: '',
      onSelect: null,
    },

    checkpointTagMenuData: {
      checkpoint: null,
    },

    menuEl: null,
    position: { x: 0, y: 0 },
    anchorMenu: false,
    isAnimating: false,
    contextMenuVisible: false,
    assetPopUpMenuVisible: false,
    popupTrigger: null,

    popUpMenuTop: 0,
    popUpMenuLeft: 0,
    popUpMenuWidth: 0,
    popUpMenuHeight: 0,

    clickOutsideMask: null,
    listBoxExpanded: false,
    contextMenuBounds: null,
    contextMenu: null,

    popUpMenu: null,
    assetListContainer: null,

    flyoutTop: 0,
    flyoutLeft: 0,
    flyoutWidth: 0,
    showStatusOptions: false,
  }),

  getters: {
    getContextMenuBounds: (state) => {
      return state.contextMenuBounds || document.body;
    }
  },
  actions: {
    async showContextMenu(event, menuName, value, options = false) {
      this.setMenuVisibility(menuName, value);

      const targetRect = event.target.getBoundingClientRect();
      const targetTop = targetRect.y + targetRect.height + 15;
      const targetLeft = targetRect.x;
      const targetRight = targetRect.x + targetRect.width + 15;
      
      let x;
      let y;
      
      // Handle backward compatibility - if options is boolean (old anchor parameter)
      if (typeof options === 'boolean') {
        if (options) {
          this.anchorMenu = true;
          y = targetTop;
          x = targetLeft;
        } else {
          this.anchorMenu = false;
          y = event.clientY;
          x = event.clientX;
        }
      } else if (options && typeof options === 'object') {
        // New options object format
        this.anchorMenu = options.anchor || false;
        
        if (options.anchor) {
          switch (options.position) {
            case 'right':
              x = targetRight;
              y = targetRect.y; // Align to top of target
              break;
            case 'left':
              x = targetLeft - 15; // Add some spacing
              y = targetRect.y;
              break;
            case 'bottom':
              x = targetLeft;
              y = targetTop;
              break;
            case 'top':
              x = targetLeft;
              y = targetRect.y - 15; // Above the target
              break;
            default:
              // Default to bottom positioning (original behavior)
              x = targetLeft;
              y = targetTop;
          }
        } else {
          y = event.clientY;
          x = event.clientX;
        }
      } else {
        // No options provided, use default behavior
        this.anchorMenu = false;
        y = event.clientY;
        x = event.clientX;
      }

      const newPosition = { x, y };

      if (this.contextMenuVisible && !this.isAnimating) {
        this.contextMenuVisible = false;
        await nextTick();
        this.position = newPosition;
        this.contextMenuVisible = true;
      } else {
        this.position = newPosition;
        this.contextMenuVisible = true;
      }
    },

    async showCompactEditMenu(event, data) {
      this.compactEditMenuData = {
        key: data.key || '',
        title: data.title || '',
        loading: data.loading || false,
        options: data.options || [],
        selectedId: data.selectedId || '',
        onSelect: data.onSelect || null,
      };
      await this.showContextMenu(event, 'compactEditMenu', true, { anchor: true, position: 'bottom' });
    },

    updateCompactEditMenu(key, updates) {
      if (this.compactEditMenuData.key !== key) return;
      this.compactEditMenuData = { ...this.compactEditMenuData, ...updates };
    },

    hideContextMenu(event) {
      if (this.contextMenuVisible) {
        this.contextMenuVisible = false;
        this.disableAllMenus();
        this.resetSubMenu();
      }
    },

    setMenuVisibility(menuName, value) {
      if (this.menuStates.hasOwnProperty(menuName)) {
        // Check if the menu is already active
        if (value && this.activeMenu !== null && this.activeMenu !== menuName) {
          // Disable the currently active menu
          this.menuStates[this.activeMenu] = false;
        }

        this.menuStates[menuName] = value;
        this.activeMenu = value ? menuName : null;
      }
    },

    disableAllMenus() {
      for (const menuName in this.menuStates) {
        this.menuStates[menuName] = false;
      }
      this.compactEditMenuData = {
        key: '',
        title: '',
        loading: false,
        options: [],
        selectedId: '',
        onSelect: null,
      };
      this.activeMenu = null;
      this.contextMenuVisible = false;
    },

    isAnyMenuActive() {
      return Object.values(this.menuStates).some((isVisible) => isVisible);
    },

    triggerMenuItem(menu, page) {
      this.setModalVisibility(menu, false);
      this.showTraySearch = false;
      router.push({ name: page });
    },

    calculatePopUpPosition(event) {
      const targetTop = event.target.getBoundingClientRect().top;
      const targetHeight = event.target.getBoundingClientRect().height;
      const targetLeft = event.target.getBoundingClientRect().left;

      const container = this.contextMenuBounds.getBoundingClientRect();
      const contextMenu = this.contextMenu.getBoundingClientRect();

      const popUpMenuHeight = contextMenu.height;
      const top = container.top;
      const bottom = container.bottom;
      const spaceAbove = targetTop - top;
      const spaceBelow = bottom - targetTop;
      const halfHeight = popUpMenuHeight / 2;
      const offset = 10;

      if (spaceAbove < halfHeight) {
        this.popUpMenuTop = top + offset;
      } else if (spaceBelow < halfHeight) {
        this.popUpMenuTop = bottom - popUpMenuHeight - offset;
      } else {
        this.popUpMenuTop = targetTop - halfHeight + targetHeight / 2;
      }

      this.popUpMenuLeft = targetLeft - this.popUpMenuWidth - 10;
    },

    hidePopUpMenu() {
      this.assetPopUpMenuVisible = false;
    },

    // Force reposition of menu (useful when menu content changes height)
    forceRepositionMenu() {
      if (this.contextMenuVisible) {
        this.contextMenuVisible = false;
        nextTick(() => {
          this.contextMenuVisible = true;
        });
      }
    },

    // Sub-menu navigation actions
    showSubMenu(sourceMenuOrType, initialNavItem = null) {
      this.subMenuState.active = true;
      this.subMenuState.slideDirection = 'left';
      
      // Check if this is a move-to-collection request
      const isMoveToCollection = sourceMenuOrType === 'move-to-collection' || 
                                  initialNavItem?.type === 'move-to-collection';
      
      if (isMoveToCollection) {
        // For move-to-collection, navigation stack is set up in the component's onMounted
        // sourceMenuOrType might be 'assetMenu' (from context menu) or 'move-to-collection' (from DetailsPane)
        if (sourceMenuOrType !== 'move-to-collection') {
          this.subMenuState.sourceMenu = sourceMenuOrType;
          this.setMenuVisibility(sourceMenuOrType, false);
        } else {
          // Called from DetailsPane without a context menu - need to show the context menu container
          this.subMenuState.sourceMenu = null;
          this.contextMenuVisible = true;
        }
        this.subMenuState.navigationStack = []; // Will be initialized in onMounted
        this.setMenuVisibility('moveToCollectionSubMenu', true);
      } else {
        // For copyToProject and others, use the provided initialNavItem
        this.subMenuState.sourceMenu = sourceMenuOrType;
        if (initialNavItem) {
          this.subMenuState.navigationStack = [initialNavItem];
        }
        this.setMenuVisibility(sourceMenuOrType, false);
        this.setMenuVisibility('copyToProjectSubMenu', true);
      }
    },

    navigateSubMenuForward(navItem) {
      this.subMenuState.slideDirection = 'left';
      this.subMenuState.navigationStack.push(navItem);
    },

    navigateSubMenuBack() {
      if (this.subMenuState.navigationStack.length > 1) {
        this.subMenuState.slideDirection = 'right';
        this.subMenuState.navigationStack.pop();
      } else {
        // At root level, go back to source menu
        this.hideSubMenu();
      }
    },

    hideSubMenu() {
      const sourceMenu = this.subMenuState.sourceMenu;
      
      // Reset sub-menu state
      this.subMenuState.active = false;
      this.subMenuState.navigationStack = [];
      this.subMenuState.selectedProject = null;
      this.subMenuState.selectedAssetIds = [];
      this.subMenuState.startingCollectionId = '';
      this.subMenuState.slideDirection = 'left';
      
      // Hide sub-menus and show source menu
      this.setMenuVisibility('copyToProjectSubMenu', false);
      this.setMenuVisibility('moveToCollectionSubMenu', false);
      if (sourceMenu) {
        this.setMenuVisibility(sourceMenu, true);
        this.subMenuState.sourceMenu = null;
      }
    },

    resetSubMenu() {
      this.subMenuState.active = false;
      this.subMenuState.sourceMenu = null;
      this.subMenuState.navigationStack = [];
      this.subMenuState.selectedProject = null;
      this.subMenuState.selectedAssetIds = [];
      this.subMenuState.startingCollectionId = '';
      this.subMenuState.slideDirection = 'left';
    },
  },
});
