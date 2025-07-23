# Quick Test Guide: Progress Tracking Feature (v0.2.0)

## 🚀 5-Minute Quick Test

### Test 1: First Run (No History)
1. **Start App**: Double-click `web-licensed.exe`
2. **Fill Form**:
   - Mode: `Initial (Fresh start)`
   - Browser: `Yes (Headless)`
   - From: 3 days ago
   - To: Today
3. **Click**: `Start Scraping` button
4. **Watch For**:
   - ✅ Progress shows "Calculating..." initially
   - ✅ After first file: Shows time (e.g., "2 minutes")
   - ✅ Progress bar fills smoothly
   - ✅ Pipeline stages light up in order

### Test 2: Second Run (With History)
1. **Restart App**: Close and reopen
2. **Same Settings**: Use exact same dates
3. **Watch For**:
   - ✅ Shows time IMMEDIATELY (e.g., "2 minutes (estimated)")
   - ✅ "(estimated)" label visible
   - ✅ Estimate reasonably close to first run

### Test 3: Error Test
1. **Start Download**: Begin normally
2. **Disconnect Internet**: After 2 files
3. **Check Console**: Look for error message
4. **Reconnect & Retry**: Should resume

---

## 📍 Where to Look

### Progress Information Located:
```
┌─────────────────────────────────────┐
│ ISX Data Collection Form            │
│ [Mode ▼] [Browser ▼]                │
│ [From Date] [To Date]               │
│ [Start Scraping]                    │
├─────────────────────────────────────┤
│ Pipeline: ○━○━○━○                   │ ← Stages light up
├─────────────────────────────────────┤
│ ISX Data Download Progress          │ ← Progress section
│ ████████░░░░░░░ 45%                │    appears here
│ Downloaded: 5  Existing: 2          │
│ Time remaining: 2 minutes           │
│ Status: Downloading file 6 of 12    │
└─────────────────────────────────────┘
```

### Console Output (Bottom):
```
┌─────────────────────────────────────┐
│ Console Output                   🗑  │
├─────────────────────────────────────┤
│ [INIT] Starting scraper...          │
│ [DOWNLOAD] File 1/12: 2024...xlsx   │
│ [SUCCESS] Downloaded in 2.3s        │
│ [ERROR] Failed to download...       │ ← Errors show here
└─────────────────────────────────────┘
```

---

## ✅ Success Indicators

### Good First Run:
- "Calculating..." → "X minutes" transition
- Smooth progress increase
- All stages complete (green checkmarks)

### Good Second Run:
- Immediate time estimate
- "(estimated)" label present
- Close to actual time

### Good Error Handling:
- Clear error in console
- Specific file mentioned
- Helpful recovery hint

---

## ❌ Report These Issues

1. **Progress stuck at "Calculating..."** forever
2. **Time estimates wildly wrong** (>50% off)
3. **"undefined" or "null"** in displays
4. **Progress bar jumps** (0% → 80% instantly)
5. **Stages don't light up** in sequence
6. **No error message** when internet disconnected

---

## 📧 Reporting Results

**Quick Report Format:**
```
Test Date: [Date]
Version: v0.2.0-alpha
Overall: PASS/FAIL

Test 1 (First Run): ✅/❌ [Brief note]
Test 2 (With History): ✅/❌ [Brief note]  
Test 3 (Error): ✅/❌ [Brief note]

Issues Found: [List any problems]
```

**Send to**: [support email]  
**Subject**: "Quick Test - v0.2.0 Progress Tracking"

---

*Thank you for testing! Your feedback helps improve the application.*