#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
const versionFile = path.join(repoRoot, 'frontend', 'src', 'version.ts');

const major = (process.env.APP_VERSION_MAJOR || '1').trim();
const minor = (process.env.APP_VERSION_MINOR || '0').trim();

function pad2(n) {
  return String(n).padStart(2, '0');
}

function buildTimestamp() {
  const now = new Date();
  const yy = String(now.getFullYear()).slice(-2);
  const mm = pad2(now.getMonth() + 1);
  const dd = pad2(now.getDate());
  const hh = pad2(now.getHours());
  const mi = pad2(now.getMinutes());
  return `${yy}${mm}${dd}${hh}${mi}`;
}

const version = `V${major}.${minor}.${buildTimestamp()}`;
const next = `export const APP_VERSION = '${version}';\n`;

fs.writeFileSync(versionFile, next, 'utf8');
console.log(version);
