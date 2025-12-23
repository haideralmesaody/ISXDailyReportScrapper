#!/usr/bin/env node

/**
 * Script to update all SVG files to use currentColor instead of hardcoded colors
 * This enables CSS-based color control for light/dark mode theming
 */

const fs = require('fs');
const path = require('path');

const ICONS_DIR = path.join(__dirname, '..', 'public', 'stock-icons');

// Colors to replace with currentColor
const COLORS_TO_REPLACE = [
  '#4B4B4D',
  '#4b4b4d',
  '#333333',
  '#333',
  '#000000',
  '#000',
  '#666666',
  '#666',
  '#999999',
  '#999'
];

// Process all SVG files in the directory
function processAllSvgs() {
  console.log(`📂 Processing SVGs in: ${ICONS_DIR}`);
  
  if (!fs.existsSync(ICONS_DIR)) {
    console.error('❌ Icons directory not found!');
    process.exit(1);
  }
  
  const files = fs.readdirSync(ICONS_DIR);
  const svgFiles = files.filter(file => file.endsWith('.svg'));
  
  console.log(`📊 Found ${svgFiles.length} SVG files`);
  
  let updatedCount = 0;
  let totalReplacements = 0;
  
  svgFiles.forEach(file => {
    const filePath = path.join(ICONS_DIR, file);
    const { updated, replacements } = processSvgFile(filePath);
    
    if (updated) {
      updatedCount++;
      totalReplacements += replacements;
      console.log(`✅ Updated: ${file} (${replacements} replacements)`);
    } else {
      console.log(`⏭️  Skipped: ${file} (already using currentColor or no matches)`);
    }
  });
  
  console.log('\n📊 Summary:');
  console.log(`   Files updated: ${updatedCount}/${svgFiles.length}`);
  console.log(`   Total replacements: ${totalReplacements}`);
  console.log('✨ Done!');
}

// Process a single SVG file
function processSvgFile(filePath) {
  let content = fs.readFileSync(filePath, 'utf8');
  const originalContent = content;
  let replacements = 0;
  
  // Replace fill attributes with hardcoded colors
  COLORS_TO_REPLACE.forEach(color => {
    const fillRegex = new RegExp(`fill="${color}"`, 'gi');
    const matches = content.match(fillRegex);
    if (matches) {
      replacements += matches.length;
      content = content.replace(fillRegex, 'fill="currentColor"');
    }
  });
  
  // Replace stroke attributes with hardcoded colors
  COLORS_TO_REPLACE.forEach(color => {
    const strokeRegex = new RegExp(`stroke="${color}"`, 'gi');
    const matches = content.match(strokeRegex);
    if (matches) {
      replacements += matches.length;
      content = content.replace(strokeRegex, 'stroke="currentColor"');
    }
  });
  
  // Also handle CSS style attributes (less common but might exist)
  COLORS_TO_REPLACE.forEach(color => {
    const styleRegex = new RegExp(`fill:\\s*${color}`, 'gi');
    const matches = content.match(styleRegex);
    if (matches) {
      replacements += matches.length;
      content = content.replace(styleRegex, 'fill: currentColor');
    }
    
    const strokeStyleRegex = new RegExp(`stroke:\\s*${color}`, 'gi');
    const strokeMatches = content.match(strokeStyleRegex);
    if (strokeMatches) {
      replacements += strokeMatches.length;
      content = content.replace(strokeStyleRegex, 'stroke: currentColor');
    }
  });
  
  // Only write if content changed
  if (content !== originalContent) {
    fs.writeFileSync(filePath, content, 'utf8');
    return { updated: true, replacements };
  }
  
  return { updated: false, replacements: 0 };
}

// Run the script
console.log('🎨 SVG Color Update Script');
console.log('=============================');
processAllSvgs();