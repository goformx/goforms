import fs from "fs";
import path from "path";
import type { Plugin } from "vite";

/**
 * File system utility for checking template existence
 */
class FileSystem {
  static exists(filePath: string): boolean {
    return fs.existsSync(filePath);
  }

  static readFile(filePath: string): string {
    return fs.readFileSync(filePath, "utf-8");
  }
}

/**
 * Template resolution utility
 */
class TemplateResolver {
  static resolve(id: string, importer?: string): string | null {
    const resolvedPath = path.isAbsolute(id)
      ? id
      : importer
        ? path.resolve(path.dirname(importer), id)
        : path.resolve(process.cwd(), id);

    return this.findTemplate(resolvedPath);
  }

  private static findTemplate(id: string): string | null {
    const ejsJsPath = id.replace(".ejs", ".ejs.js");

    if (FileSystem.exists(ejsJsPath)) {
      return ejsJsPath;
    }
    return FileSystem.exists(id) ? id : null;
  }
}

/**
 * Vite plugin for handling EJS templates
 */
export default function ejsPlugin(): Plugin {
  return {
    name: "ejs-loader",
    enforce: "pre",

    resolveId(id, importer) {
      return id.endsWith(".ejs")
        ? TemplateResolver.resolve(id, importer)
        : null;
    },

    load(id) {
      if (!id.endsWith(".ejs") && !id.endsWith(".ejs.js")) return null;

      try {
        const content = FileSystem.readFile(id);
        return {
          code: id.endsWith(".ejs.js")
            ? content
            : `export default ${JSON.stringify(content)};`,
          map: null,
        };
      } catch (error) {
        console.warn(`Error processing file: ${id}`, error);
        return null;
      }
    },
  };
}
