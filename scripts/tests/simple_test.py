#!/usr/bin/env python3
"""
Simple test script to verify enhanced processing stage WebSocket telemetry
"""
import asyncio
import websockets
import json
import time
import requests
from datetime import datetime

async def test_processing_stage():
    """Test the enhanced processing stage"""
    ws_url = "ws://localhost:8080/ws"
    base_url = "http://localhost:8080"

    messages = []

    try:
        print("Connecting to WebSocket...")
        async with websockets.connect(ws_url) as websocket:
            print("WebSocket connected!")

            # Start listening for messages
            async def listen():
                try:
                    async for message in websocket:
                        data = json.loads(message)
                        timestamp = datetime.now().isoformat()

                        print(f"\n[{timestamp}] Message received:")
                        print(f"Type: {data.get('type', 'unknown')}")
                        print(f"Keys: {list(data.keys())}")

                        if data.get('type') in ['operation_update', 'operation:snapshot']:
                            # Debug: print full message structure for first few messages
                            if len(messages) < 3:
                                print(f"Full data: {json.dumps(data, indent=2)}")

                            operation_data = data.get('data', {})
                            if operation_data:
                                print(f"Operation data keys: {list(operation_data.keys())}")
                                operation_id = operation_data.get('operation_id')
                                status = operation_data.get('status')
                                progress = operation_data.get('progress', 0)

                                # Find processing step
                                processing_step = None
                                steps = operation_data.get('steps', [])
                                for step in steps:
                                    if step.get('id') == 'processing':
                                        processing_step = step
                                        break

                                if processing_step:
                                    step_status = processing_step.get('status')
                                    step_progress = processing_step.get('progress', 0)

                                    print(f"Operation ID: {operation_id}")
                                    print(f"Overall Status: {status}")
                                    print(f"Overall Progress: {progress}%")
                                    print(f"Processing Step Status: {step_status}")
                                    print(f"Processing Step Progress: {step_progress}%")

                                    # Check for enhanced features
                                    if step_status in ['running', 'pending']:
                                        print("\nProcessing Stage Analysis:")

                                        # Check for problematic initial progress of 50
                                        if len(messages) <= 2 and step_progress == 50:
                                            print("ISSUE: Initial processing progress shows 50% (should be fixed!)")
                                        elif len(messages) <= 2:
                                            print(f"GOOD: Initial processing progress is {step_progress}% (not 50)")

                                        # Check for enhanced processing stage features
                                        has_file_statuses = 'file_statuses' in operation_data
                                        has_total_files = 'total_files' in operation_data
                                        has_current_phase = 'current_phase' in operation_data

                                        if has_file_statuses:
                                            file_statuses = operation_data.get('file_statuses', [])
                                            print(f"File statuses: {len(file_statuses)} files")
                                            for i, fs in enumerate(file_statuses[:3]):
                                                print(f"  {i+1}. {fs.get('filename', 'unknown')} - {fs.get('status', 'unknown')}")
                                        else:
                                            print("WARNING: No file_statuses array found!")

                                        if has_total_files:
                                            total_files = operation_data.get('total_files')
                                            print(f"Total files: {total_files}")
                                        else:
                                            print("WARNING: No total_files count found!")

                                        if has_current_phase:
                                            current_phase = operation_data.get('current_phase')
                                            print(f"Current phase: {current_phase}")
                                        else:
                                            print("WARNING: No current_phase metadata found!")

                                    elif step_status == 'completed':
                                        print("\nProcessing Step Completed!")
                                else:
                                    print(f"Operation ID: {operation_id}")
                                    print(f"Status: {status}")
                                    print(f"Progress: {progress}%")
                                    print("Processing step not found in steps array")
                            else:
                                print("No operation data found!")

                        messages.append({'timestamp': timestamp, 'data': data})

                except websockets.exceptions.ConnectionClosed:
                    print("WebSocket connection closed")
                except Exception as e:
                    print(f"WebSocket error: {e}")

            # Start listener
            listener_task = asyncio.create_task(listen())

            # Wait a moment for connection
            await asyncio.sleep(1)

            # Trigger processing operation
            print("\nTriggering processing operation...")
            operations_url = f"{base_url}/api/operations/start"
            payload = {
                "operation": "full_pipeline",
                "stage_id": "processing"
            }

            response = requests.post(operations_url, json=payload)

            if response.status_code in [200, 201, 202]:
                result = response.json()
                operation_id = result.get('id')
                print(f"Processing operation triggered! Operation ID: {operation_id}")
            else:
                print(f"Failed to trigger operation: {response.status_code}")
                print(f"Response: {response.text}")
                return

            # Wait for operation to complete (5 minutes max)
            await asyncio.sleep(300)

            # Cancel listener
            listener_task.cancel()

    except Exception as e:
        print(f"Test failed: {e}")

    # Print summary
    print(f"\nTest Summary:")
    print(f"Total messages: {len(messages)}")

    processing_messages = [msg for msg in messages
                          if msg['data'].get('type') in ['operation_update', 'operation:snapshot']
                          and msg['data'].get('operation', {}).get('current_stage') == 'processing']

    print(f"Processing messages: {len(processing_messages)}")

    if processing_messages:
        first_msg = processing_messages[0]['data']['operation']
        initial_progress = first_msg.get('progress', 0)
        has_file_statuses = 'file_statuses' in first_msg and len(first_msg['file_statuses']) > 0
        has_total_files = 'total_files' in first_msg
        has_current_phase = 'current_phase' in first_msg

        print(f"\nEnhanced Processing Stage Features:")
        print(f"Initial progress not 50: {initial_progress != 50} (was {initial_progress}%)")
        print(f"Has file_statuses array: {has_file_statuses}")
        print(f"Has total_files count: {has_total_files}")
        print(f"Has current_phase metadata: {has_current_phase}")

if __name__ == "__main__":
    asyncio.run(test_processing_stage())