// ============================================
// ISX LICENSE SYSTEM - GOOGLE APPS SCRIPT V3.5 (ENHANCED)
// ============================================
// Version: 3.5.0 - Fixed validation and action routing
// Last Updated: 2025-09-02
// Changes from V3.4:
//   - Fixed action routing with case-insensitive matching
//   - Enhanced validation handler for proper parameter extraction
//   - Added support for validation without activation ID
//   - Improved error logging and debugging
// ============================================
// CRITICAL: This version reads licenses dynamically from Google Sheets
// No need to update the script when adding new licenses
// ============================================

// ============================================
// CONFIGURATION - Your Sheet ID and Settings
// ============================================

const SHEET_ID = '1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc';
const MAX_ATTEMPTS_PER_HOUR = 10;
const BLOCK_DURATION_HOURS = 24;

// Shared secret for HMAC verification (must match embedded Go credentials)
// This MUST match the secret in your Go backend's embedded_credentials.go
const SHARED_SECRET = 'ISX-Pulse-S3cur3-K3y-2024-@lm3s@0dy';

// Test license for development/testing
const TEST_LICENSE_CODE = 'ISX-X7MY-V4QB-UZG3-UYKN';

// ============================================
// HELPER FUNCTIONS
// ============================================

/**
 * Sorts object keys alphabetically (recursive for nested objects)
 * This ensures consistent JSON stringification with Go's json.Marshal
 * @param {Object} obj - The object to sort
 * @returns {Object} Object with sorted keys
 */
function sortObjectKeys(obj) {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }
  
  if (Array.isArray(obj)) {
    return obj.map(sortObjectKeys);
  }
  
  const sorted = {};
  Object.keys(obj).sort().forEach(key => {
    sorted[key] = sortObjectKeys(obj[key]);
  });
  return sorted;
}

/**
 * Parse license duration string (e.g., '30d', '1m', '3m', '6m', '1y')
 * @param {string} duration - Duration string
 * @returns {Date} Expiration date
 */
function parseDuration(duration) {
  const now = new Date();
  const match = duration.match(/^(\d+)([dmyh])$/);
  
  if (!match) {
    // Default to 30 days if invalid format
    return new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
  }
  
  const [, value, unit] = match;
  const num = parseInt(value, 10);
  
  switch (unit) {
    case 'h':
      return new Date(now.getTime() + num * 60 * 60 * 1000);
    case 'd':
      return new Date(now.getTime() + num * 24 * 60 * 60 * 1000);
    case 'm':
      now.setMonth(now.getMonth() + num);
      return now;
    case 'y':
      now.setFullYear(now.getFullYear() + num);
      return now;
    default:
      return new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
  }
}

/**
 * Get license duration from key prefix or from sheet data
 * @param {string} key - License key
 * @param {string} sheetDuration - Duration from sheet (optional)
 * @returns {string} Duration string
 */
function getLicenseDuration(key, sheetDuration) {
  // First check if we have duration from sheet
  if (sheetDuration && sheetDuration !== '') {
    return sheetDuration;
  }
  
  // Otherwise, derive from key prefix
  if (key.startsWith('ISX1M-')) return '1m';
  if (key.startsWith('ISX3M-')) return '3m';
  if (key.startsWith('ISX6M-')) return '6m';
  if (key.startsWith('ISX1Y-')) return '1y';
  return '30d'; // Default 30 days for regular keys
}

// ============================================
// MAIN ENTRY POINT - Handles POST requests
// ============================================

function doPost(e) {
  const lock = LockService.getScriptLock();
  
  try {
    // Acquire lock for atomic operation (wait max 10 seconds)
    lock.waitLock(10000);
    
    // Log incoming request for debugging
    console.log('=== Incoming POST request ===');
    console.log('Raw content:', e.postData.contents);
    
    // Parse request
    const requestData = JSON.parse(e.postData.contents);
    console.log('Parsed request:', JSON.stringify(requestData, null, 2));
    
    // Handle signed requests from ISX Pulse Go backend
    if (requestData.payload) {
      // This is a signed request from the Go backend
      console.log('Processing signed request from ISX Pulse Go backend');
      
      const payload = requestData.payload;
      const fingerprint = requestData.fingerprint || '';
      const signature = requestData.signature || '';
      const timestamp = requestData.timestamp || 0;
      const nonce = requestData.nonce || '';
      const requestId = requestData.request_id || '';
      
      // Log request details
      console.log('Request ID:', requestId);
      console.log('Action:', payload.action);
      console.log('Fingerprint:', fingerprint);
      
      // Verify request signature if provided
      if (signature) {
        const isValid = verifyRequestSignature(requestData);
        if (!isValid) {
          console.error('Request signature verification failed');
          return createSignedResponse(false, 'Invalid request signature', null, requestId);
        }
        console.log('Request signature verified successfully');
      }
      
      // Extract action from payload - normalize to lowercase for consistency
      const action = (payload.action || 'activate').toLowerCase();
      console.log('Processing action (normalized):', action);
      
      // Create a compatible request object for our handlers
      const compatibleRequest = {
        action: action,
        code: payload.code || payload.license_key,
        license_key: payload.code || payload.license_key,
        activation_id: payload.activation_id || payload.activationId,
        activationId: payload.activation_id || payload.activationId,
        deviceInfo: payload.deviceInfo || {
          fingerprint: fingerprint,
          ip: e.parameter.ip || 'unknown',
          userAgent: e.parameter['user-agent'] || '',
          requestId: requestId
        },
        // Pass the original payload for validation handler
        payload: payload,
        fingerprint: fingerprint
      };
      
      // Route to appropriate handler based on action
      switch(action) {
        case 'activate':
        case 'activation':
          return handleActivation(compatibleRequest, requestId);
        case 'validate':
        case 'validation':
        case 'verify':
          return handleValidation(compatibleRequest, requestId);
        case 'revoke':
        case 'revocation':
          return handleRevocation(compatibleRequest, requestId);
        case 'checkstatus':
        case 'check_status':
        case 'status':
          return handleStatusCheck(compatibleRequest, requestId);
        case 'ping':
          return handlePing(requestId);
        default:
          console.error('Unknown action:', action);
          console.error('Original action value:', payload.action);
          return createSignedResponse(false, 'Unknown action: ' + action, null, requestId);
      }
    }
    
    // Handle direct requests (not from Go backend) - for testing
    const action = (requestData.action || 'activate').toLowerCase();
    console.log('Processing direct request with action:', action);
    
    // Route to appropriate handler
    switch(action) {
      case 'activate':
      case 'activation':
        return handleActivation(requestData, '');
      case 'validate':
      case 'validation':
      case 'verify':
        return handleValidation(requestData, '');
      case 'revoke':
      case 'revocation':
        return handleRevocation(requestData, '');
      case 'checkstatus':
      case 'check_status':
      case 'status':
        return handleStatusCheck(requestData, '');
      case 'ping':
        return handlePing('');
      default:
        return createSignedResponse(false, 'Unknown action: ' + action, null, '');
    }
    
  } catch (error) {
    console.error('Error in doPost:', error.toString());
    console.error('Stack trace:', error.stack);
    return createSignedResponse(false, 'Server error: ' + error.toString(), null, '');
  } finally {
    lock.releaseLock();
  }
}

// ============================================
// HMAC SIGNATURE VERIFICATION AND GENERATION
// ============================================

/**
 * Verify the HMAC signature of an incoming request
 * @param {Object} requestData - The parsed request data
 * @returns {boolean} True if signature is valid
 */
function verifyRequestSignature(requestData) {
  try {
    if (!requestData.signature) {
      console.log('No signature provided in request');
      return false;
    }
    
    // Create canonical string matching Go backend format
    const canonical = `${requestData.timestamp}|${requestData.nonce}|${requestData.request_id}|${requestData.fingerprint}|${JSON.stringify(requestData.payload)}`;
    
    // Compute HMAC-SHA256
    const signature = Utilities.computeHmacSha256Signature(canonical, SHARED_SECRET);
    const expectedSignature = Utilities.base64Encode(signature);
    
    // Compare signatures
    const isValid = expectedSignature === requestData.signature;
    
    if (!isValid) {
      console.log('Signature mismatch');
      console.log('Expected:', expectedSignature);
      console.log('Received:', requestData.signature);
    }
    
    return isValid;
  } catch (error) {
    console.error('Error verifying request signature:', error);
    return false;
  }
}

/**
 * Generate HMAC signature for response
 * Matches the Go backend's expected format exactly
 * CRITICAL: timestamp must be Unix epoch in SECONDS as NUMBER, not milliseconds
 * @param {Object} responseData - The response data to sign
 * @returns {string} Base64-encoded HMAC signature
 */
function generateResponseSignature(responseData) {
  try {
    // CRITICAL FIX: The canonical string MUST use the numeric timestamp in SECONDS
    // Go backend expects: fmt.Sprintf("%d|%s|%t", resp.Timestamp, resp.RequestID, resp.Success)
    // Format: timestamp|request_id|success[|data_json][|error]
    let canonical = `${responseData.timestamp}|${responseData.request_id}|${responseData.success}`;
    
    // Add data if present (must be JSON string with SORTED keys to match Go's json.Marshal)
    if (responseData.data && Object.keys(responseData.data).length > 0) {
      // CRITICAL FIX: Sort the keys to match Go's json.Marshal behavior
      const sortedData = sortObjectKeys(responseData.data);
      canonical += '|' + JSON.stringify(sortedData);
    }
    
    // Add error if present
    if (responseData.error) {
      canonical += '|' + responseData.error;
    }
    
    console.log('Canonical string for signature:', canonical);
    console.log('Timestamp type:', typeof responseData.timestamp);
    console.log('Timestamp value (seconds):', responseData.timestamp);
    console.log('Data keys sorted:', responseData.data ? Object.keys(sortObjectKeys(responseData.data)) : 'no data');
    
    // Compute HMAC-SHA256 signature
    const signatureBytes = Utilities.computeHmacSha256Signature(canonical, SHARED_SECRET);
    const signature = Utilities.base64Encode(signatureBytes);
    
    console.log('Generated signature:', signature);
    
    return signature;
  } catch (error) {
    console.error('Error generating response signature:', error);
    return '';
  }
}

// ============================================
// RESPONSE CREATION WITH PROPER HMAC SIGNING
// ============================================

/**
 * Create a properly signed response matching Go backend expectations
 * CRITICAL: timestamp MUST be Unix epoch SECONDS as a NUMBER
 * @param {boolean} success - Whether the operation succeeded
 * @param {string} message - Error message or success message
 * @param {Object} data - Response data
 * @param {string} requestId - The request ID from the original request
 * @returns {GoogleAppsScript.Content.TextOutput} The JSON response
 */
function createSignedResponse(success, message, data, requestId) {
  try {
    // CRITICAL FIX: Use Math.floor(Date.now() / 1000) for Unix epoch in SECONDS
    // This matches what Go backend expects when it does time.Unix(resp.Timestamp, 0)
    const timestampSec = Math.floor(Date.now() / 1000);
    
    // Create response matching the Go backend's SignedResponse structure
    const response = {
      timestamp: timestampSec,  // Unix timestamp in SECONDS as NUMBER
      request_id: requestId || Utilities.getUuid(),
      success: success,
      data: data || {},
      error: success ? '' : message,
      signature: '' // Will be calculated below
    };
    
    // Add message to data if successful
    if (success && message) {
      response.data.message = message;
    }
    
    // Generate HMAC signature for the response
    response.signature = generateResponseSignature(response);
    
    console.log('=== Sending signed response ===');
    console.log('Timestamp (seconds):', timestampSec);
    console.log('Timestamp type:', typeof timestampSec);
    console.log('Response:', JSON.stringify(response, null, 2));
    
    // Return as JSON
    return ContentService.createTextOutput(
      JSON.stringify(response)
    ).setMimeType(ContentService.MimeType.JSON);
    
  } catch (error) {
    console.error('Error creating signed response:', error);
    
    // Fallback response without signature - also use SECONDS
    const fallbackResponse = {
      timestamp: Math.floor(Date.now() / 1000),  // Still use SECONDS as number
      request_id: requestId || '',
      success: false,
      data: {},
      error: 'Failed to create signed response: ' + error.toString(),
      signature: ''
    };
    
    return ContentService.createTextOutput(
      JSON.stringify(fallbackResponse)
    ).setMimeType(ContentService.MimeType.JSON);
  }
}

// ============================================
// LICENSE ACTIVATION HANDLER
// ============================================

function handleActivation(request, requestId) {
  try {
    // Extract license code - support both 'code' and 'license_key' fields
    let code = request.code || request.license_key;
    const deviceInfo = request.deviceInfo || {};
    
    // Additional extraction for nested payload structure
    if (request.payload) {
      code = request.payload.code || request.payload.license_key || code;
    }
    
    console.log('=== Activation Request ===');
    console.log('License code:', code);
    console.log('Device fingerprint:', deviceInfo.fingerprint);
    console.log('Device IP:', deviceInfo.ip);
    console.log('Request ID:', requestId);
    
    // Validate license format
    if (!code || !code.startsWith('ISX-')) {
      logActivationAttempt(code, deviceInfo, false, 'Invalid format');
      return createSignedResponse(false, 'Invalid license format. Expected: ISX-XXXX-XXXX-XXXX-XXXX', null, requestId);
    }
    
    // Check blacklist first
    if (isBlacklisted(deviceInfo.ip) || isBlacklisted(deviceInfo.fingerprint)) {
      logActivationAttempt(code, deviceInfo, false, 'Blacklisted');
      return createSignedResponse(false, 'Access denied', null, requestId);
    }
    
    // Check rate limiting
    if (!checkRateLimit(deviceInfo.ip)) {
      addToBlacklist(deviceInfo.ip, 'IP', 'Rate limit exceeded');
      logActivationAttempt(code, deviceInfo, false, 'Rate limited');
      return createSignedResponse(false, 'Too many attempts. Try again later.', null, requestId);
    }
    
    // Get licenses sheet
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
    if (!sheet) {
      console.error('Licenses sheet not found!');
      return createSignedResponse(false, 'System error: Licenses sheet not found. Run setupAllSheets() first.', null, requestId);
    }
    
    const dataRange = sheet.getDataRange();
    const values = dataRange.getValues();
    
    // CRITICAL CHANGE: Look for the license in the sheet DIRECTLY
    // No hardcoded array check - all licenses come from the sheet
    let licenseFound = false;
    let licenseRow = -1;
    
    for (let i = 1; i < values.length; i++) {
      if (values[i][0] === code) { // Column A: Code
        licenseFound = true;
        licenseRow = i;
        console.log('License found at row:', i + 1);
        console.log('Current status:', values[i][2]);
        break;
      }
    }
    
    // If license not found in sheet, it's invalid
    if (!licenseFound) {
      console.log('License not found in sheet:', code);
      logActivationAttempt(code, deviceInfo, false, 'Invalid code');
      return createSignedResponse(false, 'Invalid license code', null, requestId);
    }
    
    // License exists in sheet - proceed with activation logic
    const currentStatus = values[licenseRow][2]; // Column C: Status
    
    if (currentStatus !== 'Available') {
      // License is already activated - check if it's the same device
      const storedFingerprint = values[licenseRow][5]; // Column F: DeviceFingerprint
      const incomingFingerprint = deviceInfo.fingerprint || '';
      
      console.log('License already activated');
      console.log('Stored fingerprint:', storedFingerprint);
      console.log('Incoming fingerprint:', incomingFingerprint);
      
      if (storedFingerprint && incomingFingerprint) {
        // Calculate fingerprint similarity
        const similarity = calculateFingerprintSimilarity(storedFingerprint, incomingFingerprint);
        console.log('Fingerprint similarity:', Math.round(similarity * 100) + '%');
        
        if (similarity >= 0.80) {
          // Same device - allow reactivation
          console.log('Same device detected, allowing reactivation');
          
          const reactivationResult = checkReactivationLimits(code, values[licenseRow]);
          
          if (reactivationResult.allowed) {
            // Update for reactivation
            const now = new Date();
            const reactivationCount = (values[licenseRow][10] || 0) + 1;
            const duration = values[licenseRow][1] || '1m'; // Column B: Duration
            const expiryDate = calculateExpiryDate(duration);
            
            sheet.getRange(licenseRow + 1, 4).setValue(now); // D: ActivationDate
            sheet.getRange(licenseRow + 1, 8).setValue(expiryDate); // H: ExpiryDate
            sheet.getRange(licenseRow + 1, 10).setValue(now); // J: LastChecked
            sheet.getRange(licenseRow + 1, 11).setValue(reactivationCount); // K: CheckCount
            
            logActivationAttempt(code, deviceInfo, true, 'Reactivation success');
            logAudit('REACTIVATION', code, deviceInfo.ip, `Reactivated (${reactivationCount}/5)`);
            
            return createSignedResponse(true, 'License reactivated successfully', {
              status: 'reactivated',
              license_key: code,
              activation_id: values[licenseRow][8],
              expires_at: expiryDate.toISOString(),
              device_id: deviceInfo.fingerprint || '',
              duration: duration,
              reactivation_count: reactivationCount,
              similarity_score: Math.round(similarity * 100),
              features: ['all']
            }, requestId);
          } else {
            // Reactivation limit exceeded
            logActivationAttempt(code, deviceInfo, false, 'Reactivation limit exceeded');
            return createSignedResponse(false, reactivationResult.reason, {
              status: 'reactivation_blocked',
              attempts_used: reactivationResult.attemptsUsed,
              max_attempts: 5,
              reset_date: reactivationResult.resetDate
            }, requestId);
          }
        }
      }
      
      // Different device
      logActivationAttempt(code, deviceInfo, false, 'Already activated');
      return createSignedResponse(false, 'License already activated on a different device', {
        status: 'already_activated',
        activation_date: values[licenseRow][3] ? new Date(values[licenseRow][3]).toISOString() : '',
        expiry_date: values[licenseRow][7] ? new Date(values[licenseRow][7]).toISOString() : ''
      }, requestId);
    }
    
    // License is available - activate it
    console.log('Activating available license');
    
    // Generate unique activation ID
    const activationId = Utilities.getUuid();
    const now = new Date();
    const duration = values[licenseRow][1] || '1m'; // Column B: Duration
    const expiryDate = calculateExpiryDate(duration);
    
    console.log('Activation ID:', activationId);
    console.log('Duration:', duration);
    console.log('Expiry date:', expiryDate);
    
    // Update the license row
    sheet.getRange(licenseRow + 1, 3).setValue('Activated'); // C: Status
    sheet.getRange(licenseRow + 1, 4).setValue(now); // D: ActivationDate
    sheet.getRange(licenseRow + 1, 5).setValue(deviceInfo.ip || 'unknown'); // E: ActivationIP
    sheet.getRange(licenseRow + 1, 6).setValue(deviceInfo.fingerprint || ''); // F: DeviceFingerprint
    sheet.getRange(licenseRow + 1, 7).setValue(deviceInfo.email || ''); // G: Email
    sheet.getRange(licenseRow + 1, 8).setValue(expiryDate); // H: ExpiryDate
    sheet.getRange(licenseRow + 1, 9).setValue(activationId); // I: ActivationID
    sheet.getRange(licenseRow + 1, 10).setValue(now); // J: LastChecked
    sheet.getRange(licenseRow + 1, 11).setValue(1); // K: CheckCount
    
    // Log successful activation
    logActivationAttempt(code, deviceInfo, true, 'Success');
    logAudit('ACTIVATION', code, deviceInfo.ip, 'License activated successfully');
    
    console.log('License activated successfully');
    
    // Return success response
    return createSignedResponse(true, 'License activated successfully', {
      status: 'activated',
      license_key: code,
      activation_id: activationId,
      expires_at: expiryDate.toISOString(),
      device_id: deviceInfo.fingerprint || '',
      duration: duration,
      features: ['all']
    }, requestId);
    
  } catch (error) {
    console.error('Error in handleActivation:', error);
    console.error('Stack trace:', error.stack);
    return createSignedResponse(false, 'Activation failed: ' + error.toString(), null, requestId);
  }
}

// ============================================
// LICENSE VALIDATION HANDLER (ENHANCED)
// ============================================

function handleValidation(request, requestId) {
  try {
    // Extract parameters from nested payload if it exists (signed request from Go backend)
    let code, activationId, deviceFingerprint;
    
    if (request.payload) {
      // This is a signed request from Go backend
      code = request.payload.code || request.payload.license_key;
      activationId = request.payload.activation_id || request.payload.activationId;
      deviceFingerprint = request.fingerprint || request.payload.device_fingerprint || '';
    } else {
      // Direct request (for testing)
      code = request.code || request.license_key;
      activationId = request.activationId || request.activation_id;
      deviceFingerprint = request.deviceFingerprint || request.fingerprint || '';
    }
    
    console.log('=== Validation Request ===');
    console.log('License code:', code);
    console.log('Activation ID:', activationId);
    console.log('Device fingerprint:', deviceFingerprint);
    console.log('Request payload:', request.payload ? JSON.stringify(request.payload) : 'none');
    
    // Validation can work with just the license key if no activation ID is provided
    if (!code) {
      return createSignedResponse(false, 'License key is required', null, requestId);
    }
    
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
    if (!sheet) {
      return createSignedResponse(false, 'System error: Licenses sheet not found', null, requestId);
    }
    
    const dataRange = sheet.getDataRange();
    const values = dataRange.getValues();
    
    for (let i = 1; i < values.length; i++) {
      // Match by license code first
      if (values[i][0] === code) {
        console.log('License found at row:', i + 1);
        
        // If activation ID is provided, it must match; otherwise just validate the license
        if (activationId && values[i][8] !== activationId) {
          console.log('Activation ID mismatch. Expected:', values[i][8], 'Got:', activationId);
          continue; // This license exists but activation ID doesn't match
        }
        
        const status = values[i][2]; // Column C: Status
        const expiryDate = values[i][7] ? new Date(values[i][7]) : null; // Column H: ExpiryDate
        const storedFingerprint = values[i][5]; // Column F: DeviceFingerprint
        const storedActivationId = values[i][8]; // Column I: ActivationID
        
        console.log('License status:', status);
        console.log('Expiry date:', expiryDate);
        
        // Check if license is activated
        if (status !== 'Activated') {
          console.log('License not activated. Current status:', status);
          return createSignedResponse(false, 'License not activated', {
            status: 'not_activated',
            license_status: status
          }, requestId);
        }
        
        // Check expiry
        if (expiryDate && new Date() > expiryDate) {
          console.log('License expired. Expiry date was:', expiryDate);
          sheet.getRange(i + 1, 3).setValue('Expired'); // Update status to Expired
          return createSignedResponse(false, 'License expired', {
            status: 'expired',
            expiry_date: expiryDate.toISOString()
          }, requestId);
        }
        
        // Optionally verify device fingerprint if provided
        let similarity = 100;
        if (deviceFingerprint && storedFingerprint) {
          similarity = calculateFingerprintSimilarity(storedFingerprint, deviceFingerprint);
          console.log('Fingerprint similarity:', Math.round(similarity * 100) + '%');
          
          if (similarity < 0.80) {
            console.warn('Device fingerprint mismatch (similarity: ' + Math.round(similarity * 100) + '%), but allowing validation');
            // You might want to log this as a suspicious activity
            logAudit('VALIDATION_WARNING', code, deviceFingerprint, 'Low fingerprint similarity: ' + Math.round(similarity * 100) + '%');
          }
        }
        
        // Update last checked and increment check count
        sheet.getRange(i + 1, 10).setValue(new Date()); // Column J: LastChecked
        const currentCheckCount = values[i][10] || 0; // Column K: CheckCount
        const newCheckCount = currentCheckCount + 1;
        sheet.getRange(i + 1, 11).setValue(newCheckCount);
        
        console.log('Validation successful. Check count:', newCheckCount);
        
        // Return success with all relevant information
        return createSignedResponse(true, 'License valid', {
          status: 'valid',
          license_status: status,
          license_key: code,
          activation_id: storedActivationId || '',
          expires_at: expiryDate ? expiryDate.toISOString() : null,
          check_count: newCheckCount,
          device_match: deviceFingerprint ? Math.round(similarity * 100) : 100,
          features: ['all']
        }, requestId);
      }
    }
    
    // No matching license found
    console.log('No matching license found for code:', code);
    return createSignedResponse(false, 'Invalid license or activation ID', null, requestId);
    
  } catch (error) {
    console.error('Error in handleValidation:', error);
    console.error('Stack trace:', error.stack);
    return createSignedResponse(false, 'Validation failed: ' + error.toString(), null, requestId);
  }
}

// ============================================
// STATUS CHECK HANDLER
// ============================================

function handleStatusCheck(request, requestId) {
  try {
    // Extract parameters properly
    let code;
    
    if (request.payload) {
      code = request.payload.code || request.payload.license_key;
    } else {
      code = request.code || request.license_key;
    }
    
    console.log('=== Status Check Request ===');
    console.log('License code:', code);
    
    if (!code) {
      return createSignedResponse(false, 'License key is required', null, requestId);
    }
    
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
    if (!sheet) {
      return createSignedResponse(false, 'System error: Licenses sheet not found', null, requestId);
    }
    
    const dataRange = sheet.getDataRange();
    const values = dataRange.getValues();
    
    // Look for the license in the sheet
    for (let i = 1; i < values.length; i++) {
      if (values[i][0] === code) {
        const status = values[i][2]; // Column C: Status
        const activatedAt = values[i][3]; // Column D: ActivationDate
        const expiryDate = values[i][7]; // Column H: ExpiryDate
        const deviceId = values[i][5]; // Column F: DeviceFingerprint
        const activationId = values[i][8]; // Column I: ActivationID
        const checkCount = values[i][10] || 0; // Column K: CheckCount
        
        return createSignedResponse(true, 'Status retrieved successfully', {
          status: status,
          activated_at: activatedAt ? new Date(activatedAt).toISOString() : null,
          expires_at: expiryDate ? new Date(expiryDate).toISOString() : null,
          device_id: deviceId,
          activation_id: activationId,
          check_count: checkCount,
          features: ['all']
        }, requestId);
      }
    }
    
    // License not found in sheet
    return createSignedResponse(false, 'License not found', null, requestId);
    
  } catch (error) {
    console.error('Error in handleStatusCheck:', error);
    return createSignedResponse(false, 'Status check failed: ' + error.toString(), null, requestId);
  }
}

// ============================================
// OTHER HANDLERS
// ============================================

function handleRevocation(request, requestId) {
  try {
    // Extract parameters
    let code, reason;
    
    if (request.payload) {
      code = request.payload.code || request.payload.license_key;
      reason = request.payload.reason || 'Manual revocation';
    } else {
      code = request.code || request.license_key;
      reason = request.reason || 'Manual revocation';
    }
    
    console.log('=== Revocation Request ===');
    console.log('License code:', code);
    console.log('Reason:', reason);
    
    if (!code) {
      return createSignedResponse(false, 'License key is required', null, requestId);
    }
    
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
    if (!sheet) {
      return createSignedResponse(false, 'System error: Licenses sheet not found', null, requestId);
    }
    
    const dataRange = sheet.getDataRange();
    const values = dataRange.getValues();
    
    for (let i = 1; i < values.length; i++) {
      if (values[i][0] === code) {
        // Update status to Revoked
        sheet.getRange(i + 1, 3).setValue('Revoked'); // Column C: Status
        sheet.getRange(i + 1, 14).setValue(reason + ' - ' + new Date().toISOString()); // Column N: Notes
        
        logAudit('REVOCATION', code, 'System', reason);
        
        return createSignedResponse(true, 'License revoked successfully', {
          status: 'revoked',
          license_key: code,
          reason: reason
        }, requestId);
      }
    }
    
    return createSignedResponse(false, 'License not found', null, requestId);
    
  } catch (error) {
    console.error('Error in handleRevocation:', error);
    return createSignedResponse(false, 'Revocation failed: ' + error.toString(), null, requestId);
  }
}

function handlePing(requestId) {
  // Simple ping/pong for connectivity testing
  // Note: Keep timestamp as ISO string here for human readability in ping response data
  return createSignedResponse(true, 'pong', {
    timestamp: new Date().toISOString(),
    server: 'Google Apps Script',
    version: '3.5.0'
  }, requestId);
}

// ============================================
// TEST ENDPOINT - Handles GET requests
// ============================================

function doGet(e) {
  return ContentService.createTextOutput(
    JSON.stringify({
      status: 'OK',
      message: 'ISX License Manager API v3.5.0',
      version: '3.5.0',
      features: [
        'Dynamic license validation from Google Sheets',
        'No hardcoded licenses',
        'HMAC signing with sorted JSON keys',
        'Device fingerprinting',
        'Smart reactivation',
        'Fixed timestamp to SECONDS',
        'Enhanced validation without activation ID',
        'Case-insensitive action routing'
      ],
      sheet_id: SHEET_ID,
      test_license: TEST_LICENSE_CODE,
      timestamp: new Date().toISOString()
    })
  ).setMimeType(ContentService.MimeType.JSON);
}

// ============================================
// HELPER FUNCTIONS
// ============================================

function calculateExpiryDate(duration) {
  const now = new Date();
  let months = 1;
  
  // Parse duration string if it's in the format '1m', '3m', etc.
  if (typeof duration === 'string') {
    const match = duration.match(/^(\d+)([myhd])$/);
    if (match) {
      const [, value, unit] = match;
      const num = parseInt(value, 10);
      
      switch(unit) {
        case 'd': // days
          now.setDate(now.getDate() + num);
          return now;
        case 'm': // months
          months = num;
          break;
        case 'y': // years
          months = num * 12;
          break;
        case 'h': // hours
          now.setHours(now.getHours() + num);
          return now;
        default:
          months = 1;
      }
    }
  }
  
  now.setMonth(now.getMonth() + months);
  now.setDate(now.getDate() + 1);
  now.setHours(0, 0, 0, 0);
  
  return now;
}

function calculateFingerprintSimilarity(fp1, fp2) {
  if (!fp1 || !fp2) return 0.0;
  
  const norm1 = fp1.toLowerCase().replace(/\s+/g, '');
  const norm2 = fp2.toLowerCase().replace(/\s+/g, '');
  
  if (norm1 === norm2) return 1.0;
  
  // Create bigrams
  const getBigrams = (str) => {
    const bigrams = new Set();
    for (let i = 0; i < str.length - 1; i++) {
      bigrams.add(str.substring(i, i + 2));
    }
    return bigrams;
  };
  
  const bigrams1 = getBigrams(norm1);
  const bigrams2 = getBigrams(norm2);
  
  const intersection = new Set([...bigrams1].filter(x => bigrams2.has(x)));
  const union = new Set([...bigrams1, ...bigrams2]);
  
  if (union.size === 0) return 0.0;
  
  return intersection.size / union.size;
}

function checkReactivationLimits(code, licenseRow) {
  const maxReactivations = 5;
  const currentCount = licenseRow[10] || 0;
  const lastChecked = licenseRow[9];
  
  if (currentCount === 0) {
    return { allowed: true, attemptsUsed: 0 };
  }
  
  const now = new Date();
  const thirtyDaysAgo = new Date(now.getTime() - (30 * 24 * 60 * 60 * 1000));
  
  let effectiveCount = currentCount;
  if (lastChecked && new Date(lastChecked) < thirtyDaysAgo) {
    effectiveCount = 0;
  }
  
  if (effectiveCount >= maxReactivations) {
    const resetDate = lastChecked ? 
      new Date(new Date(lastChecked).getTime() + (30 * 24 * 60 * 60 * 1000)) : null;
    
    return {
      allowed: false,
      reason: `Maximum ${maxReactivations} reactivations per 30 days exceeded`,
      attemptsUsed: effectiveCount,
      resetDate: resetDate ? resetDate.toISOString() : null
    };
  }
  
  return { allowed: true, attemptsUsed: effectiveCount };
}

function checkRateLimit(ip) {
  if (!ip || ip === 'unknown') return true;
  
  try {
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('ActivationAttempts');
    if (!sheet) return true;
    
    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    
    const data = sheet.getDataRange().getValues();
    let recentAttempts = 0;
    
    for (let i = 1; i < data.length; i++) {
      if (data[i][2] === ip && new Date(data[i][0]) > oneHourAgo) {
        recentAttempts++;
      }
    }
    
    return recentAttempts < MAX_ATTEMPTS_PER_HOUR;
  } catch (error) {
    console.error('Error checking rate limit:', error);
    return true;
  }
}

function isBlacklisted(identifier) {
  if (!identifier || identifier === 'unknown') return false;
  
  try {
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Blacklist');
    if (!sheet) return false;
    
    const data = sheet.getDataRange().getValues();
    
    for (let i = 1; i < data.length; i++) {
      if (data[i][0] === identifier) {
        if (data[i][5] && new Date(data[i][5]) < new Date()) {
          continue;
        }
        return true;
      }
    }
    
    return false;
  } catch (error) {
    console.error('Error checking blacklist:', error);
    return false;
  }
}

function addToBlacklist(identifier, type, reason) {
  try {
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Blacklist');
    if (!sheet) return;
    
    const now = new Date();
    const expiryDate = new Date(now.getTime() + BLOCK_DURATION_HOURS * 60 * 60 * 1000);
    
    sheet.appendRow([identifier, type, reason, now, 'System', expiryDate]);
  } catch (error) {
    console.error('Error adding to blacklist:', error);
  }
}

function logActivationAttempt(code, deviceInfo, success, error) {
  try {
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('ActivationAttempts');
    if (!sheet) return;
    
    sheet.appendRow([
      new Date(),
      code || 'unknown',
      deviceInfo.ip || 'unknown',
      success,
      error || '',
      deviceInfo.fingerprint || '',
      deviceInfo.userAgent || ''
    ]);
  } catch (error) {
    console.error('Error logging activation attempt:', error);
  }
}

function logAudit(action, code, performer, details) {
  try {
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('AuditLog');
    if (!sheet) return;
    
    sheet.appendRow([
      new Date(),
      action,
      code,
      performer,
      details,
      'Success'
    ]);
  } catch (error) {
    console.error('Error logging audit:', error);
  }
}

// ============================================
// SHEET SETUP AND MANAGEMENT
// ============================================

function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('🎫 ISX License System v3.5')
    .addItem('🔧 Setup All Sheets', 'setupAllSheets')
    .addItem('➕ Add Test License', 'addTestLicense')
    .addItem('🎲 Generate 100 Licenses', 'generate100Licenses')
    .addItem('📊 Show Statistics', 'showStats')
    .addSeparator()
    .addItem('🧪 Test HMAC Signature', 'testHMACSignature')
    .addItem('🔍 Test Fingerprint Similarity', 'testFingerprintSimilarity')
    .addItem('🗑️ Clear All Data', 'clearAllData')
    .addToUi();
}

function setupAllSheets() {
  const spreadsheet = SpreadsheetApp.openById(SHEET_ID);
  
  setupLicensesSheet(spreadsheet);
  setupActivationAttemptsSheet(spreadsheet);
  setupBlacklistSheet(spreadsheet);
  setupAuditLogSheet(spreadsheet);
  
  SpreadsheetApp.getUi().alert(
    '✅ Setup Complete!',
    'All 4 sheets have been set up with correct structure.\n\n' +
    'Next steps:\n' +
    '1. Add test license (ISX License System menu)\n' +
    '2. Deploy as Web App\n' +
    '3. Test activation with Go backend',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

function setupLicensesSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('Licenses');
  if (!sheet) {
    sheet = spreadsheet.insertSheet('Licenses');
  }
  
  sheet.clear();
  
  const headers = [
    'Code',               // A
    'Duration',           // B
    'Status',             // C
    'ActivationDate',     // D
    'ActivationIP',       // E
    'DeviceFingerprint',  // F
    'Email',              // G
    'ExpiryDate',         // H
    'ActivationID',       // I
    'LastChecked',        // J
    'CheckCount',         // K
    'CreatedDate',        // L
    'BatchID',            // M
    'Notes'               // N
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#4285f4');
  headerRange.setFontColor('#ffffff');
  headerRange.setFontWeight('bold');
  headerRange.setHorizontalAlignment('center');
  
  sheet.setColumnWidth(1, 200);
  sheet.setColumnWidth(6, 300);
  sheet.setFrozenRows(1);
}

function setupActivationAttemptsSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('ActivationAttempts');
  if (!sheet) {
    sheet = spreadsheet.insertSheet('ActivationAttempts');
  }
  
  sheet.clear();
  
  const headers = [
    'Timestamp',
    'Code',
    'IP',
    'Success',
    'Error',
    'DeviceFingerprint',
    'UserAgent'
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#ea4335');
  headerRange.setFontColor('#ffffff');
  headerRange.setFontWeight('bold');
  sheet.setFrozenRows(1);
}

function setupBlacklistSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('Blacklist');
  if (!sheet) {
    sheet = spreadsheet.insertSheet('Blacklist');
  }
  
  sheet.clear();
  
  const headers = [
    'Identifier',
    'Type',
    'Reason',
    'AddedDate',
    'AddedBy',
    'ExpiryDate'
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#fbbc04');
  headerRange.setFontColor('#000000');
  headerRange.setFontWeight('bold');
  sheet.setFrozenRows(1);
}

function setupAuditLogSheet(spreadsheet) {
  let sheet = spreadsheet.getSheetByName('AuditLog');
  if (!sheet) {
    sheet = spreadsheet.insertSheet('AuditLog');
  }
  
  sheet.clear();
  
  const headers = [
    'Timestamp',
    'Action',
    'LicenseCode',
    'PerformedBy',
    'Details',
    'Result'
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  const headerRange = sheet.getRange(1, 1, 1, headers.length);
  headerRange.setBackground('#34a853');
  headerRange.setFontColor('#ffffff');
  headerRange.setFontWeight('bold');
  sheet.setFrozenRows(1);
}

function addTestLicense() {
  const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
  
  if (!sheet) {
    SpreadsheetApp.getUi().alert('Error', 'Please run Setup All Sheets first!', SpreadsheetApp.getUi().ButtonSet.OK);
    return;
  }
  
  const data = sheet.getDataRange().getValues();
  for (let i = 1; i < data.length; i++) {
    if (data[i][0] === TEST_LICENSE_CODE) {
      SpreadsheetApp.getUi().alert(
        'Test License Exists',
        `License ${TEST_LICENSE_CODE} already exists.\nStatus: ${data[i][2]}`,
        SpreadsheetApp.getUi().ButtonSet.OK
      );
      return;
    }
  }
  
  const createdDate = new Date();
  sheet.appendRow([
    TEST_LICENSE_CODE,
    '1m',
    'Available',
    '',
    '',
    '',
    '',
    '',
    '',
    '',
    0,
    createdDate,
    'TEST-BATCH',
    'Test license for development'
  ]);
  
  SpreadsheetApp.getUi().alert(
    '✅ Test License Added!',
    `License: ${TEST_LICENSE_CODE}\nDuration: 1 month\nStatus: Available`,
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

function generate100Licenses() {
  const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
  
  if (!sheet) {
    SpreadsheetApp.getUi().alert('Error', 'Please run Setup All Sheets first!', SpreadsheetApp.getUi().ButtonSet.OK);
    return;
  }
  
  const batchId = 'BATCH-' + Utilities.formatDate(new Date(), 'GMT', 'yyyyMMdd-HHmmss');
  const createdDate = new Date();
  
  const existingData = sheet.getDataRange().getValues();
  const existingCodes = new Set();
  for (let i = 1; i < existingData.length; i++) {
    if (existingData[i][0]) existingCodes.add(existingData[i][0]);
  }
  
  const newLicenses = [];
  const chars = '23456789ABCDEFGHJKMNPQRSTUVWXYZ';
  let generated = 0;
  
  while (generated < 100) {
    let code = 'ISX';
    for (let segment = 0; segment < 4; segment++) {
      code += '-';
      for (let i = 0; i < 4; i++) {
        code += chars.charAt(Math.floor(Math.random() * chars.length));
      }
    }
    
    if (!existingCodes.has(code)) {
      newLicenses.push([
        code,
        '1m',
        'Available',
        '',
        '',
        '',
        '',
        '',
        '',
        '',
        0,
        createdDate,
        batchId,
        ''
      ]);
      existingCodes.add(code);
      generated++;
    }
  }
  
  if (newLicenses.length > 0) {
    const lastRow = sheet.getLastRow();
    sheet.getRange(lastRow + 1, 1, newLicenses.length, 14).setValues(newLicenses);
  }
  
  SpreadsheetApp.getUi().alert(
    '✅ Licenses Generated!',
    `Generated: 100 licenses\nBatch ID: ${batchId}\nDuration: 1 month`,
    SpreadsheetApp.getUi().ButtonSet.OK
  );
  
  logAudit('GENERATE_BATCH', batchId, Session.getActiveUser().getEmail(), 'Generated 100 licenses');
}

function showStats() {
  const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName('Licenses');
  
  if (!sheet || sheet.getLastRow() <= 1) {
    SpreadsheetApp.getUi().alert('No Data', 'No licenses found.', SpreadsheetApp.getUi().ButtonSet.OK);
    return;
  }
  
  const data = sheet.getDataRange().getValues();
  
  let stats = {
    total: 0,
    available: 0,
    activated: 0,
    expired: 0,
    revoked: 0
  };
  
  for (let i = 1; i < data.length; i++) {
    if (data[i][0]) {
      stats.total++;
      const status = data[i][2];
      if (status === 'Available') stats.available++;
      else if (status === 'Activated') stats.activated++;
      else if (status === 'Expired') stats.expired++;
      else if (status === 'Revoked') stats.revoked++;
    }
  }
  
  SpreadsheetApp.getUi().alert(
    '📊 License Statistics',
    `Total: ${stats.total}\n\n` +
    `✅ Available: ${stats.available}\n` +
    `🔒 Activated: ${stats.activated}\n` +
    `⏰ Expired: ${stats.expired}\n` +
    `🚫 Revoked: ${stats.revoked}\n\n` +
    `Usage Rate: ${Math.round((stats.activated/stats.total)*100)}%`,
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

function testHMACSignature() {
  // Test HMAC signature generation with correct timestamp format IN SECONDS
  const timestampSec = Math.floor(Date.now() / 1000);  // SECONDS, not milliseconds
  const testResponse = {
    timestamp: timestampSec,  // Unix timestamp in SECONDS
    request_id: 'test_request_123',
    success: true,
    data: {
      status: 'activated',
      license_key: 'ISX-TEST-TEST-TEST-TEST',
      activation_id: 'test_activation_123'
    },
    error: ''
  };
  
  const signature = generateResponseSignature(testResponse);
  
  SpreadsheetApp.getUi().alert(
    '🔐 HMAC Signature Test',
    'Test Response:\n' + JSON.stringify(testResponse, null, 2) + '\n\n' +
    'Generated Signature:\n' + signature + '\n\n' +
    'Timestamp (SECONDS): ' + timestampSec + '\n' +
    'Timestamp type: ' + typeof timestampSec + '\n' +
    'Shared Secret (first 10 chars): ' + SHARED_SECRET.substring(0, 10) + '...',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

function testFingerprintSimilarity() {
  const testCases = [
    {
      fp1: 'abc123def456',
      fp2: 'abc123def456',
      expected: '100%'
    },
    {
      fp1: 'abc123def456',
      fp2: 'abc123def457',
      expected: '~90%'
    },
    {
      fp1: 'abc123def456',
      fp2: 'xyz789uvw012',
      expected: '~0%'
    }
  ];
  
  let results = 'Fingerprint Similarity Tests:\n\n';
  
  testCases.forEach((test, i) => {
    const similarity = calculateFingerprintSimilarity(test.fp1, test.fp2);
    results += `Test ${i + 1}:\n`;
    results += `FP1: ${test.fp1}\n`;
    results += `FP2: ${test.fp2}\n`;
    results += `Expected: ${test.expected}\n`;
    results += `Actual: ${Math.round(similarity * 100)}%\n\n`;
  });
  
  SpreadsheetApp.getUi().alert('🔍 Fingerprint Tests', results, SpreadsheetApp.getUi().ButtonSet.OK);
}

function clearAllData() {
  const response = SpreadsheetApp.getUi().alert(
    '⚠️ Confirm Clear',
    'This will delete ALL data from all sheets. Are you sure?',
    SpreadsheetApp.getUi().ButtonSet.YES_NO
  );
  
  if (response !== SpreadsheetApp.getUi().Button.YES) {
    return;
  }
  
  const sheetNames = ['Licenses', 'ActivationAttempts', 'Blacklist', 'AuditLog'];
  
  sheetNames.forEach(name => {
    const sheet = SpreadsheetApp.openById(SHEET_ID).getSheetByName(name);
    if (sheet && sheet.getLastRow() > 1) {
      sheet.getRange(2, 1, sheet.getLastRow() - 1, sheet.getLastColumn()).clear();
    }
  });
  
  SpreadsheetApp.getUi().alert('✅ Data Cleared', 'All data has been cleared.', SpreadsheetApp.getUi().ButtonSet.OK);
}

// ============================================
// END OF GOOGLE APPS SCRIPT v3.5.0
// ============================================
// This version includes all fixes for validation and action routing
// Properly handles signed requests from Go backend
// ============================================