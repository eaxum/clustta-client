import { Buffer } from "buffer";
import { FSService } from "@/../bindings/clustta/services/index";
import { nextTick } from "vue";
import { useCollectionStore } from "@/stores/collections";
import { md5 } from "./crypto.js";

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
    return `v0.3.60-beta`;
  },
   async getRawClusttaVersion() {
    return `0.3.60`;
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
    const tempFilePath = await FSService.JoinPath(
      tempdirPath,
      "clst-preview.png"
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
      // Use the URL constructor to parse the link
      const url = new URL(link);

      // Return the hostname part of the URL
      return url.hostname;
    } catch (e) {
      // If the URL constructor throws an error, return an empty string or handle the error as needed
      return "";
    }
  },
  sortAlphabetically(data) {
    return data.sort((a, b) => a.name.localeCompare(b.name));
  },
  sortPathAlphabetically(data, type) {
    if (type === "task") {
      return data.sort((a, b) => a.task_path.localeCompare(b.task_path));
    } else if (type === "entity") {
      return data.sort((a, b) => a.entity_path.localeCompare(b.entity_path));
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
  formatDate(checkpointDate) {
    const date = new Date(checkpointDate);

    const formattedDate = date.toLocaleDateString("en-US", {
      day: "numeric",
      month: "long",
      year: "numeric",
    });

    const hours = date.getHours();
    const minutes = date.getMinutes().toString().padStart(2, "0");
    const ampm = hours >= 12 ? "PM" : "AM";
    const formattedTime = `${hours % 12}:${minutes} ${ampm}`;

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
  getUntrackedEntityparent(untracked) {
    const collectionStore = useCollectionStore();
    let parentPaths = this.getParentPaths(untracked.entity_path);
    for (let parent of parentPaths) {
      let entity = collectionStore.collections.find(
        (item) => item.entity_path === parent
      );
      if (entity !== undefined) {
        return entity;
      }
    }
    return null;
  },
  
  getMD5Hash(text) {
    return md5(text);
  },

};

export default utils;



