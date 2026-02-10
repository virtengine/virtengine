#!/usr/bin/env -S node --experimental-strip-types --no-warnings

import { execSync } from 'child_process';
import { accessSync, constants as fsConstants, readFileSync, globSync } from 'fs';
import { join as joinPath, dirname } from 'path';
import type PackageJson from '../package.json';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PACKAGE_ROOT = joinPath(__dirname, '..');
const packageJson = JSON.parse(readFileSync(joinPath(PACKAGE_ROOT, 'package.json'), 'utf8')) as typeof PackageJson;
const packageExports = Object.entries(packageJson.exports);

console.log(`Validating package exports for ${packageJson.name} in node ${process.version}...`);
const expandedExports = packageExports.reduce<typeof packageExports>((all, [subPath, config]) => {
  if (subPath.includes('*')) {
    const regex = new RegExp(escapeRegExp(config.import.slice(2)).replace('\\*', '([\\w.-]+)'));
    const files = globSync(joinPath(PACKAGE_ROOT, config.import))
      .map(path => {
        const matches = path.match(regex);
        if (!matches) throw new Error(`File ${path} does not match regex ${regex}`);
        return [subPath.replace('*', matches[1]), {
          types: config.types.replace('*', matches[1]),
          require: config.require.replace('*', matches[1]),
          import: config.import.replace('*', matches[1]),
        }] as typeof packageExports[number];
      });

    return all.concat(files);
  }
  all.push([subPath, config]);
  return all;
}, []);
for (const [subPath, config] of expandedExports) {
  if (subPath.includes('*')) continue;

  console.log(`Validating export ${subPath === '.' ? 'root' : subPath}...`);
  // Test commonjs require in commonjs runtime
  const exportPathCommonjs = joinPath(PACKAGE_ROOT, config.require);
  accessSync(exportPathCommonjs, fsConstants.R_OK);
  testExport('commonjs', subPath);

  // Test ESM
  const exportPathEsm = joinPath(PACKAGE_ROOT, config.import);
  accessSync(exportPathEsm, fsConstants.R_OK);
  testExport('module', subPath);
  testExport('dynamic-module', subPath);
}

function testExport(importType: 'commonjs' | 'module' | 'dynamic-module', subPath: string) {
  const moduleImport = packageJson.name + subPath.slice(1);
  let runtimeType = importType;
  let jsCode: string;
  switch (importType) {
    case 'commonjs':
      jsCode = `console.log(Object.keys(require('${moduleImport}')).length > 0);`;
      break;
    case 'module':
      jsCode = `import * as sdk from '${moduleImport}';`
      jsCode += `console.log(Object.keys(sdk).length > 0);`;
      break;
    case 'dynamic-module':
      jsCode = `import('${moduleImport}').then((sdk) => {`
      jsCode += `console.log(Object.keys(sdk).length > 0);`;
      jsCode += `});`;
      runtimeType = 'module';
      break;
  }

  const command = `node --input-type=${runtimeType} --eval="${jsCode}";`;
  const result = execSync(command, { cwd: PACKAGE_ROOT }).toString().trim();

  if (result !== 'true') {
    throw new Error(`Export ${subPath} is not valid for ${importType} runtime`);
  }
}

function escapeRegExp(s: string) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
