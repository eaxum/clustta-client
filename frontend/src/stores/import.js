import { defineStore } from "pinia";

export const useImportStore = defineStore("import", {
  state: () => ({
    defaultFileType: ["^.*$", "resource"],
    defaultEntityType: ["^.*$", "folder"],
    defaultTaskType: ["^.*$", "generic"],
    defaultResourceType: ["^.*$", "generic"],
    fileTypeRule: [
      ["^.*.blend$|^.*.BLEND$", "task"],
      ["^.*.spp$|^.*.SPP$", "task"],
      ["^.*.afdesign$|^.*.AFDESIGN$", "task"],
    ],
    entityTypeRule: [
      [".*/assets/$", "folder"],
      [".*/characters/$", "character"],
      [".*/props/$", "prop"],
      [".*/environment/$", "environment"],
      [".*/episodes/$", "folder"],
      ["^(?:.*/)?episodes/[^/]+/$", "episode"],
      [".*/episodes/[^/]+/[^/]+/?$", "sequence"],
      [".*/episodes/[^/]+/[^/]+/[^/]+/?$", "shot"],
      [".*/sequences/$", "folder"],
      ["^(?:.*/)?sequences/[^/]+/$", "sequence"],
      [".*/sequences/[^/]+/[^/]+/?$", "shot"],
    ],
    taskTypeRule: [
      [".*/characters/.*.blend$", "character creation"],
      [".*/props/.*.blend$", "prop creation"],
      [".*/environment/.*.blend$", "environment creation"],
      ["^.*animation.blend$", "animation"],
      ["^.*fx.blend$", "fx"],
      ["^.*lighting.blend$", "lighting"],
      ["^.*.spp$|^.*.SPP$", "texturing"],
      ["^.*.afdesign$|^.*.AFDESIGN$", "texturing"],
    ],
    resourceTypeRule: [
      ["^.*.pdf$|^.*.PDF$", "library"],
      ["^.*.png$|^.*.PNG$", "texture"],
      ["^.*.tif$|^.*.TIF$", "texture"],
      ["^.*.mp3$|^.*.MP3$", "video"],
      ["^.*.csv$|^.*.CSV$", "image"],
    ],
  }),
  getters: {},
  actions: {},
});
