import fs from "fs";
import path from "path";
import type { Plugin } from "vite";

/**
 * File system utility for checking template existence and reading files.
 * Provides a simple interface for file operations used by the EJS plugin.
 */
class FileSystem {
  /**
   * Checks if a file exists at the given path.
   * @param filePath - The path to check
   * @returns true if the file exists, false otherwise
   */
  static exists(filePath: string): boolean {
    return fs.existsSync(filePath);
  }

  /**
   * Reads the contents of a file as UTF-8 text.
   * @param filePath - The path of the file to read
   * @returns The file contents as a string
   */
  static readFile(filePath: string): string {
    return fs.readFileSync(filePath, "utf-8");
  }
}

/**
 * Template resolution utility for handling EJS template paths.
 * Resolves both raw .ejs files and their compiled .ejs.js counterparts.
 */
class TemplateResolver {
  /**
   * Resolves a template path relative to the importer or current working directory.
   * @param id - The template path to resolve
   * @param importer - Optional path of the importing file
   * @returns The resolved template path or null if not found
   */
  static resolve(id: string, importer?: string): string | null {
    const resolvedPath = path.isAbsolute(id)
      ? id
      : importer
        ? path.resolve(path.dirname(importer), id)
        : path.resolve(process.cwd(), id);

    return this.findTemplate(resolvedPath);
  }

  /**
   * Finds the appropriate template file, preferring compiled .ejs.js files.
   * @param id - The template path to check
   * @returns The path to the template file or null if not found
   */
  private static findTemplate(id: string): string | null {
    const ejsJsPath = id.replace(".ejs", ".ejs.js");

    return FileSystem.exists(ejsJsPath)
      ? ejsJsPath
      : FileSystem.exists(id)
        ? id
        : null;
  }
}

/**
 * Vite plugin for handling EJS templates.
 * Provides seamless integration of EJS templates in the Vite build process.
 *
 * Features:
 * - Resolves both raw .ejs and compiled .ejs.js files
 * - Handles template imports in TypeScript/JavaScript files
 * - Provides proper module exports for template content
 */
export default function ejsPlugin(): Plugin {
  return {
    name: "ejs-loader",
    enforce: "pre", // Run before other plugins to handle template resolution

    /**
     * Resolves .ejs file imports to their actual file paths.
     * @param id - The import path
     * @param importer - The file doing the import
     * @returns The resolved file path or null if not an EJS file
     */
    resolveId(id, importer) {
      return id.endsWith(".ejs")
        ? TemplateResolver.resolve(id, importer)
        : null;
    },

    /**
     * Loads and processes EJS template files.
     * For .ejs.js files, returns the compiled JavaScript directly.
     * For raw .ejs files, wraps the content in a default export.
     *
     * @param id - The file path to load
     * @returns The processed file content or null if not an EJS file
     */
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
      } catch (_error) {
        // Silently handle file processing errors in build tools
        return null;
      }
    },
  };
}
