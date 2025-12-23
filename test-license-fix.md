# License Redirect Fix Testing Guide

## Problem Fixed
The application was redirecting to the license page even with a valid license due to:
1. **Race condition** between client-side license check and server-side middleware cache
2. **No retry mechanism** when license status check failed
3. **No forced refresh** of license status to bypass stale cache

## Solutions Implemented

### 1. Enhanced Logging
- Added comprehensive debug logs to root page (`[ROOT PAGE]` prefix)
- Logs license status field value and type
- Logs API response timestamp and full object keys
- Added debug endpoint at `/api/license/debug`

### 2. Improved Error Handling with Retry
- Added retry logic with 2-second delay
- Uses `apiClient.refreshLicenseStatus()` to force cache bypass
- Falls back to license page only after retry fails

### 3. Debug Function
- Added `window.debugLicense()` function to browser console
- Fetches diagnostic info from `/api/license/debug`
- Shows license file info and cache status
- Forces fresh license status check

## Testing Steps

### Step 1: Verify License File
1. Check if `dist/license.dat` exists and is readable
2. Run in browser: `window.debugLicense()`
3. Check console for license file info

### Step 2: Clear Browser Cache
1. Open browser developer tools (F12)
2. Go to Application tab → Storage → Local Storage
3. Clear all items for `localhost` (or your domain)
4. Go to Network tab and check "Disable cache"
5. Refresh the page

### Step 3: Test with Valid License
1. Ensure license is activated through the license page
2. Navigate to root URL: `http://localhost:8080/`
3. Check console for `[ROOT PAGE]` logs
4. Verify it redirects to `/dashboard`

### Step 4: Debugging if Still Issues
If still redirected to license page:

1. **Check Network Tab**:
   - Look for `/api/license/status` request
   - Verify response JSON has `license_status: "active"`
   - Check for any error responses

2. **Run Debug Function**:
   ```javascript
   // In browser console
   window.debugLicense()
   ```

3. **Check Server Logs**:
   - Look for license validation logs
   - Check for cache hit/miss patterns
   - Verify license file is being read

### Step 5: Test Cache Invalidation
1. Activate a new license (or reactivate existing)
2. Check server logs for "Middleware license cache invalidated"
3. Immediately navigate to root URL
4. Should redirect to dashboard, not license page

## Expected Behavior

✅ **With Valid License**:
- Console shows: `[ROOT PAGE] Is license active? true`
- Redirects to `/dashboard` (Market Overview)
- No infinite redirect loops

❌ **With Invalid/No License**:
- Console shows: `[ROOT PAGE] Is license active? false`
- Redirects to `/license` page
- Clear error messages in console

## Troubleshooting

### Issue: Still redirecting to license with valid license
1. **Check the exact status value**:
   - It must be exactly `"active"` or `"warning"`
   - `"Active"` (capital A) will NOT work

2. **Verify middleware cache**:
   - Check server logs for cache hits
   - Look for "License validation cache hit" messages

3. **Check license file format**:
   - Ensure `license.dat` is valid JSON
   - Check `expiry_date` is in the future
   - Verify no corruption in file

4. **Check for multiple tabs**:
   - Close all other browser tabs
   - Each tab makes separate requests which can interfere

### Issue: Console shows errors
1. **Check API endpoint**:
   - Ensure `/api/license/status` returns 200 OK
   - Look for CORS errors in console
   - Check for network connectivity issues

2. **Check browser compatibility**:
   - Test in different browser
   - Disable ad blockers temporarily
   - Try incognito/private mode

## Files Modified

1. **`web/app/page.tsx`**:
   - Enhanced logging
   - Retry logic with `refreshLicenseStatus()`
   - Added debug function

2. **Cache invalidation**:
   - Already implemented in `activation.go` line 328
   - Middleware cache cleared after successful activation

## Support

If issues persist after applying these fixes:
1. Collect console logs
2. Run `window.debugLicense()` in browser
3. Check server logs around the time of redirect
4. Provide all this information in support request