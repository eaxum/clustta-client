const languageIcons = {
  cpp: '/code-icons/cpp.svg',
  csharp: '/code-icons/csharp.svg',
  css: '/code-icons/css.svg',
  dockerfile: '/code-icons/docker.svg',
  go: '/code-icons/go.svg',
  html: '/code-icons/html.svg',
  java: '/code-icons/java.svg',
  javascript: '/code-icons/javascript.svg',
  json: '/code-icons/json.svg',
  markdown: '/code-icons/markdown.svg',
  python: '/code-icons/python.svg',
  rust: '/code-icons/rust.svg',
  shell: '/code-icons/shell.svg',
  typescript: '/code-icons/typescript.svg',
  vue: '/code-icons/vue.svg',
  xml: '/code-icons/xml.svg',
  yaml: '/code-icons/yaml.svg',
};

const extensionIcons = {
  vue: '/code-icons/vue.svg',
};

export const getCodeFileIcon = (file) => {
  const extension = file.name?.split('.').pop()?.toLowerCase();
  return extensionIcons[extension]
    || languageIcons[file.language_id || file.classification?.language_id]
    || '/file-icons/default.svg';
};
