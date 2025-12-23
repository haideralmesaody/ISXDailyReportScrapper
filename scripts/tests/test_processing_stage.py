#!/usr/bin/env python3
"""
Test script to verify enhanced processing stage WebSocket telemetry
"""
import asyncio
import websockets
import json
import time
import requests
from datetime import datetime

class ProcessingStageTester:
    def __init__(self, base_url="http://localhost:8080", ws_url="ws://localhost:8080/ws"):
        self.base_url = base_url
        self.ws_url = ws_url
        self.received_messages = []
        self.operation_id = None

    async def connect_websocket(self):
        """Connect to WebSocket and listen for messages"""
        try:
            print(f"Connecting to WebSocket: {self.ws_url}")
            async with websockets.connect(self.ws_url) as websocket:
                print("WebSocket connected successfully")

                # Start listening for messages
                listener_task = asyncio.create_task(self.listen_for_messages(websocket))

                # Wait a moment for connection to establish
                await asyncio.sleep(1)

                # Trigger processing operation
                await self.trigger_processing_operation()

                # Wait for operation to complete (or timeout after 5 minutes)
                await asyncio.sleep(300)

                # Cancel listener task
                listener_task.cancel()

        except Exception as e:
            print(f"❌ WebSocket connection failed: {e}")

    async def listen_for_messages(self, websocket):
        """Listen for WebSocket messages and capture telemetry"""
        try:
            async for message in websocket:
                try:
                    data = json.loads(message)
                    timestamp = datetime.now().isoformat()

                    print(f"\n📨 [{timestamp}] WebSocket Message Received:")
                    print(f"   Type: {data.get('type', 'unknown')}")

                    if data.get('type') == 'operation_update':
                        operation = data.get('operation', {})
                        operation_id = operation.get('id')
                        status = operation.get('status')
                        progress = operation.get('progress', 0)
                        stage = operation.get('current_stage')

                        # Store operation ID for reference
                        if not self.operation_id and operation_id:
                            self.operation_id = operation_id
                            print(f"🎯 Operation ID: {operation_id}")

                        print(f"   Status: {status}")
                        print(f"   Progress: {progress}%")
                        print(f"   Stage: {stage}")

                        # Check for enhanced processing stage features
                        if stage == 'processing' or status == 'running':
                            await self.analyze_processing_stage_data(operation)

                    self.received_messages.append({
                        'timestamp': timestamp,
                        'data': data
                    })

                except json.JSONDecodeError as e:
                    print(f"⚠️  Failed to parse JSON message: {e}")
                    print(f"   Raw message: {message[:200]}...")

        except websockets.exceptions.ConnectionClosed:
            print("🔌 WebSocket connection closed")
        except asyncio.CancelledError:
            print("🛑 WebSocket listener cancelled")
        except Exception as e:
            print(f"❌ WebSocket listener error: {e}")

    async def analyze_processing_stage_data(self, operation):
        """Analyze processing stage data for enhanced features"""
        progress = operation.get('progress', 0)
        stage_progress = operation.get('stage_progress', {})

        print(f"\n🔍 Processing Stage Analysis:")
        print(f"   Overall Progress: {progress}%")

        # Check for file_statuses array
        file_statuses = operation.get('file_statuses', [])
        if file_statuses:
            print(f"   📁 File Statuses ({len(file_statuses)} files):")
            for i, file_status in enumerate(file_statuses[:5]):  # Show first 5
                filename = file_status.get('filename', 'unknown')
                status = file_status.get('status', 'unknown')
                print(f"      {i+1}. {filename} - {status}")
            if len(file_statuses) > 5:
                print(f"      ... and {len(file_statuses) - 5} more files")
        else:
            print("   ⚠️  No file_statuses array found!")

        # Check for total_files count
        total_files = operation.get('total_files')
        if total_files is not None:
            print(f"   📊 Total Files: {total_files}")
        else:
            print("   ⚠️  No total_files count found!")

        # Check for current_phase metadata
        current_phase = operation.get('current_phase')
        if current_phase:
            print(f"   🎯 Current Phase: {current_phase}")
        else:
            print("   ⚠️  No current_phase metadata found!")

        # Check for problematic initial progress of 50
        if len(self.received_messages) <= 2 and progress == 50:
            print("   🚨 ISSUE: Initial progress shows 50% (this should be fixed!)")
        elif len(self.received_messages) <= 2:
            print(f"   ✅ Initial progress is {progress}% (not 50, good!)")

    async def trigger_processing_operation(self):
        """Trigger a processing operation via HTTP API"""
        try:
            print(f"\n🚀 Triggering processing operation...")

            # Check for existing files in downloads directory first
            files_url = f"{self.base_url}/api/v1/operations/downloads/files"
            response = requests.get(files_url)

            if response.status_code == 200:
                files_data = response.json()
                files = files_data.get('files', [])
                print(f"📁 Found {len(files)} files in downloads directory")

                if not files:
                    print("⚠️  No files found in downloads directory. Processing may fail or have no work to do.")
                else:
                    for file in files[:3]:  # Show first 3 files
                        print(f"   - {file}")
            else:
                print(f"⚠️  Could not check downloads directory: {response.status_code}")

            # Trigger the processing operation
            operations_url = f"{self.base_url}/api/v1/operations"
            payload = {
                "operation_type": "processing",
                "config": {}
            }

            response = requests.post(operations_url, json=payload)

            if response.status_code in [200, 201, 202]:
                result = response.json()
                operation_id = result.get('id')
                print(f"✅ Processing operation triggered successfully")
                print(f"   Operation ID: {operation_id}")

                # Store operation ID
                self.operation_id = operation_id

                return operation_id
            else:
                print(f"❌ Failed to trigger processing operation: {response.status_code}")
                print(f"   Response: {response.text}")
                return None

        except Exception as e:
            print(f"❌ Error triggering processing operation: {e}")
            return None

    def print_summary(self):
        """Print summary of captured telemetry"""
        print(f"\n📊 Test Summary:")
        print(f"   Total messages received: {len(self.received_messages)}")
        print(f"   Operation ID: {self.operation_id}")

        # Analyze messages for key features
        processing_messages = [msg for msg in self.received_messages
                             if msg['data'].get('type') == 'operation_update'
                             and msg['data'].get('operation', {}).get('current_stage') == 'processing']

        print(f"   Processing stage messages: {len(processing_messages)}")

        if processing_messages:
            first_msg = processing_messages[0]['data']['operation']

            # Check fixes
            initial_progress = first_msg.get('progress', 0)
            has_file_statuses = 'file_statuses' in first_msg and len(first_msg['file_statuses']) > 0
            has_total_files = 'total_files' in first_msg
            has_current_phase = 'current_phase' in first_msg

            print(f"\n🔧 Enhanced Processing Stage Features:")
            print(f"   ✅ Initial progress not 50: {initial_progress != 50} (was {initial_progress}%)")
            print(f"   ✅ Has file_statuses array: {has_file_statuses} ({len(first_msg.get('file_statuses', []))} files)")
            print(f"   ✅ Has total_files count: {has_total_files} ({first_msg.get('total_files', 'N/A')})")
            print(f"   ✅ Has current_phase metadata: {has_current_phase} ({first_msg.get('current_phase', 'N/A')})")

            # Check progress monotonicity
            progress_values = [msg['data']['operation'].get('progress', 0) for msg in processing_messages]
            is_monotonic = all(progress_values[i] <= progress_values[i+1] for i in range(len(progress_values)-1))
            print(f"   ✅ Progress is monotonic: {is_monotonic}")

            if not is_monotonic:
                print(f"      ⚠️  Progress sequence: {progress_values}")

async def main():
    """Main test function"""
    print("ISX Pulse Enhanced Processing Stage Test")
    print("=" * 50)

    tester = ProcessingStageTester()

    try:
        # Start WebSocket connection and monitoring
        await tester.connect_websocket()

    except KeyboardInterrupt:
        print("\n🛑 Test interrupted by user")
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
    finally:
        # Print summary
        tester.print_summary()

if __name__ == "__main__":
    asyncio.run(main())