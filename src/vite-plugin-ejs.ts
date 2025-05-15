import fs from "fs";
import path from "path";
import type { Plugin } from "vite";

export default function ejsPlugin(): Plugin {
  const findTemplate = (id: string): string | null => {
    // For Form.io templates in node_modules
    if (id.includes("node_modules/@formio")) {
      const ejsJsPath = id.replace(".ejs", ".ejs.js");
      if (fs.existsSync(ejsJsPath)) {
        console.log("Found Form.io template at:", ejsJsPath);
        return ejsJsPath;
      }
      console.log("Form.io template not found at:", ejsJsPath);
    }

    // For any .ejs file, try to find a corresponding .ejs.js file first
    const ejsJsPath = id.replace(".ejs", ".ejs.js");
    if (fs.existsSync(ejsJsPath)) {
      console.log("Found .ejs.js template at:", ejsJsPath);
      return ejsJsPath;
    }

    // Then try the original .ejs file
    if (fs.existsSync(id)) {
      console.log("Found .ejs template at:", id);
      return id;
    }

    return null;
  };

  const extractEjsJsTemplate = (content: string): string | null => {
    try {
      // First try to match the entire function body
      const functionMatch = content.match(
        /exports\.default\s*=\s*function\s*\([^)]*\)\s*{([\s\S]*?)return\s+__p\s*}/,
      );
      if (!functionMatch) {
        return null;
      }

      const functionBody = functionMatch[1];

      // Extract all string concatenations
      const templateParts: string[] = [];
      const stringMatches = functionBody.match(/'([^']+)'/g);

      if (!stringMatches) {
        return null;
      }

      for (const match of stringMatches) {
        // Remove the quotes
        const str = match.slice(1, -1);
        if (str.trim()) {
          templateParts.push(str);
        }
      }

      if (templateParts.length === 0) {
        return null;
      }

      // Join all parts and unescape
      return templateParts
        .join("")
        .replace(/\\n/g, "\n")
        .replace(/\\'/g, "'")
        .replace(/\\\\/g, "\\");
    } catch (error) {
      console.warn("Error extracting template:", error);
      return null;
    }
  };

  return {
    name: "ejs-loader",
    enforce: "pre",
    resolveId(id, importer) {
      if (!id.endsWith(".ejs")) return null;

      let resolveId = id;

      // Handle relative imports
      if ((id.startsWith("./") || id.startsWith("../")) && importer) {
        resolveId = path.resolve(path.dirname(importer), id);
      } else if (!path.isAbsolute(id)) {
        // Handle non-relative, non-absolute imports as relative to cwd
        resolveId = path.resolve(process.cwd(), id);
      }

      const resolvedPath = findTemplate(resolveId);
      if (resolvedPath) {
        return resolvedPath;
      }

      return null;
    },
    load(id) {
      if (!id.endsWith(".ejs") && !id.endsWith(".ejs.js")) return null;

      try {
        const content = fs.readFileSync(id, "utf-8");

        // Handle .ejs.js files (Form.io templates)
        if (id.endsWith(".ejs.js")) {
          const template = extractEjsJsTemplate(content);
          if (template) {
            return {
              code: `export default ${JSON.stringify(template)};`,
              map: null,
            };
          }
          // If extraction fails, return the content as is
          return {
            code: content,
            map: null,
          };
        }

        // Handle regular .ejs files
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
