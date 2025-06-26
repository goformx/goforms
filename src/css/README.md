# CSS Architecture Guide

## Overview

This project uses a layered CSS architecture with PostCSS processing and CSS custom properties for maintainability and consistency.

## Directory Structure

```
src/css/
├── base/              # Foundation styles
│   ├── variables.css  # CSS custom properties
│   └── reset.css      # CSS reset/normalize
├── layouts/           # Layout components
│   ├── container.css
│   ├── grid.css
│   ├── hero.css
│   └── footer.css
├── components/        # Reusable UI components
│   ├── nav.css
│   ├── buttons.css
│   ├── form.css       # Core form elements
│   ├── forms.css      # Form-specific components
│   ├── messages.css
│   └── ...
├── pages/            # Page-specific styles
│   ├── error.css
│   ├── demo/
│   └── form-preview.css
├── dashboard/        # Dashboard-specific styles
│   ├── layout.css
│   ├── forms.css     # Dashboard form components
│   ├── stats.css
│   └── ...
├── themes/           # Theme variations
│   └── dark.css
├── utils/            # Utility classes
│   └── animations.css
├── form-builder.css  # Form builder specific styles
├── formio-custom.css # Form.io customizations
└── main.css          # Main entry point
```

## CSS Layers

The project uses CSS `@layer` directives to control specificity:

1. **base** - Variables, reset, foundation styles
2. **layouts** - Layout components and grid systems
3. **components** - Reusable UI components
4. **pages** - Page-specific styles
5. **utils** - Utility classes and helpers
6. **theme** - Theme variations and overrides

## Naming Conventions

### CSS Classes
- Use kebab-case for class names: `.form-group`, `.nav-item`
- Use BEM methodology for complex components: `.form__input--error`
- Use semantic names that describe purpose, not appearance

### CSS Custom Properties
- Use kebab-case with descriptive prefixes:
  - `--form-*` for form-related properties
  - `--color-*` for color values
  - `--spacing-*` for spacing values
  - `--font-*` for typography values

## Component Guidelines

### Form Components
- `form.css` - Core form elements (inputs, labels, etc.)
- `forms.css` - Form-specific components and layouts
- `dashboard/forms.css` - Dashboard-specific form components

### Separation of Concerns
- Keep core form elements in `components/form.css`
- Put form layouts and complex components in `components/forms.css`
- Place page-specific form styles in respective page directories

## Best Practices

1. **Use CSS Custom Properties** for consistent theming
2. **Follow the layer order** to maintain proper specificity
3. **Keep components focused** on a single responsibility
4. **Use semantic class names** that describe purpose
5. **Avoid deep nesting** - prefer flat, readable selectors
6. **Document complex components** with comments

## Build Process

The CSS is processed with:
- **PostCSS Import** - For importing CSS files
- **PostCSS Nested** - For nested CSS syntax
- **Autoprefixer** - For vendor prefixes
- **CSSNano** - For minification (production only)

## Adding New Styles

1. **Determine the layer** - Where does this style belong?
2. **Choose the right file** - Existing component or new file?
3. **Follow naming conventions** - Use consistent patterns
4. **Update main.css** - Add import if creating new file
5. **Test in context** - Ensure styles work as expected 