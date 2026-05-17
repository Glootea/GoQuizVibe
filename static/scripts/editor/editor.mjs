import { EditorView, basicSetup } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { keymap, lineNumbers } from '@codemirror/view';
import { defaultKeymap, historyKeymap } from '@codemirror/commands';
import { syntaxHighlighting, defaultHighlightStyle, StreamLanguage } from '@codemirror/language';

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

import * as compiler from './compiler.mjs';

let editorView = null;
let compileTimeout = null;

window.addEventListener('DOMContentLoaded', async () => {
    const materialId = window.MATERIAL_ID || null;
    const initialSource = window.INITIAL_SOURCE || '';
    const pdfURL = window.PDF_URL || '';

    const previewFrame = document.getElementById('preview-frame');
    const placeholder = document.getElementById('preview-placeholder');

    if (pdfURL) {
        previewFrame.src = pdfURL;
        previewFrame.style.display = 'block';
        placeholder.style.display = 'none';
    } else {
        previewFrame.style.display = 'none';
        placeholder.style.display = 'flex';
    }

    if (materialId) {
        await compiler.initCompiler(materialId);
    }

    initEditor(initialSource);
    setupResizeHandlers();
});

function initEditor(initialSource) {
    const container = document.getElementById('editor-container');
    if (!container) {
        console.error('Editor container not found');
        return;
    }

    const startState = EditorState.create({
        doc: initialSource || '',
        extensions: [
            basicSetup,
            lineNumbers(),
            keymap.of([...defaultKeymap, ...historyKeymap]),
            syntaxHighlighting(defaultHighlightStyle),
            typstLanguage,
            EditorView.lineWrapping,
            EditorView.updateListener.of((update) => {
                if (update.docChanged) {
                    scheduleCompile(update.state.doc.toString());
                }
            }),
        ],
    });

    editorView = new EditorView({
        state: startState,
        parent: container,
    });
}

function scheduleCompile(content) {
    if (compileTimeout) clearTimeout(compileTimeout);
    compileTimeout = setTimeout(async () => {
        await compileAndPreview(content);
    }, 300);
}

async function compileAndPreview(content) {
    try {
        const result = await compiler.compileWithDiagnostics(content);

        const previewFrame = document.getElementById('preview-frame');
        const placeholder = document.getElementById('preview-placeholder');

        if (result.url) {
            previewFrame.src = result.url;
            previewFrame.style.display = 'block';
            placeholder.style.display = 'none';
        }

        renderErrors(result.diagnostics);
    } catch (err) {
        showError('Compilation failed: ' + err.message);
    }
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
        let initialEditorWidth = 0;
        let initialPreviewWidth = 0;

        handler.addEventListener('mousedown', (e) => {
            isResizing = true;
            startPos = handler.classList.contains('resize-handler-v') ? e.clientX : e.clientY;

            const resizeType = handler.dataset.resize;

            if (resizeType === 'files') {
                const filesPanel = document.getElementById('files-panel');
                initialWidth = filesPanel.offsetWidth;
            } else if (resizeType === 'editor-preview') {
                const editorPanel = document.querySelector('.editor-panel');
                const previewPanel = document.querySelector('.preview-panel');
                initialEditorWidth = editorPanel.offsetWidth;
                initialPreviewWidth = previewPanel.offsetWidth;
            } else if (resizeType === 'errors') {
                const middleRow = document.querySelector('.middle-row');
                const errorsPanel = document.querySelector('.errors-panel');
                initialMiddleRowHeight = middleRow.offsetHeight;
                initialEditorWidth = errorsPanel.offsetHeight;
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
                const delta = e.clientX - startPos;
                const newEditorWidth = Math.max(100, initialEditorWidth + delta);
                const newPreviewWidth = Math.max(100, initialPreviewWidth - delta);
                if (newEditorWidth >= 100 && newPreviewWidth >= 100) {
                    editorPanel.style.flex = '0 0 ' + newEditorWidth + 'px';
                }
            } else if (resizeType === 'errors') {
                const middleRow = document.querySelector('.middle-row');
                const errorsPanel = document.querySelector('.errors-panel');
                const delta = e.clientY - startPos;
                const newMiddleRowHeight = Math.max(100, initialMiddleRowHeight + delta);
                const newErrorsHeight = Math.max(60, initialEditorWidth - delta);
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