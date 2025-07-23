# Component Naming Standards

This document defines the standardized naming conventions for all frontend components in the ISX Daily Reports Scrapper application.

## Component Categories

### 1. Layout Components
These are structural components that define the overall page layout.

| Component ID | Template Path | Description |
|-------------|---------------|-------------|
| `header` | `layout/header.html` | Top navigation bar with branding |
| `sidebar` | `layout/sidebar.html` | Left navigation menu |
| `footer` | `layout/footer.html` | Bottom footer with license info |

### 2. Section Components
These are the main content areas that users navigate between.

| Component ID | Template Path | Description |
|-------------|---------------|-------------|
| `scraper` | `sections/scraper.html` | Data collection interface |
| `processor` | `sections/processor.html` | Report processing interface |
| `indexcsv` | `sections/indexcsv.html` | Market indices charts |
| `tickercharts` | `sections/tickercharts.html` | Individual ticker charts |
| `marketmovers` | `sections/marketmovers.html` | Top gainers/losers |
| `files` | `sections/files.html` | File archive browser |

### 3. Reusable Components
These are smaller, reusable UI elements.

| Component ID | Template Path | Description |
|-------------|---------------|-------------|
| `pipeline-stages` | `components/pipeline-stages.html` | Pipeline progress visualization |
| `data-table` | `components/data-table.html` | Generic data table |
| `chart-container` | `components/chart-container.html` | Generic chart wrapper |

## Naming Rules

1. **Component IDs**: Use lowercase with hyphens (kebab-case)
   - ✅ `market-movers`
   - ❌ `marketMovers`, `MarketMovers`

2. **Template Files**: Match component ID exactly with `.html` extension
   - ✅ `market-movers.html`
   - ❌ `marketmovers.html`, `market_movers.html`

3. **Container IDs**: Use component ID with `-container` suffix
   - ✅ `<div id="sidebar-container"></div>`
   - ❌ `<div id="navigation-container"></div>`

4. **CSS Classes**: Use component ID as base class
   - ✅ `.sidebar { ... }`
   - ❌ `.navigation { ... }`

5. **JavaScript References**: Use exact component ID
   - ✅ `componentManager.loadComponent('sidebar', ...)`
   - ❌ `componentManager.loadComponent('navigation', ...)`

## Section Navigation Mapping

The sidebar navigation must use these exact section IDs:

```javascript
// In sidebar.html
<a data-section="scraper">Data Collection</a>
<a data-section="processor">Process Reports</a>
<a data-section="indexcsv">Market Indices</a>
<a data-section="tickercharts">Ticker Charts</a>
<a data-section="marketmovers">Market Movers</a>
<a data-section="files">File Archive</a>
```

## Component Registration

When registering component initializers in `componentManager.js`:

```javascript
// Layout components
this.initializers.set('header', ...);
this.initializers.set('sidebar', ...);
this.initializers.set('footer', ...);

// Section components
this.initializers.set('scraper', ...);
this.initializers.set('processor', ...);
this.initializers.set('indexcsv', ...);
this.initializers.set('tickercharts', ...);
this.initializers.set('marketmovers', ...);
this.initializers.set('files', ...);
```

## Template Path Resolution

The `getTemplatePath()` function in `componentManager.js` must match these arrays:

```javascript
const layoutComponents = ['header', 'sidebar', 'footer'];
const sectionComponents = ['scraper', 'processor', 'indexcsv', 'tickercharts', 'marketmovers', 'files'];
```

## Common Mistakes to Avoid

1. **Don't mix navigation/sidebar terminology**
   - Component is called `sidebar`, not `navigation`
   - Template is `sidebar.html`, not `navigation.html`

2. **Don't create mismatched IDs**
   - If component is `marketmovers`, don't use `market-movers` in some places

3. **Don't forget to update all references**
   - Component ID in JavaScript
   - Template filename
   - Container ID in HTML
   - CSS classes
   - Documentation

## Validation Checklist

When adding or modifying a component:

- [ ] Component ID follows kebab-case convention
- [ ] Template filename matches component ID exactly
- [ ] Container element uses `{component-id}-container` ID
- [ ] CSS classes use component ID as base
- [ ] JavaScript references use exact component ID
- [ ] Component is registered in `componentManager.js`
- [ ] Component is listed in appropriate array in `getTemplatePath()`
- [ ] Documentation is updated with new component

## Implementation Status

As of v0.3.0-alpha:
- ✅ All layout components standardized
- ✅ All section components standardized
- ✅ Component manager configured correctly
- ✅ Template paths match naming convention