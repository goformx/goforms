declare module "goforms-template" {
  const goforms: {
    framework: string;
    templates: {
      [key: string]: {
        form?: (ctx: any) => string;
        html?: (ctx: any) => string;
        align?: (ctx: any) => string;
      };
    };
    cssClasses: {
      [key: string]: string;
    };
    iconClass: (iconset: string, name: string, spinning?: boolean) => string;
    transform: (type: string, text: string) => string;
  };
  export default goforms;
}
