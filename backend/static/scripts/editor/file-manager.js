const FILES_DIR = 'files';
const DEFAULT_FILE = 'untitled.typ';
const STORAGE_KEY = 'typst-editor-files';

let opfsRoot = null;
let filesDirHandle = null;
let useOPFS = false;

export async function initOPFS() {
  try {
    if ('storage' in navigator && 'getDirectory' in navigator.storage) {
      opfsRoot = await navigator.storage.getDirectory();
      filesDirHandle = await opfsRoot.getDirectoryHandle(FILES_DIR, { create: true });

      const fileHandle = await filesDirHandle.getFileHandle(DEFAULT_FILE, { create: true });
      const writable = await fileHandle.createWritable();
      await writable.write('= Welcome to Typst Editor\n\nStart typing your document here.\n');
      await writable.close();

      useOPFS = true;
      return true;
    }
  } catch (error) {
    console.warn('OPFS not available, using localStorage fallback:', error);
  }

  initLocalStorage();
  return true;
}

function initLocalStorage() {
  if (!localStorage.getItem(STORAGE_KEY)) {
    const initialFiles = {};
    initialFiles[DEFAULT_FILE] = '= Welcome to Typst Editor\n\nStart typing your document here.\n';
    localStorage.setItem(STORAGE_KEY, JSON.stringify(initialFiles));
  }
}

export async function listFiles() {
  if (useOPFS) {
    const files = [];
    try {
      for await (let [name, handle] of filesDirHandle.entries()) {
        if (handle.kind === 'file') {
          files.push(name);
        }
      }
      return files.sort();
    } catch (error) {
      console.error('Failed to list files:', error);
      return [];
    }
  } else {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return Object.keys(JSON.parse(stored)).sort();
    }
    return [];
  }
}

export async function createFile(name) {
  if (!name.endsWith('.typ')) {
    name += '.typ';
  }

  if (useOPFS) {
    try {
      const fileHandle = await filesDirHandle.getFileHandle(name, { create: true });
      const writable = await fileHandle.createWritable();
      await writable.write('');
      await writable.close();
      return name;
    } catch (error) {
      console.error('Failed to create file:', error);
      return null;
    }
  } else {
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    stored[name] = '';
    localStorage.setItem(STORAGE_KEY, JSON.stringify(stored));
    return name;
  }
}

export async function deleteFile(name) {
  if (useOPFS) {
    try {
      await filesDirHandle.removeEntry(name);
      return true;
    } catch (error) {
      console.error('Failed to delete file:', error);
      return false;
    }
  } else {
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    delete stored[name];
    localStorage.setItem(STORAGE_KEY, JSON.stringify(stored));
    return true;
  }
}

export async function openFile(name) {
  if (useOPFS) {
    try {
      const fileHandle = await filesDirHandle.getFileHandle(name);
      const file = await fileHandle.getFile();
      const content = await file.text();
      return {
        name,
        content,
        lastModified: file.lastModified,
      };
    } catch (error) {
      console.error('Failed to open file:', error);
      return null;
    }
  } else {
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    if (stored[name] !== undefined) {
      return {
        name,
        content: stored[name],
        lastModified: Date.now(),
      };
    }
    return null;
  }
}

export async function saveFile(name, content) {
  if (useOPFS) {
    try {
      const fileHandle = await filesDirHandle.getFileHandle(name);
      const writable = await fileHandle.createWritable();
      await writable.write(content);
      await writable.close();
      return true;
    } catch (error) {
      console.error('Failed to save file:', error);
      return false;
    }
  } else {
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    stored[name] = content;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(stored));
    return true;
  }
}

export { DEFAULT_FILE };