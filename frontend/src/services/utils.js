import { Buffer } from "buffer";
import { AppService, FSService } from "@/services";
import { nextTick } from "vue";
import { useCollectionStore } from "@/stores/collections";
import { md5 } from "./crypto.js";

// Fallback version used when the Go layer is unavailable (e.g. web mode).
// Keep in sync with build/config.yml and internal/constants fallbackVersion.
const FALLBACK_VERSION = "0.4.35";

const utils = {
  async getIcon(ext) {
    let base64 = "data:image/png;base64," + icon;
    return base64;
  },
  async getXxhashChecksum(path) {
    // let hash = await invoke("generate_xxhash_checksum", { path: path });
    return hash;
  },
  async getClusttaVersion() {
    const raw = await this.getRawClusttaVersion();
    return `v${raw}-beta`;
  },
   async getRawClusttaVersion() {
    try {
      const version = await AppService.GetVersion();
      if (version) return version;
    } catch (error) {
      // Go layer unavailable; fall back to the bundled version string.
    }
    return FALLBACK_VERSION;
  },
  base64ToUint8Array(base64) {
    const binary = atob(base64);
    const len = binary.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  },
  async base64ToFile(dataString) {
    const tempdirPath = await FSService.TempDir();
    const uniqueName = "clst-preview-" + crypto.randomUUID() + ".png";
    const tempFilePath = await FSService.JoinPath(
      tempdirPath,
      uniqueName
    );
    await FSService.WriteFile(tempFilePath, dataString)
      .then(() => {
        console.log("File created successfully");
      })
      .catch((err) => {
        console.log("Error creating file", err);
      });
    return tempFilePath;
  },
  async base64FromFile(filePath) {
    const base64String = await FSService.ReadFile(filePath);
    return "data:image/png;base64," + base64String;
  },
  async resizeBase64Img(base64, newWidth, newHeight) {
    return new Promise((resolve, reject) => {
      const canvas = document.createElement("canvas");
      canvas.width = newWidth;
      canvas.height = newHeight;
      let context = canvas.getContext("2d");
      let img = document.createElement("img");
      img.src = base64;
      img.onload = function () {
        //console.log("image size", img.width, img.height);
        context.scale(newWidth / img.width, newHeight / img.height);
        context.drawImage(img, 0, 0);
        resolve(canvas.toDataURL());
      };
    });
  },
  async revealInExplorer(filePath) {
    // invoke("reveal_in_explorer", { path: filePath });
  },
  async openUrl(url) {
    // invoke("open_url", { url: url });
  },
  async deleteFolder(folder) {
    // let message = await invoke("delete_folder", { folder: folder });
    //console.log(message);
  },
  async deleteFile(file) {
    // let message = await invoke("delete_file", { path: file });
    //console.log(message);
  },
  getFileExtension(filepath) {
    const match = filepath.match(/\.([^.]+)$/);
    return match ? match[0] : "";
  },
  getDomainName(link) {
    try {
      const url = new URL(link);

      const hostname = url.hostname;
      const parts = hostname.split('.');
      
      if (parts.length >= 2) {
        return parts.slice(-2).join('.');
      }
      
      return hostname;
    } catch (e) {
      return "";
    }
  },
  sortAlphabetically(data) {
    return data.sort((a, b) => a.name.localeCompare(b.name));
  },
  sortPathAlphabetically(data, type) {
    if (type === "asset") {
      return data.sort((a, b) => a.asset_path.localeCompare(b.asset_path));
    } else if (type === "collection") {
      return data.sort((a, b) => a.collection_path.localeCompare(b.collection_path));
    } else if (type === "resource") {
      return data.sort((a, b) =>
        a.resource_path.localeCompare(b.resource_path)
      );
    }
  },
  sortListAlphabetically(data) {
    return data.sort((a, b) => a.localeCompare(b));
  },
  scrollListItems(arr) {
    return arr.map((item, index) => ({
      name: item.name || item,
      icon: item.icon || "",
      meta: item.role || "",
      index: item.index || index.toString(),
    }));
  },
  startTransition(el) {
    el.style.height = el.scrollHeight + "px";
  },
  endTransition(el) {
    el.style.height = "";
  },
  formatDate(checkpointDate, locale) {
    const date = new Date(checkpointDate);
    const lng = locale || undefined;

    const formattedDate = date.toLocaleDateString(lng, {
      day: "numeric",
      month: "long",
      year: "numeric",
    });

    const formattedTime = date.toLocaleTimeString(lng, {
      hour: "numeric",
      minute: "2-digit",
    });

    return `${formattedDate} ${formattedTime}`;
  },
  capitalizeStr(str) {
    const formattedTxt = str?.replace(/_/g, " ");
    return formattedTxt?.charAt(0).toUpperCase() + formattedTxt?.slice(1);
  },
  handleHover(event) {
    let element = event.target;
    const elementChild = event.target.children[0];
    elementChild.style.overflow = "";
    elementChild.style.textOverflow = "";

    nextTick(() => {
      const isOverflowing = element.scrollWidth > element.offsetWidth;
      const scrollDist = element.scrollWidth - element.offsetWidth;
      if (isOverflowing) {
        //
        elementChild.style.transform = "translateX(" + -scrollDist + "px)";
        elementChild.style.transition = scrollDist / 12 + "s linear";
      }
    });
  },
  resetScroll(event) {
    let element = event.target;
    const elementChild = event.target.children[0];
    elementChild.style.transform = "translateX(0px)";
    elementChild.style.transition = 0 + "s linear";
    elementChild.style.overflow = "hidden";
    elementChild.style.textOverflow = "ellipsis";
  },
  /**
   * Finds the minimum common parent directory among the given paths
   * @param {string[]} paths - Array of file/directory paths
   * @returns {string} The common parent directory
   */
  findMinCommonParent(paths) {
    if (paths.length === 0) {
      return "";
    }
    if (paths.length === 1) {
      // For a single path, return its directory
      return paths[0].split("/").slice(0, -1).join("/");
    }

    // Normalize paths to use forward slashes
    const normalizedPaths = paths.map((path) => path.replace(/\\/g, "/"));

    // Split first path into components to use as reference
    const components = normalizedPaths[0].split("/");
    const commonPrefix = [];

    // Compare each component with all other paths
    for (let i = 0; i < components.length; i++) {
      const currentComponent = components[i];
      let isCommon = true;

      // Check if this component exists in all other paths at the same position
      for (let j = 1; j < normalizedPaths.length; j++) {
        const pathComponents = normalizedPaths[j].split("/");
        if (
          i >= pathComponents.length ||
          pathComponents[i] !== currentComponent
        ) {
          isCommon = false;
          break;
        }
      }

      if (isCommon) {
        commonPrefix.push(currentComponent);
      } else {
        break;
      }
    }

    return commonPrefix.join("/");
  },

  getParentPaths(path) {
    // Remove leading and trailing slashes
    path = path.replace(/^\/+|\/+$/g, "");

    // Initialize result array with the full path
    const paths = [path];

    // Split the path into segments
    const segments = path.split("/");

    // Generate parent paths
    while (segments.length > 1) {
      segments.pop();
      paths.push(segments.join("/"));
    }

    // Add empty string at the end
    // paths.push('');

    return paths;
  },
  getUntrackedCollectionparent(untracked) {
    const collectionStore = useCollectionStore();
    let parentPaths = this.getParentPaths(untracked.collection_path);
    for (let parent of parentPaths) {
      let collection = collectionStore.collections.find(
        (item) => item.collection_path === parent
      );
      if (collection !== undefined) {
        return collection;
      }
    }
    return null;
  },
  
  getMD5Hash(text) {
    return md5(text);
  },

  formatBytes(bytes, decimals = 1) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
  },

};

// Safely decodes HTML entities to plain text, stripping any HTML tags.
export const decodeEmoji = (html) => {
  const el = document.createElement('span');
  el.innerHTML = html;
  return el.textContent || '';
};

export default utils;



