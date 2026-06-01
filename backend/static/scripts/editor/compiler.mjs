let currentMaterialId = null;

export async function initCompiler(materialId) {
    currentMaterialId = materialId;
    return true;
}

export async function compileTypst(source) {
    const resp = await fetch('/api/typst/compile/' + currentMaterialId, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            source: source
        }),
    });
    const data = await resp.json();
    return {
        url: data.url,
        diagnostics: []
    };
}

export async function compileWithDiagnostics(source) {
    return compileTypst(source);
}

export function isInitialized() {
    return true;
}