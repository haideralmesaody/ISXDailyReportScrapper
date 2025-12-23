#!/usr/bin/env python3
"""
Test script to trigger enhanced processing stage and capture WebSocket telemetry
"""
import requests
import json
import time
import websockets
import asyncio
from datetime import datetime

# API base URL
BASE_URL = "http://localhost:8080"

async def test_processing_stage():
    print("Testing Enhanced Processing Stage")
    print("=" * 50)

    # 1. Start WebSocket connection to capture telemetry
    ws_url = "ws://localhost:8080/ws"
    print(f"Connecting to WebSocket: {ws_url}")

    try:
        async with websockets.connect(ws_url) as websocket:
            print("WebSocket connected successfully")

            # 2. Check license status first
            print("\nChecking license status...")
            license_response = requests.get(f"{BASE_URL}/api/license/status")
            if license_response.status_code == 200:
                license_data = license_response.json()
                print(f"License Status: {license_data.get('license_status', 'Unknown')}")
                print(f"Expires: {license_data.get('expiry_date', 'Unknown')}")
            else:
                print(f"License check failed: {license_response.status_code}")
                return

            # 3. Trigger processing operation
            print("\nTriggering Processing Operation...")
            operation_payload = {
                "operation": "run_stage",
                "stage_id": "processing",
                "steps": [
                    {
                        "id": "processing",
                        "type": "processing"
                    }
                ]
            }

            print(f"Sending request: {json.dumps(operation_payload, indent=2)}")

            start_response = requests.post(
                f"{BASE_URL}/api/operations/start",
                json=operation_payload,
                headers={"Content-Type": "application/json"}
            )

            if start_response.status_code in [200, 202]:
                operation_data = start_response.json()
                operation_id = operation_data.get('operation_id')
                job_id = operation_data.get('job_id')
                status = operation_data.get('status', 'unknown')
                print(f"Operation queued successfully")
                print(f"Operation ID: {operation_id}")
                print(f"Job ID: {job_id}")
                print(f"Status: {status}")

                # 4. Capture WebSocket telemetry
                print(f"\nCapturing WebSocket telemetry...")
                print("-" * 50)

                telemetry_messages = []
                start_time = time.time()

                while time.time() - start_time < 60:  # Listen for 60 seconds
                    try:
                        message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                        message_data = json.loads(message)

                        timestamp = datetime.now().isoformat()
                        print(f"[{timestamp}] WebSocket Message:")
                        print(json.dumps(message_data, indent=2))
                        print("-" * 50)

                        telemetry_messages.append({
                            'timestamp': timestamp,
                            'message': message_data
                        })

                        # Check if operation is complete
                        if message_data.get('type') == 'operation:complete':
                            print("Operation completed")
                            break
                        elif message_data.get('data', {}).get('status') == 'completed':
                            print("Processing completed")
                            break

                    except asyncio.TimeoutError:
                        print("Waiting for messages...")
                        continue
                    except Exception as e:
                        print(f"WebSocket error: {e}")
                        break

                # 5. Analyze captured telemetry
                print(f"\nTelemetry Analysis")
                print("=" * 50)
                print(f"Total messages captured: {len(telemetry_messages)}")

                # Look for initial progress message
                if telemetry_messages:
                    first_message = telemetry_messages[0]['message']
                    print(f"First message type: {first_message.get('type', 'unknown')}")

                    if first_message.get('type') == 'operation:snapshot':
                        data = first_message.get('data', {})
                        print(f"Initial progress: {data.get('progress', 'unknown')}")
                        print(f"Total files: {data.get('total_files', 'unknown')}")
                        print(f"File statuses: {len(data.get('file_statuses', []))}")
                        print(f"Current phase: {data.get('current_phase', 'unknown')}")
                        print(f"Stage index: {data.get('stage_index', 'unknown')}")

                        # Check for the fixes we implemented
                        if data.get('progress') == 0:
                            print("SUCCESS: Initial progress is 0% (not 50%)")
                        else:
                            print(f"ISSUE: Initial progress is {data.get('progress')}% (expected 0%)")

                        if data.get('total_files'):
                            print(f"SUCCESS: Total files reported: {data.get('total_files')}")
                        else:
                            print("ISSUE: Missing total_files field")

                        if data.get('file_statuses'):
                            print(f"SUCCESS: File statuses provided: {len(data.get('file_statuses'))} files")
                        else:
                            print("ISSUE: Missing file_statuses array")

                        if data.get('stage_index') == 2:
                            print("SUCCESS: Stage index correctly set to 2")
                        else:
                            print(f"ISSUE: Stage index: {data.get('stage_index')} (expected 2)")

                print("\nTest completed successfully!")

            else:
                print(f"Failed to start operation: {start_response.status_code}")
                print(f"Response: {start_response.text}")

    except Exception as e:
        print(f"Test failed: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    try:
        asyncio.run(test_processing_stage())
    except KeyboardInterrupt:
        print("\nTest interrupted by user")
    except Exception as e:
        print(f"Test error: {e}")