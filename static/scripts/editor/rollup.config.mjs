import resolve from '@rollup/plugin-node-resolve';
import terser from '@rollup/plugin-terser';

export default {
  input: 'editor.mjs',
  output: {
    file: 'editor.bundle.js',
    format: 'iife',
    name: 'editor',
    sourcemap: false,
  },
  plugins: [
    resolve({
      browser: true,
      preferBuiltins: false,
    }),
    terser({
      compress: {
        drop_console: false,
      },
    }),
  ],
  inlineDynamicImports: true,
};