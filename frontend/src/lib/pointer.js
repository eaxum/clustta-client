import { FSService } from "@/../bindings/clustta/services";

export function isValidWeblink(link) {
  try {
    // Use the URL constructor to parse the link
    const url = new URL(link);

    // Check if the protocol is either http or https
    if (url.protocol === "http:" || url.protocol === "https:") {
      return true;
    }
  } catch (e) {
    // If the URL constructor throws an error, the link is not valid
  }

  return false;
}
export async function isValidPointer(pointer) {
  if (!isValidWeblink(pointer) && !(await FSService.Exists(pointer))) {
    return false;
  }
  return true;
}
