import * as esbuild from 'esbuild';
import { promises as fs, globSync } from "node:fs";
import { join, extname } from "node:path";

/**
 * @param {esbuild.BuildOptions} config
 */
const baseConfig = (config) => ({
  ...config,
  entryPoints: config.format === 'esm' ? [
    `src/index.ts`,
    `src/index.web.ts`,
    'src/generated/protos/index.*',
  ] : globSync(["src/**/*.ts"], { exclude: (filename) => filename.endsWith('.spec.ts') }),
  bundle: config.format === 'esm',
  sourcemap: true,
  packages: "external",
  platform: "neutral",
  external: config.format === 'esm' ? ["node:*"]: undefined,
  outExtension: config.format === 'cjs' ? { '.js': '.cjs' } : undefined,
  minify: false,
  target: [`es2020`],
  splitting: config.format === 'esm',
  outdir: `dist/${config.format}`,
  metafile: true,
  tsconfig: './tsconfig.build.json',
  plugins: config.format === 'cjs' ? [replaceTsToCjsPlugin()] : []
});


await Promise.all([
  esbuild.build(baseConfig({ format: 'esm' })),
  esbuild.build(baseConfig({
    format: 'cjs',
    supported: {
      'dynamic-import': false
    }
  })),
]);
await fs.copyFile('src/sdl/sdl-schema.yaml', 'dist/sdl-schema.yaml');
console.log('Building JS SDK finished');

// TODO: get rid of it when this https://github.com/evanw/esbuild/issues/2435#issuecomment-3303686541 will be done
function replaceTsToCjsPlugin(opts = {}) {
  const toExt = opts.toExt ?? ".cjs";

  const fromPattern = escapeReg('.ts');
  // only touch *relative* specifiers (./ or ../), avoid bare/deps/urls
  const reFrom   = new RegExp(`(\\bfrom\\s+["'])(\\.{1,2}\\/[^"']+)(?:${fromPattern})(["'])`, "g");
  const reImport = new RegExp(`(\\bimport\\(\\s*["'])(\\.{1,2}\\/[^"']+)(?:${fromPattern})(["']\\s*\\))`, "g");
  const reReq    = new RegExp(`(\\brequire\\(\\s*["'])(\\.{1,2}\\/[^"']+)(?:${fromPattern})(["']\\s*\\))`, "g");

  return {
    name: "replace-ts-to-cjs",
    setup(build) {
      build.onEnd(async () => {
        const outdir  = build.initialOptions.outdir;
        const outfile = build.initialOptions.outfile;
        const targets = [];

        if (outfile) targets.push(outfile);
        if (outdir)  targets.push(...await listFiles(outdir, [".cjs"]));

        await Promise.all(targets.map(async (f) => {
          let code = await fs.readFile(f, "utf8");
          const next = code
            .replace(reFrom,   `$1$2${toExt}$3`)
            .replace(reImport, `$1$2${toExt}$3`)
            .replace(reReq,    `$1$2${toExt}$3`);

          if (next !== code) await fs.writeFile(f, next);
        }));
      });
    },
  };
}

function escapeReg(s) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

async function listFiles(dir, allowExts) {
  const out = [];
  const entries = await fs.readdir(dir, { withFileTypes: true });
  for (const e of entries) {
    const p = join(dir, e.name);
    if (e.isDirectory()) {
      out.push(...await listFiles(p, allowExts));
    } else if (allowExts.includes(extname(p))) {
      out.push(p);
    }
  }
  return out;
}
