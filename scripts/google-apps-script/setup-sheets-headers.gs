/**
 * Google Apps Script to set up ISX Pulse Scratch Card License System Sheets
 * This script creates and configures all required sheets with proper headers
 * 
 * Usage:
 * 1. Open your Google Sheet
 * 2. Go to Extensions > Apps Script
 * 3. Delete any existing code
 * 4. Paste this entire script
 * 5. Save (Ctrl+S or Cmd+S)
 * 6. Run the setupAllSheets() function
 */

/**
 * Main function to set up all sheets
 */
function setupAllSheets() {
  const spreadsheet = SpreadsheetApp.getActiveSpreadsheet();
  
  // Set up each sheet
  setupLicensesSheet(spreadsheet);
  setupActivationAttemptsSheet(spreadsheet);
  setupBlacklistSheet(spreadsheet);
  setupAuditLogSheet(spreadsheet);
  
  // Show completion message
  SpreadsheetApp.getUi().alert(
    'Setup Complete!',
    'All 4 sheets have been created/updated with proper headers.\n\n' +
    'Sheets created:\n' +
    '• Licenses\n' +
    '• ActivationAttempts\n' +
    '• Blacklist\n' +
    '• AuditLog',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

/**
 * Set up the Licenses sheet
 */
function setupLicensesSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('Licenses');
  
  // Create sheet if it doesn't exist
  if (!sheet) {
    sheet = spreadsheet.insertSheet('Licenses');
  }
  
  // Clear existing content
  sheet.clear();
  
  // Set headers
  const headers = [
    'licenseKey',
    'duration', 
    'status',
    'activatedBy',
    'activationDate',
    'expiryDate',
    'deviceFingerprint',
    'activationID',
    'createdDate',
    'batchID',
    'qrCode',
    'notes'
  ];
  
  // Add headers to first row
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#4285f4');
  headerRange.setFontColor('#ffffff');
  headerRange.setFontWeight('bold');
  headerRange.setHorizontalAlignment('center');
  
  // Set column widths
  sheet.setColumnWidth(1, 200); // licenseKey
  sheet.setColumnWidth(2, 80);  // duration
  sheet.setColumnWidth(3, 100); // status
  sheet.setColumnWidth(4, 150); // activatedBy
  sheet.setColumnWidth(5, 150); // activationDate
  sheet.setColumnWidth(6, 150); // expiryDate
  sheet.setColumnWidth(7, 200); // deviceFingerprint
  sheet.setColumnWidth(8, 150); // activationID
  sheet.setColumnWidth(9, 150); // createdDate
  sheet.setColumnWidth(10, 200); // batchID
  sheet.setColumnWidth(11, 100); // qrCode
  sheet.setColumnWidth(12, 200); // notes
  
  // Freeze header row
  sheet.setFrozenRows(1);
  
  console.log('Licenses sheet setup complete');
}

/**
 * Set up the ActivationAttempts sheet
 */
function setupActivationAttemptsSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('ActivationAttempts');
  
  // Create sheet if it doesn't exist
  if (!sheet) {
    sheet = spreadsheet.insertSheet('ActivationAttempts');
  }
  
  // Clear existing content
  sheet.clear();
  
  // Set headers
  const headers = [
    'timestamp',
    'licenseKey',
    'deviceFingerprint',
    'clientIP',
    'userAgent',
    'success',
    'errorType',
    'attemptID'
  ];
  
  // Add headers to first row
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#ea4335');
  headerRange.setFontColor('#ffffff');
  headerRange.setFontWeight('bold');
  headerRange.setHorizontalAlignment('center');
  
  // Set column widths
  sheet.setColumnWidth(1, 150); // timestamp
  sheet.setColumnWidth(2, 200); // licenseKey
  sheet.setColumnWidth(3, 200); // deviceFingerprint
  sheet.setColumnWidth(4, 120); // clientIP
  sheet.setColumnWidth(5, 250); // userAgent
  sheet.setColumnWidth(6, 80);  // success
  sheet.setColumnWidth(7, 150); // errorType
  sheet.setColumnWidth(8, 150); // attemptID
  
  // Freeze header row
  sheet.setFrozenRows(1);
  
  console.log('ActivationAttempts sheet setup complete');
}

/**
 * Set up the Blacklist sheet
 */
function setupBlacklistSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('Blacklist');
  
  // Create sheet if it doesn't exist
  if (!sheet) {
    sheet = spreadsheet.insertSheet('Blacklist');
  }
  
  // Clear existing content
  sheet.clear();
  
  // Set headers
  const headers = [
    'identifier',
    'type',
    'reason',
    'addedDate',
    'addedBy',
    'notes'
  ];
  
  // Add headers to first row
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#fbbc04');
  headerRange.setFontColor('#000000');
  headerRange.setFontWeight('bold');
  headerRange.setHorizontalAlignment('center');
  
  // Set column widths
  sheet.setColumnWidth(1, 200); // identifier
  sheet.setColumnWidth(2, 120); // type
  sheet.setColumnWidth(3, 250); // reason
  sheet.setColumnWidth(4, 150); // addedDate
  sheet.setColumnWidth(5, 150); // addedBy
  sheet.setColumnWidth(6, 300); // notes
  
  // Freeze header row
  sheet.setFrozenRows(1);
  
  console.log('Blacklist sheet setup complete');
}

/**
 * Set up the AuditLog sheet
 */
function setupAuditLogSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('AuditLog');
  
  // Create sheet if it doesn't exist
  if (!sheet) {
    sheet = spreadsheet.insertSheet('AuditLog');
  }
  
  // Clear existing content
  sheet.clear();
  
  // Set headers
  const headers = [
    'timestamp',
    'action',
    'licenseKey',
    'deviceFingerprint',
    'clientIP',
    'userAgent',
    'result',
    'errorDetails',
    'correlationID'
  ];
  
  // Add headers to first row
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#34a853');
  headerRange.setFontColor('#ffffff');
  headerRange.setFontWeight('bold');
  headerRange.setHorizontalAlignment('center');
  
  // Set column widths
  sheet.setColumnWidth(1, 150); // timestamp
  sheet.setColumnWidth(2, 100); // action
  sheet.setColumnWidth(3, 200); // licenseKey
  sheet.setColumnWidth(4, 200); // deviceFingerprint
  sheet.setColumnWidth(5, 120); // clientIP
  sheet.setColumnWidth(6, 250); // userAgent
  sheet.setColumnWidth(7, 100); // result
  sheet.setColumnWidth(8, 300); // errorDetails
  sheet.setColumnWidth(9, 150); // correlationID
  
  // Freeze header row
  sheet.setFrozenRows(1);
  
  console.log('AuditLog sheet setup complete');
}

/**
 * Generate sample licenses for testing
 * Run this after setupAllSheets() to add test data
 */
function generateSampleLicenses() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Licenses');
  
  if (!sheet) {
    SpreadsheetApp.getUi().alert('Error', 'Please run setupAllSheets() first!', SpreadsheetApp.getUi().ButtonSet.OK);
    return;
  }
  
  const batchID = 'BATCH-' + Utilities.formatDate(new Date(), 'GMT', 'yyyyMMdd-HHmmss');
  const createdDate = Utilities.formatDate(new Date(), 'GMT', 'yyyy-MM-dd HH:mm:ss');
  const licenses = [];
  
  // Generate 10 sample licenses
  for (let i = 0; i < 10; i++) {
    const license = [
      generateLicenseKey(),     // licenseKey
      '1m',                      // duration
      'Available',               // status
      '',                        // activatedBy
      '',                        // activationDate
      '',                        // expiryDate
      '',                        // deviceFingerprint
      '',                        // activationID
      createdDate,               // createdDate
      batchID,                   // batchID
      '',                        // qrCode
      'Sample license'           // notes
    ];
    licenses.push(license);
  }
  
  // Add licenses to sheet (starting from row 2)
  if (licenses.length > 0) {
    sheet.getRange(2, 1, licenses.length, licenses[0].length).setValues(licenses);
  }
  
  SpreadsheetApp.getUi().alert(
    'Sample Licenses Generated',
    `Generated ${licenses.length} sample licenses with batch ID: ${batchID}`,
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

/**
 * Generate a random license key in ISX-XXXX-XXXX-XXXX format
 */
function generateLicenseKey() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  let key = 'ISX';
  
  for (let i = 0; i < 3; i++) {
    key += '-';
    for (let j = 0; j < 4; j++) {
      key += chars.charAt(Math.floor(Math.random() * chars.length));
    }
  }
  
  return key;
}

/**
 * Clear all data from sheets (keeps headers)
 */
function clearAllData() {
  const spreadsheet = SpreadsheetApp.getActiveSpreadsheet();
  const sheetNames = ['Licenses', 'ActivationAttempts', 'Blacklist', 'AuditLog'];
  
  sheetNames.forEach(name => {
    const sheet = spreadsheet.getSheetByName(name);
    if (sheet && sheet.getLastRow() > 1) {
      // Clear all rows except header
      sheet.getRange(2, 1, sheet.getLastRow() - 1, sheet.getLastColumn()).clear();
    }
  });
  
  SpreadsheetApp.getUi().alert(
    'Data Cleared',
    'All data has been cleared from all sheets (headers preserved).',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

/**
 * Menu creation for easy access
 */
function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('ISX License System')
    .addItem('Setup All Sheets', 'setupAllSheets')
    .addItem('Generate Sample Licenses', 'generateSampleLicenses')
    .addSeparator()
    .addItem('Clear All Data', 'clearAllData')
    .addToUi();
}