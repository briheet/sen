// Builds a static call graph for a Node.js project using the TypeScript API.
// Usage: node analyze.cjs <projectDir>
// Emits JSON: {root, files: [{path, functions: [{id, name, startLine, startCol, endLine, endCol}]}], edges: [{from, to}]}
const fs = require('fs');
const path = require('path');
const { createRequire } = require('module');

const projectDir = process.argv[2];

let ts;
try {
  ts = createRequire(path.join(projectDir, 'noop.js'))('typescript');
} catch (err) {
  console.error('senbon: typescript not found. Install it with `npm i -D typescript`.');
  process.exit(1);
}
if (typeof ts.createSourceFile !== 'function') {
  console.error('senbon: incompatible typescript version. Install a compiler-API build with `npm i -D typescript@^5.9`.');
  process.exit(1);
}

const SOURCE_EXTS = new Set(['.ts', '.tsx', '.js', '.jsx', '.mjs', '.cjs', '.mts', '.cts']);
const SKIP_DIRS = new Set(['node_modules', '.git', 'dist', 'build']);

function collectFiles(dir, out) {
  let entries;
  try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch (err) { return; }
  for (const entry of entries) {
    if (entry.isDirectory()) {
      if (!SKIP_DIRS.has(entry.name)) collectFiles(path.join(dir, entry.name), out);
    } else if (entry.isFile() && SOURCE_EXTS.has(path.extname(entry.name))) {
      out.push(path.join(dir, entry.name));
    }
  }
}

function resolveEntry(dir) {
  const pkgPath = path.join(dir, 'package.json');
  if (fs.existsSync(pkgPath)) {
    const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
    if (typeof pkg.main === 'string' && fs.existsSync(path.join(dir, pkg.main))) {
      return path.join(dir, pkg.main);
    }
    const bins = typeof pkg.bin === 'string' ? [pkg.bin] : Object.values(pkg.bin || {});
    for (const bin of bins) {
      if (typeof bin === 'string' && fs.existsSync(path.join(dir, bin))) return path.join(dir, bin);
    }
  }
  for (const name of ['index.js', 'index.mjs', 'index.cjs', 'index.ts']) {
    if (fs.existsSync(path.join(dir, name))) return path.join(dir, name);
  }
  console.error('senbon: no entry file found (package.json main/bin or index.*)');
  process.exit(1);
}

const entry = resolveEntry(projectDir);
const files = [];
collectFiles(projectDir, files);
if (!files.includes(entry)) files.unshift(entry);

const rootId = 0;
const declarations = new Map();
const edges = [];
const seenEdges = new Set();
const filesOut = [];

function lineAndChar(sourceFile, pos) {
  const loc = sourceFile.getLineAndCharacterOfPosition(pos);
  return { line: loc.line, character: loc.character };
}

function register(name, sourceFile, startPos, endPos, fileIndex) {
  const start = lineAndChar(sourceFile, startPos);
  const end = lineAndChar(sourceFile, endPos);
  const decl = { id: declarations.size, name, fileIndex,
    startLine: start.line, startCol: start.character,
    endLine: end.line, endCol: end.character };
  declarations.set(decl.id, decl);
  filesOut[fileIndex].functions.push(decl);
  return decl.id;
}

function calleeName(expression) {
  if (ts.isIdentifier(expression)) return expression.text;
  if (ts.isPropertyAccessExpression(expression)) return expression.name.text;
  if (ts.isElementAccessExpression(expression) && ts.isStringLiteral(expression.argumentExpression)) {
    return expression.argumentExpression.text;
  }
  return null;
}

function resolveCallee(name, fileIndex) {
  for (const decl of declarations.values()) {
    if (decl.name === name && decl.fileIndex === fileIndex) return decl.id;
  }
  for (const decl of declarations.values()) {
    if (decl.name === name) return decl.id;
  }
  return null;
}

function addEdge(from, to) {
  const key = from + ':' + to;
  if (!seenEdges.has(key)) { seenEdges.add(key); edges.push({ from, to }); }
}

files.forEach((filePath, fileIndex) => {
  filesOut[fileIndex] = { path: filePath, functions: [] };
  const text = fs.readFileSync(filePath, 'utf8');
  const sourceFile = ts.createSourceFile(filePath, text, ts.ScriptTarget.Latest, true);
  const scope = [];

  function walk(node) {
    if (ts.isFunctionLike(node)) {
      let name = '<anonymous>';
      if (node.name) name = node.name.text;
      else if (ts.isVariableDeclaration(node.parent) && ts.isIdentifier(node.parent.name)) name = node.parent.name.text;
      const id = register(name, sourceFile, node.getStart(sourceFile), node.getEnd(), fileIndex);
      scope.push(id);
      ts.forEachChild(node, walk);
      scope.pop();
      return;
    }
    if (ts.isCallExpression(node) || ts.isNewExpression(node)) {
      const name = calleeName(node.expression);
      if (name) {
        const callee = resolveCallee(name, fileIndex);
        if (callee !== null) addEdge(scope.length ? scope[scope.length - 1] : rootId, callee);
      }
    }
    ts.forEachChild(node, walk);
  }

  if (filePath === entry) {
    const start = lineAndChar(sourceFile, 0);
    const end = lineAndChar(sourceFile, text.length);
    const root = { id: rootId, name: path.basename(filePath), fileIndex,
      startLine: start.line, startCol: start.character,
      endLine: end.line, endCol: end.character };
    declarations.set(rootId, root);
    filesOut[fileIndex].functions.push(root);
    scope.push(rootId);
  }

  walk(sourceFile);
});

console.log(JSON.stringify({ root: rootId, files: filesOut, edges }));
