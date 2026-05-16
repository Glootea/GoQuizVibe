let initialized = false;

export async function initCompiler() {
  if (initialized) return true;

  $typst.setCompilerInitOptions({
    getModule: () => '/wasm/typst_ts_web_compiler_bg.wasm',
  });
  $typst.setRendererInitOptions({
    getModule: () => '/wasm/typst_ts_renderer_bg.wasm',
  });

  await $typst.getCompiler();
  await $typst.getRenderer();

  initialized = true;
  return true;
}

export async function compileTypst(source) {
  const svg = await $typst.svg({ mainContent: source });
  return { svg, diagnostics: [] };
}

export async function compileWithDiagnostics(source) {
  return compileTypst(source);
}

export function isInitialized() {
  return initialized;
}