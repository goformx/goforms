import fs from "fs";
import path from "path";
import type { Plugin } from "vite";

/**
 * Utility to locate EJS template files.
 */
class TemplateFinder {
  static find(id: string): string | null {
    if (id.includes("node_modules/@formio")) {
      const ejsJsPath = id.replace(".ejs", ".ejs.js");
      if (fs.existsSync(ejsJsPath)) {
        return ejsJsPath;
      }
    }

    return this.findRegularTemplate(id);
  }

  private static findRegularTemplate(id: string): string | null {
    const ejsJsPath = id.replace(".ejs", ".ejs.js");
    if (fs.existsSync(ejsJsPath)) {
      return ejsJsPath;
    }
    if (fs.existsSync(id)) {
      return id;
    }
    return null;
  }
}

/**
 * Utility to extract template content from .ejs.js files.
 */
class TemplateExtractor {
  static extract(content: string): string | null {
    try {
      const functionMatch = content.match(
        /exports\.default\s*=\s*function\s*\([^)]*\)\s*{([\s\S]*?)return\s+__p\s*}/,
      );
      if (!functionMatch) return null;

      return this.extractStringConcatenations(functionMatch[1]);
    } catch (error) {
      console.warn("Error extracting template:", error);
      return null;
    }
  }

  private static extractStringConcatenations(
    functionBody: string,
  ): string | null {
    const stringMatches = functionBody.match(/'([^']+)'/g);
    if (!stringMatches) return null;

    return stringMatches
      .map((match) => match.slice(1, -1))
      .filter((str) => str.trim())
      .join("")
      .replace(/\\n/g, "\n")
      .replace(/\\'/g, "'")
      .replace(/\\\\/g, "\\");
  }
}

/**
 * Main Vite plugin for EJS file handling.
 */
export default function ejsPlugin(): Plugin {
  return {
    name: "ejs-loader",
    enforce: "pre",

    resolveId(id, importer) {
      if (!id.endsWith(".ejs")) return null;

      const resolveId = path.isAbsolute(id)
        ? id
        : importer
          ? path.resolve(path.dirname(importer), id)
          : path.resolve(process.cwd(), id);

      return TemplateFinder.find(resolveId);
    },

    load(id) {
      if (!id.endsWith(".ejs") && !id.endsWith(".ejs.js")) return null;

      try {
        const content = fs.readFileSync(id, "utf-8");
        if (id.endsWith(".ejs.js")) {
          const extractedTemplate = TemplateExtractor.extract(content);
          return {
            code: `export default ${JSON.stringify(extractedTemplate ?? content)};`,
            map: null,
          };
        }
        return {
          code: `export default ${JSON.stringify(content)};`,
          map: null,
        };
      } catch (error) {
        console.warn(`Error processing file: ${id}`, error);
        return null;
      }
    },
  };
}
