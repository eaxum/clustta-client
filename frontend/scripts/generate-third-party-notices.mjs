import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const frontendDirectory = resolve(scriptDirectory, '..');
const repositoryDirectory = resolve(frontendDirectory, '..');
const sourceNoticesPath = join(frontendDirectory, 'src', 'data', 'THIRD-PARTY-NOTICES.txt');
const noticesPath = join(frontendDirectory, 'public', 'licenses', 'THIRD-PARTY-NOTICES.txt');
const clusttaLicensePath = join(frontendDirectory, 'public', 'licenses', 'AGPL-3.0.txt');
const licenseFilePattern = /^(license|licence|copying|notice)(\.|$)/i;
const excludedPackages = new Set(['@clustta/web-adapters']);
const clusttaCopyright = 'Copyright (c) 2026 Eaxum';
const openFontLicenseSeparator = '-----------------------------------------------------------';

const readJson = (filePath) => JSON.parse(readFileSync(filePath, 'utf8'));

const findLicenseFiles = (directory) => {
  if (!directory || !existsSync(directory)) return [];
  return readdirSync(directory)
    .filter((name) => licenseFilePattern.test(name))
    .sort()
    .map((name) => join(directory, name));
};

const normalizeLicenseId = (value, licenseText) => {
  const declaredLicense = String(value || '');
  if (declaredLicense) return declaredLicense;
  if (licenseText.includes('GNU AFFERO GENERAL PUBLIC LICENSE')) return 'AGPL-3.0';
  if (licenseText.includes('Apache License')) return 'Apache-2.0';
  if (licenseText.includes('Mozilla Public License Version 2.0')) return 'MPL-2.0';
  if (licenseText.includes('CC0 1.0 Universal')) return 'CC0-1.0';
  if (
    licenseText.includes('ISC License') ||
    licenseText.includes('Permission to use, copy, modify, and/or distribute this software')
  ) return 'ISC';
  if (licenseText.includes('Permission is hereby granted, free of charge')) return 'MIT';
  if (licenseText.includes('Redistribution and use in source and binary forms')) return 'BSD';
  return 'See notice';
};

const readLicenseNotice = (directory) => {
  return findLicenseFiles(directory)
    .map((filePath) => `FILE: ${filePath.split(/[\\/]/).pop()}\n\n${readFileSync(filePath, 'utf8').trim()}`)
    .join('\n\n');
};

const resolveNodePackage = (packageName, baseDirectory) => {
  const localPath = join(baseDirectory, 'node_modules', packageName, 'package.json');
  if (existsSync(localPath)) return localPath;
  const rootPath = join(frontendDirectory, 'node_modules', packageName, 'package.json');
  return existsSync(rootPath) ? rootPath : '';
};

const collectNodePackages = () => {
  const applicationPackage = readJson(join(frontendDirectory, 'package.json'));
  const pending = Object.keys(applicationPackage.dependencies || {}).map((name) => ({
    name,
    baseDirectory: frontendDirectory,
  }));
  pending.push({ name: '@wailsio/runtime', baseDirectory: frontendDirectory });

  const packages = [];
  const visited = new Set();
  while (pending.length) {
    const current = pending.shift();
    if (excludedPackages.has(current.name)) continue;
    const packagePath = resolveNodePackage(current.name, current.baseDirectory);
    if (!packagePath) continue;

    const packageData = readJson(packagePath);
    const key = `${packageData.name}@${packageData.version}`;
    if (visited.has(key)) continue;
    visited.add(key);

    const packageDirectory = dirname(packagePath);
    const notice = readLicenseNotice(packageDirectory);
    packages.push({
      name: packageData.name,
      version: packageData.version,
      license: normalizeLicenseId(packageData.license, notice),
      notice,
    });

    for (const dependencyName of Object.keys(packageData.dependencies || {})) {
      pending.push({ name: dependencyName, baseDirectory: packageDirectory });
    }
  }
  return packages;
};

const collectGoPackages = () => {
  const distDirectory = join(frontendDirectory, 'dist');
  const distExisted = existsSync(distDirectory);
  const embedPlaceholder = join(distDirectory, `.license-generation-${process.pid}`);
  const template = '{{with .Module}}{{if .Dir}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}{{end}}';

  mkdirSync(distDirectory, { recursive: true });
  writeFileSync(embedPlaceholder, '');

  let output;
  try {
    output = execFileSync('go', ['list', '-deps', '-f', template, '.'], {
      cwd: repositoryDirectory,
      encoding: 'utf8',
    });
  } finally {
    rmSync(embedPlaceholder, { force: true });
    if (!distExisted) rmSync(distDirectory, { recursive: true, force: true });
  }

  return output
    .split(/\r?\n/)
    .filter(Boolean)
    .map((line) => {
      const [name, version, directory] = line.split('|');
      const notice = readLicenseNotice(directory);
      return {
        name,
        version: version || 'local',
        license: normalizeLicenseId('', notice),
        notice,
      };
    })
    .filter((entry) => entry.name !== 'clustta');
};

const collectGoStandardLibrary = () => {
  const goRoot = execFileSync('go', ['env', 'GOROOT'], { encoding: 'utf8' }).trim();
  const notice = readLicenseNotice(goRoot);
  return {
    name: 'Go standard library',
    version: execFileSync('go', ['env', 'GOVERSION'], { encoding: 'utf8' }).trim(),
    license: normalizeLicenseId('BSD-3-Clause', notice),
    notice,
  };
};

const collectFontPackages = () => {
  const interLicense = readFileSync(
    join(frontendDirectory, 'src', 'assets', 'fonts', 'Inter', 'OFL.txt'),
    'utf8',
  ).trim();
  const soraLicense = readFileSync(
    join(frontendDirectory, 'src', 'assets', 'fonts', 'Sora', 'OFL.txt'),
    'utf8',
  ).trim().replace(/[ \t]+$/gm, '');
  const sourGummyLicense = readFileSync(
    join(frontendDirectory, 'src', 'assets', 'fonts', 'SourGummy', 'OFL.txt'),
    'utf8',
  ).trim().replace(/[ \t]+$/gm, '');
  const licenseBodyStart = interLicense.indexOf(openFontLicenseSeparator);
  if (licenseBodyStart === -1) throw new Error('Could not find the SIL Open Font License body.');
  const openFontLicense = interLicense.slice(licenseBodyStart);
  const createOpenFontNotice = (copyright) => {
    return [
      copyright,
      '',
      'This Font Software is licensed under the SIL Open Font License, Version 1.1.',
      '',
      openFontLicense,
    ].join('\n');
  };
  return [
    {
      name: 'Inter',
      version: 'bundled font',
      license: 'OFL-1.1',
      notice: interLicense,
    },
    {
      name: 'Sora',
      version: 'bundled font',
      license: 'OFL-1.1',
      notice: soraLicense,
    },
    {
      name: 'Sour Gummy',
      version: 'bundled font',
      license: 'OFL-1.1',
      notice: sourGummyLicense,
    },
    {
      name: 'Bricolage Grotesque',
      version: '1.001',
      license: 'OFL-1.1',
      notice: createOpenFontNotice('Copyright 2022 The Bricolage Grotesque Project Authors'),
    },
    {
      name: 'Noto Color Emoji',
      version: '2.047',
      license: 'OFL-1.1',
      notice: createOpenFontNotice('Copyright 2022 Google Inc.'),
    },
    {
      name: 'Noto Emoji',
      version: '3.003',
      license: 'OFL-1.1',
      notice: createOpenFontNotice('Copyright 2013 Google LLC'),
    },
  ];
};

const deduplicatePackages = (packages) => {
  const uniquePackages = new Map();
  for (const packageEntry of packages) {
    uniquePackages.set(`${packageEntry.name}@${packageEntry.version}`, packageEntry);
  }
  return [...uniquePackages.values()].sort((left, right) => left.name.localeCompare(right.name));
};

const createNotices = (packages, clusttaLicense) => {
  const sections = packages.map((packageEntry) => {
    const heading = `${packageEntry.name} ${packageEntry.version} - ${packageEntry.license}`;
    return `${heading}\n${'='.repeat(heading.length)}\n\n${packageEntry.notice || 'No license text was found in the installed package.'}`;
  });
  return [
    `CLUSTTA LICENSE\n===============\n\n${clusttaCopyright}\n\n${clusttaLicense.trim()}`,
    'THIRD-PARTY NOTICES\n===================',
    ...sections,
  ].join('\n\n');
};

const clusttaLicense = readFileSync(join(repositoryDirectory, 'LICENSE'), 'utf8');
const packages = deduplicatePackages([
  ...collectGoPackages(),
  collectGoStandardLibrary(),
  ...collectNodePackages(),
  ...collectFontPackages(),
]);

mkdirSync(dirname(noticesPath), { recursive: true });
const notices = `${createNotices(packages, clusttaLicense)}\n`;
writeFileSync(sourceNoticesPath, notices);
writeFileSync(noticesPath, notices);
writeFileSync(clusttaLicensePath, clusttaLicense);

console.log(`Generated notices for ${packages.length} components.`);
