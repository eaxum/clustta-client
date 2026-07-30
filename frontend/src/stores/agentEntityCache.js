import { defineStore } from "pinia";
import { AssetService, CollectionService } from "@/services";
import { useAssetStore } from "./assets";
import { useProjectStore } from "./projects";
import { useUserStore } from "./users";

// Caches asset/collection/user lookups for the inline chips that appear
// in agent chat messages. Dedupes in-flight fetches and remembers misses.
export const useAgentEntityCacheStore = defineStore("agentEntityCache", {
  state: () => ({
    assets: {},
    collections: {},
    users: {},
    untrackedAssets: {},
    untrackedCollections: {},
    missing: {},
    pending: {},
  }),
  actions: {
    // Returns the cached entity, kicking off a fetch if we haven't seen it yet.
    async ensure(type, id) {
      if (!type || !id) return null;
      const key = `${type}:${id}`;
      if (this.missing[key]) return null;

      const cached = this.get(type, id);
      if (cached) return cached;

      if (this.pending[key]) return this.pending[key];
      const promise = this.fetchEntity(type, id).finally(() => {
        delete this.pending[key];
      });
      this.pending[key] = promise;
      return promise;
    },

    // Synchronous read for templates.
    get(type, id) {
      const map = this.mapFor(type);
      return map && id ? map[id] || null : null;
    },

    // True once a lookup for this entity has come back empty.
    isMissing(type, id) {
      return !!this.missing[`${type}:${id}`];
    },

    // Picks the right state bucket for an entity type.
    mapFor(type) {
      if (type === "asset") return this.assets;
      if (type === "collection") return this.collections;
      if (type === "user") return this.users;
      if (type === "untracked_asset") return this.untrackedAssets;
      if (type === "untracked_collection") return this.untrackedCollections;
      return null;
    },

    // Fetches via the appropriate service and stores the result, or marks it missing.
    async fetchEntity(type, id) {
      const projectUri = useProjectStore().activeProject?.uri;
      const key = `${type}:${id}`;

      try {
        if (type === "untracked_asset" || type === "untracked_collection") {
          return this.markMissing(key);
        }
        if (type === "asset") {
          if (!projectUri) return null;
          const asset = await AssetService.GetAssetByID(projectUri, id);
          if (!asset?.id) return this.markMissing(key);
          await useAssetStore().processAssetsIconsAndPreviews([asset]);
          this.assets[id] = asset;
          return asset;
        }

        if (type === "collection") {
          if (!projectUri) return null;
          const collection = await CollectionService.GetCollectionByID(projectUri, id);
          if (!collection?.id) return this.markMissing(key);
          this.collections[id] = collection;
          return collection;
        }

        if (type === "user") {
          const userStore = useUserStore();
          const local = userStore.users?.find?.((u) => u.id === id);
          if (local) return this.storeUser(local);

          const remote = await userStore.fetchUserData(id).catch(() => null);
          if (remote?.id) return this.storeUser(remote);
          return this.markMissing(key);
        }
      } catch (error) {
        console.warn(`agentEntityCache: failed to resolve ${type}:${id}`, error);
        return this.markMissing(key);
      }
      return null;
    },

    // Normalises a user record (full_name) before caching it.
    storeUser(user) {
      this.users[user.id] = {
        ...user,
        full_name: user.full_name || [user.first_name, user.last_name].filter(Boolean).join(" "),
      };
      return this.users[user.id];
    },

    // Records a failed lookup so we don't keep retrying it.
    markMissing(key) {
      this.missing[key] = true;
      return null;
    },

    // Retains entity snapshots returned by scoped commands, including
    // untracked entities that cannot be fetched from tracked services.
    rememberCommandResult(data) {
      const visit = (value) => {
        if (!value) return;
        if (Array.isArray(value)) {
          value.forEach(visit);
          return;
        }
        if (typeof value !== "object") return;

        const entity = value.entity && typeof value.entity === "object" ? value.entity : value;
        if (entity.id && entity.type && entity.name) {
          const map = this.mapFor(entity.type);
          if (map) {
            map[entity.id] = { ...entity };
            delete this.missing[`${entity.type}:${entity.id}`];
          }
        }
        Object.values(value).forEach(visit);
      };
      visit(data);
    },
  },
});
