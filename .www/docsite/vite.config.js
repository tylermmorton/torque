import fs from "fs";
import { defineConfig } from "vite";
import { imports } from "./routes/import-map.json";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [],
  resolve: {
    alias: {
      // use the same alias as the import map
      ...imports,
    },
  },
  build: {
    manifest: "manifest.json",
    outDir: ".build/static",
    emptyOutDir: true,
    rollupOptions: {
      preserveEntrySignatures: "strict",
      input: {
        // dynamically load element files from ./elements:
        ...fs.readdirSync("./elements").reduce((acc, file) => {
          const name = file.split(".")[0];
          acc[name] = `./elements/${file}`;
          return acc;
        }, {}),
      },
      output: {
        format: "esm",
        exports: "named",
        entryFileNames: "[name].js",
      },
    },
  },
});
