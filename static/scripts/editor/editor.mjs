import { EditorView, basicSetup } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { keymap, lineNumbers } from '@codemirror/view';
import { defaultKeymap, historyKeymap } from '@codemirror/commands';
import { syntaxHighlighting, defaultHighlightStyle, StreamLanguage } from '@codemirror/language';

// Custom Typst language mode
const typstLanguage = StreamLanguage.define({
  name: 'typst',
  startState() {
    return { inCode: false, inMath: false };
  },
  token(stream, state) {
    if (stream.sol()) {
      state.inCode = false;
      state.inMath = false;
    }

    if (stream.match(/^\/\/.*/)) return 'comment';
    if (stream.match(/^\/\*[\s\S]*?\*\//)) return 'comment';

    if (stream.match(/^\$/)) {
      state.inMath = !state.inMath;
      return 'meta';
    }

    if (state.inMath) {
      if (stream.match(/^\\[a-zA-Z_][a-zA-Z0-9_]*/)) return 'variableName.special';
      if (stream.match(/^[0-9]+(\.[0-9]+)?/)) return 'number';
      if (stream.match(/^[_a-zA-Z][a-zA-Z0-9_]*/)) return 'variableName';
      if (stream.match(/^[\+\-\*\/\^\_\=\.\,\;\:\(\)\[\]\{\}]/)) return 'operator';
      stream.next();
      return null;
    }

    if (stream.match(/^#\{/)) { state.inCode = true; return 'punctuation'; }
    if (stream.match(/^\}/) && state.inCode) { state.inCode = false; return 'punctuation'; }

    if (state.inCode) {
      if (stream.match(/^(let|if|else|for|in|while|return|import|include|set|show|break|continue|and|or|not)\b/)) return 'keyword';
      if (stream.match(/^(none|auto)\b/)) return 'constant';
      if (stream.match(/^(true|false)\b/)) return 'constant.boolean';
      if (stream.match(/^[0-9]+(\.[0-9]+)?/)) return 'number';
      if (stream.match(/^"[^*]*"/)) return 'string';
      if (stream.match(/^'[^']*'/)) return 'string';
      if (stream.match(/^[a-zA-Z_][a-zA-Z0-9_]*/)) return stream.peek() === '(' ? 'function' : 'variableName';
      if (stream.match(/^[\+\-\*\/\=\<\>\!\[\]\{\}\,\;\(\)]/)) return 'operator';
      if (stream.match(/^\.\./)) return 'operator';
      stream.next();
      return null;
    }

    if (stream.match(/^```/)) { state.inRaw = true; return 'string'; }
    if (stream.match(/^#\w+/)) return 'tag';
    if (stream.match(/^=+ /)) return 'heading';
    if (stream.match(/^\*\*[^*]+\*\*/)) return 'strong';
    if (stream.match(/^\*[^*]+\*/)) return 'emphasis';
    if (stream.match(/^`[^`]+`/)) return 'monospace';
    if (stream.match(/^- /)) return 'list';
    if (stream.match(/^http[s]?:\/\/\S+/)) return 'link';

    stream.next();
    return null;
  },
});

import * as fileManager from './file-manager.js';
import * as compiler from './compiler.mjs';

let editorView = null;
let currentFile = null;
let saveTimeout = null;
let compileTimeout = null;

window.addEventListener('DOMContentLoaded', async () => {
  console.log('DOMContentLoaded');

  const ok = await fileManager.initOPFS();
  if (!ok) {
    showError('Failed to initialize file system');
    return;
  }

  await compiler.initCompiler();
  initEditor();
  await loadFileList();
  setupResizeHandlers();
  setupFileActions();
});

function initEditor() {
  console.log('initEditor called');
  const container = document.getElementById('editor-container');
  if (!container) {
    console.error('Editor container not found');
    return;
  }

  console.log('Creating CodeMirror 6 editor');

  const startState = EditorState.create({
    doc: '',
    extensions: [
      basicSetup,
      lineNumbers(),
      keymap.of([...defaultKeymap, ...historyKeymap]),
      syntaxHighlighting(defaultHighlightStyle),
      typstLanguage,
      EditorView.lineWrapping,
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const content = update.state.doc.toString();
          scheduleSave(content);
          scheduleCompile(content);
        }
      }),
    ],
  });

  editorView = new EditorView({
    state: startState,
    parent: container,
  });

  console.log('CodeMirror 6 created:', editorView);
}

async function loadFileList() {
  const files = await fileManager.listFiles();
  renderFileList(files);

  if (files.length > 0) {
    const defaultFile = files.includes(fileManager.DEFAULT_FILE)
      ? fileManager.DEFAULT_FILE
      : files[0];
    await openFile(defaultFile);
  }
}

function renderFileList(files) {
  const fileList = document.getElementById('file-list');
  fileList.innerHTML = '';

  files.forEach((name) => {
    const item = document.createElement('div');
    item.className = 'file-item' + (name === currentFile ? ' active' : '');
    item.innerHTML = `<i class="fas fa-file-code"></i><span>${name}</span>`;
    item.addEventListener('click', () => openFile(name));
    fileList.appendChild(item);
  });
}

async function openFile(name) {
  const file = await fileManager.openFile(name);
  if (!file) return;

  currentFile = name;

  if (editorView) {
    editorView.dispatch({
      changes: {
        from: 0,
        to: editorView.state.doc.length,
        insert: file.content,
      },
    });
  }

  renderFileList(await fileManager.listFiles());

  document.getElementById('delete-file').disabled = false;

  scheduleCompile(file.content);
}

function scheduleSave(content) {
  if (saveTimeout) clearTimeout(saveTimeout);
  saveTimeout = setTimeout(() => {
    if (currentFile) {
      fileManager.saveFile(currentFile, content);
    }
  }, 500);
}

function scheduleCompile(content) {
  if (compileTimeout) clearTimeout(compileTimeout);
  compileTimeout = setTimeout(async () => {
    await compileAndPreview(content);
  }, 300);
}

async function compileAndPreview(content) {
  const result = await compiler.compileWithDiagnostics(content);

  const preview = document.getElementById('preview-container');
  preview.innerHTML = result.svg || '<div class="preview-placeholder">No preview available</div>';

  renderErrors(result.diagnostics);
}

function renderErrors(diagnostics) {
  const errorsList = document.getElementById('errors-list');
  errorsList.innerHTML = '';

  if (!diagnostics || diagnostics.length === 0) {
    const item = document.createElement('div');
    item.className = 'error-item info';
    item.innerHTML = '<i class="fas fa-check-circle"></i><span>No errors</span>';
    errorsList.appendChild(item);
    return;
  }

  diagnostics.forEach((diag) => {
    const item = document.createElement('div');
    item.className = `error-item ${diag.severity}`;

    const icon = diag.severity === 'error' ? 'fa-exclamation-circle' :
                 diag.severity === 'warning' ? 'fa-exclamation-triangle' : 'fa-info-circle';

    const location = diag.range
      ? `Line ${diag.range.startLine + 1}:${diag.range.startColumn}`
      : '';

    item.innerHTML = `
      <i class="fas ${icon}"></i>
      <span class="error-location">${location}</span>
      <span class="error-message">${escapeHtml(diag.message)}</span>
    `;
    errorsList.appendChild(item);
  });
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function setupResizeHandlers() {
  const handlers = document.querySelectorAll('.resize-handler-v, .resize-handler-h');

  handlers.forEach((handler) => {
    let isResizing = false;
    let startPos = 0;
    let initialWidth = 0;
    let initialMiddleRowHeight = 0;
    let initialErrorsHeight = 0;
    let initialEditorWidth = 0;

    handler.addEventListener('mousedown', (e) => {
      isResizing = true;
      startPos = handler.classList.contains('resize-handler-v') ? e.clientX : e.clientY;

      const resizeType = handler.dataset.resize;

      if (resizeType === 'files') {
        const filesPanel = document.getElementById('files-panel');
        initialWidth = filesPanel.offsetWidth;
      } else if (resizeType === 'editor-preview') {
        const editorPanel = document.querySelector('.editor-panel');
        initialEditorWidth = editorPanel.offsetWidth;
      } else if (resizeType === 'errors') {
        const middleRow = document.querySelector('.middle-row');
        const errorsPanel = document.querySelector('.errors-panel');
        initialMiddleRowHeight = middleRow.offsetHeight;
        initialErrorsHeight = errorsPanel.offsetHeight;
      }

      document.body.style.cursor = handler.classList.contains('resize-handler-v') ? 'col-resize' : 'row-resize';
      document.body.style.userSelect = 'none';
    });

    document.addEventListener('mousemove', (e) => {
      if (!isResizing) return;

      const resizeType = handler.dataset.resize;

      if (resizeType === 'files') {
        const filesPanel = document.getElementById('files-panel');
        const delta = e.clientX - startPos;
        const newWidth = Math.max(40, Math.min(400, initialWidth + delta));
        filesPanel.style.width = newWidth + 'px';
      } else if (resizeType === 'editor-preview') {
        const editorPanel = document.querySelector('.editor-panel');
        const previewPanel = document.querySelector('.preview-panel');
        const delta = e.clientX - startPos;
        const newEditorWidth = Math.max(100, initialEditorWidth + delta);
        const newPreviewWidth = Math.max(100, previewPanel.offsetWidth - delta);
        if (newEditorWidth >= 100 && newPreviewWidth >= 100) {
          editorPanel.style.flex = '0 0 ' + newEditorWidth + 'px';
        }
      } else if (resizeType === 'errors') {
        const middleRow = document.querySelector('.middle-row');
        const errorsPanel = document.querySelector('.errors-panel');
        const delta = e.clientY - startPos;
        const newMiddleRowHeight = Math.max(100, initialMiddleRowHeight + delta);
        const newErrorsHeight = Math.max(60, initialErrorsHeight - delta);
        if (newMiddleRowHeight >= 100 && newErrorsHeight >= 60) {
          middleRow.style.flex = '0 0 ' + newMiddleRowHeight + 'px';
          errorsPanel.style.height = newErrorsHeight + 'px';
        }
      }
    });

    document.addEventListener('mouseup', () => {
      if (isResizing) {
        isResizing = false;
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
      }
    });
  });
}

function setupFileActions() {
  const newFileBtn = document.getElementById('new-file');
  const deleteFileBtn = document.getElementById('delete-file');

  newFileBtn.addEventListener('click', async () => {
    const name = prompt('Enter file name:', 'new-document');
    if (!name) return;

    const created = await fileManager.createFile(name);
    if (created) {
      await loadFileList();
      await openFile(created);
    }
  });

  deleteFileBtn.addEventListener('click', async () => {
    if (!currentFile) return;
    if (!confirm(`Delete ${currentFile}?`)) return;

    const deleted = await fileManager.deleteFile(currentFile);
    if (deleted) {
      currentFile = null;
      await loadFileList();
      deleteFileBtn.disabled = true;
    }
  });
}

function showError(message) {
  const errorsList = document.getElementById('errors-list');
  errorsList.innerHTML = `
    <div class="error-item error">
      <i class="fas fa-exclamation-circle"></i>
      <span class="error-message">${escapeHtml(message)}</span>
    </div>
  `;
}

export { editorView };