let initialized = false;

export async function initCompiler() {
  initialized = true;
  return true;
}

export async function compileTypst(source) {
  if (!initialized) {
    await initCompiler();
  }

  return {
    svg: renderPreview(source),
    diagnostics: [],
  };
}

export async function compileWithDiagnostics(source) {
  return compileTypst(source);
}

function renderPreview(source) {
  const lines = source.split('\n');
  let html = '<div class="typst-preview">';

  for (const line of lines) {
    if (line.startsWith('= ')) {
      const level = line.match(/^=+/)[0].length;
      const text = line.replace(/^=+\s*/, '');
      const size = Math.max(1, 3 - level * 0.3);
      html += `<h${level} style="font-size: ${size}em; margin: 0.5em 0;">${escapeHtml(text)}</h${level}>`;
    } else if (line.startsWith('#')) {
      html += `<div class="typst-code">${escapeHtml(line)}</div>`;
    } else if (line.trim()) {
      html += `<p style="margin: 0.3em 0;">${escapeHtml(line)}</p>`;
    } else {
      html += '<br>';
    }
  }

  html += '</div>';
  return html;
}

function escapeHtml(text) {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

export function isInitialized() {
  return initialized;
}