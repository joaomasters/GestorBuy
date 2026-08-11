// Build simples com esbuild — sem framework, três entry points viram três
// arquivos .js na raiz da extensão (onde o manifest.json espera encontrar).
import * as esbuild from "esbuild";

const watch = process.argv.includes("--watch");

const commonOptions = {
  bundle: true,
  sourcemap: true,
  target: "chrome110",
  logLevel: "info",
};

const builds = [
  { entryPoints: ["src/background.ts"], outfile: "background.js", format: "iife" },
  { entryPoints: ["src/content-script.ts"], outfile: "content-script.js", format: "iife" },
  { entryPoints: ["src/popup.ts"], outfile: "popup.js", format: "iife" },
];

if (watch) {
  const contexts = await Promise.all(
    builds.map((b) => esbuild.context({ ...commonOptions, ...b }))
  );
  await Promise.all(contexts.map((c) => c.watch()));
  console.log("Observando mudanças... (Ctrl+C pra sair)");
} else {
  await Promise.all(builds.map((b) => esbuild.build({ ...commonOptions, ...b })));
  console.log("Build concluído.");
}
