import { defineStore } from "pinia";

export const useImportStore = defineStore("import", {
  state: () => ({
    defaultFileType: ["^.*$", "resource"],
    defaultCollectionType: ["^.*$", "folder"],
    defaultAssetType: ["^.*$", "generic"],
    defaultResourceType: ["^.*$", "generic"],
    fileTypeRule: [
      ["^.*.blend$|^.*.BLEND$", "asset"],
      ["^.*.spp$|^.*.SPP$", "asset"],
      ["^.*.afdesign$|^.*.AFDESIGN$", "asset"],
    ],
    collectionTypeRule: [
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
    assetTypeRule: [
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
