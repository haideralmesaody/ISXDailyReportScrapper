# WebSocket Message Framing Specification

## Overview
This document specifies how WebSocket messages must be framed to ensure proper JSON parsing on the frontend.

## The Problem

When multiple JSON messages are sent in a single WebSocket frame (separated by newlines), the frontend's `JSON.parse()` fails because it expects a single valid JSON object:

```javascript
// This fails when event.data contains: {"type":"output"}\n{"type":"status"}
const message = JSON.parse(event.data);
```

## The Solution

Each JSON message must be sent as a separate WebSocket frame. This ensures the frontend receives one complete JSON object per `onmessage` event.

## Implementation Requirements

### Backend (Go)

**Correct Implementation:**
```go
// Send each message as a separate WebSocket frame
func (c *Client) writePump() {
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            // Send message as complete frame
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }
            
            // Send any queued messages as separate frames
            n := len(c.send)
            for i := 0; i < n; i++ {
                select {
                case msg := <-c.send:
                    c.conn.SetWriteDeadline(time.Now().Add(writeWait))
                    if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
                        return
                    }
                default:
                    // Channel was empty
                }
            }
        }
    }
}
```

**Incorrect Implementation (DO NOT USE):**
```go
// DO NOT batch multiple messages in one frame
w, err := c.conn.NextWriter(websocket.TextMessage)
w.Write(message)
for i := 0; i < n; i++ {
    w.Write(newline)  // Adding newlines doesn't help
    w.Write(<-c.send) // Multiple JSON objects in one frame
}
w.Close() // Sends all as one WebSocket message
```

### Frontend (JavaScript)

**Current Implementation (Expects one JSON per frame):**
```javascript
ws.onmessage = function(event) {
    try {
        const message = JSON.parse(event.data);
        handleMessage(message);
    } catch (e) {
        console.error('JSON parse error:', e);
    }
};
```

**Alternative Implementation (If batching was needed):**
```javascript
ws.onmessage = function(event) {
    const messages = event.data.trim().split('\n');
    messages.forEach(msgStr => {
        if (msgStr) {
            try {
                const message = JSON.parse(msgStr);
                handleMessage(message);
            } catch (e) {
                console.error('JSON parse error:', e);
            }
        }
    });
};
```

## Message Format

Each WebSocket frame must contain exactly one JSON object:

```json
{
    "type": "message_type",
    "data": {
        // Message-specific data
    },
    "timestamp": "2024-01-18T10:00:00Z"
}
```

## Common Errors

### 1. JSON Parse Errors
**Symptom**: "Unexpected non-whitespace character after JSON at position X"
**Cause**: Multiple JSON objects in a single WebSocket frame
**Solution**: Send each message as a separate frame

### 2. Debug Output Interference
**Symptom**: Random text appearing in WebSocket messages
**Cause**: `fmt.Printf()` or `fmt.Println()` writing to stdout
**Solution**: Use `log.Printf()` for debug output

### 3. Message Loss
**Symptom**: Some messages never reach the frontend
**Cause**: Messages being batched and only the first is parsed
**Solution**: Proper message framing

## Testing

To verify proper message framing:

1. **Browser DevTools**:
   - Open Network tab → WS
   - Click on WebSocket connection
   - Messages tab should show one JSON object per message

2. **Console Logging**:
   ```javascript
   ws.onmessage = function(event) {
       console.log('Raw message:', event.data);
       console.log('Message length:', event.data.length);
       // Should see complete JSON, no extra characters
   };
   ```

3. **Stress Test**:
   - Send rapid bursts of messages
   - All should parse successfully
   - No JSON errors in console

## Best Practices

1. **Always validate JSON** before sending:
   ```go
   data, err := json.Marshal(message)
   if err != nil {
       log.Printf("JSON marshal error: %v", err)
       return
   }
   ```

2. **Use structured logging** instead of stdout:
   ```go
   // Good
   log.Printf("[DEBUG] Processing: %v", data)
   
   // Bad - interferes with WebSocket
   fmt.Printf("Processing: %v\n", data)
   ```

3. **Set write deadlines** to prevent hanging:
   ```go
   c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
   ```

4. **Handle errors gracefully**:
   ```go
   if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
       log.Printf("WebSocket write error: %v", err)
       return
   }
   ```

## Code References

- WebSocket client: `dev/internal/websocket/client.go:65-98`
- Frontend handler: `dev/web/index.html:1949`
- Hub implementation: `dev/internal/websocket/hub.go`