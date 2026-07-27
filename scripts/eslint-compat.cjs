// Preload hook: redirects `require('typescript')` to `@typescript/typescript6`
// for eslint + typescript-eslint compatibility with TypeScript 7.
//
// TypeScript 7 ("Project Corsa") removes the compiler API from the typescript
// package. typescript-eslint needs the full API (createProgram, sys, etc).
// Microsoft provides @typescript/typescript6 as a compat shim.
//
// Usage:
//   node --require ./scripts/eslint-compat.cjs ./node_modules/.bin/eslint .
'use strict';
const Module = require('module');
const origResolveFilename = Module._resolveFilename;
Module._resolveFilename = function (request, parent, isMain, options) {
  if (request === 'typescript') {
    return origResolveFilename.call(
      this,
      '@typescript/typescript6',
      parent,
      isMain,
      options,
    );
  }
  return origResolveFilename.call(this, request, parent, isMain, options);
};
