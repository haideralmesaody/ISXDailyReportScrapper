#!/usr/bin/env python3
"""
Script to create a test completed operation with full pipeline data structure
"""

import requests
import json
import time
import uuid
from datetime import datetime, timedelta

# Base URL
BASE_URL = "http://localhost:8080"

def create_test_pipeline_operation():
    """Create a test full pipeline operation"""

    # Create a full pipeline operation request
    operation_request = {
        "mode": "full",
        "steps": [
            {
                "id": "scraping",
                "type": "scraping",
                "parameters": {
                    "scraping_mode": "full",
                    "test_mode": True
                }
            },
            {
                "id": "processing",
                "type": "processing",
                "dependencies": ["scraping"]
            },
            {
                "id": "indices",
                "type": "indices",
                "dependencies": ["processing"]
            },
            {
                "id": "liquidity",
                "type": "liquidity",
                "dependencies": ["indices"]
            },
            {
                "id": "indicators",
                "type": "indicators",
                "dependencies": ["liquidity"]
            }
        ]
    }

    print("Creating full pipeline operation...")
    response = requests.post(
        f"{BASE_URL}/api/operations/start",
        json=operation_request,
        headers={"Content-Type": "application/json"}
    )

    if response.status_code == 200:
        job_data = response.json()
        print(f"Operation started: {job_data['job_id']}")
        return job_data['job_id']
    else:
        print(f"Failed to start operation: {response.status_code} - {response.text}")
        return None

def create_enhanced_test_operation():
    """Create a mock completed operation with enhanced data structure"""

    # This will create a more realistic test operation with all the fields
    # that the enhanced interface expects

    test_operation = {
        "job_id": f"test-enhanced-{uuid.uuid4().hex[:8]}",
        "operation_id": f"test-enhanced-{uuid.uuid4().hex[:8]}",
        "status": "completed",
        "progress": 100,
        "stage_id": "full_pipeline",
        "stage_name": "Complete ISX Data Pipeline",
        "message": "Job completed successfully",
        "created_at": (datetime.now() - timedelta(minutes=15)).isoformat(),
        "started_at": (datetime.now() - timedelta(minutes=15)).isoformat(),
        "completed_at": (datetime.now() - timedelta(minutes=5)).isoformat(),
        "duration": "10m 15.234s",
        "is_complete": True,
        "metadata": {
            "mode": "full_pipeline",
            "request_id": f"req-{int(time.time())}",
            "steps_count": 5,
            "trace_id": uuid.uuid4().hex
        },
        "stage_results": [
            {
                "stage_id": "scraping",
                "stage_name": "ISX Data Scraping",
                "status": "completed",
                "started_at": (datetime.now() - timedelta(minutes=15)).isoformat(),
                "completed_at": (datetime.now() - timedelta(minutes=12)).isoformat(),
                "duration": "3m 12.456s",
                "progress": 100,
                "message": "Successfully downloaded 5 Excel files from ISX website",
                "output": {
                    "files_downloaded": 5,
                    "total_size_mb": 2.4,
                    "download_path": "/data/downloads/2025-10-24",
                    "files": [
                        "market_summary_2025-10-24.xlsx",
                        "price_list_2025-10-24.xlsx",
                        "sector_data_2025-10-24.xlsx",
                        "trading_volume_2025-10-24.xlsx",
                        "index_constituents_2025-10-24.xlsx"
                    ]
                },
                "performance": {
                    "download_speed_mbps": 1.2,
                    "total_requests": 15,
                    "successful_requests": 15,
                    "failed_requests": 0
                }
            },
            {
                "stage_id": "processing",
                "stage_name": "Data Processing & Validation",
                "status": "completed",
                "started_at": (datetime.now() - timedelta(minutes=12)).isoformat(),
                "completed_at": (datetime.now() - timedelta(minutes=9)).isoformat(),
                "duration": "2m 45.123s",
                "progress": 100,
                "message": "Processed 95 ticker records with validation",
                "output": {
                    "records_processed": 95,
                    "records_valid": 93,
                    "records_invalid": 2,
                    "validation_errors": [
                        {
                            "ticker": "ISX-INVALID1",
                            "error": "Missing price data"
                        },
                        {
                            "ticker": "ISX-INVALID2",
                            "error": "Invalid volume format"
                        }
                    ],
                    "output_files": [
                        "/data/reports/price_data_2025-10-24.csv",
                        "/data/reports/market_data_2025-10-24.csv"
                    ]
                },
                "performance": {
                    "processing_rate_records_per_second": 0.54,
                    "memory_usage_mb": 64.2,
                    "cpu_usage_percent": 15.3
                }
            },
            {
                "stage_id": "indices",
                "stage_name": "ISX60 Index Extraction",
                "status": "completed",
                "started_at": (datetime.now() - timedelta(minutes=9)).isoformat(),
                "completed_at": (datetime.now() - timedelta(minutes=7)).isoformat(),
                "duration": "1m 30.789s",
                "progress": 100,
                "message": "Extracted ISX60 index data successfully",
                "output": {
                    "index_value": 1250.45,
                    "index_change": 15.23,
                    "index_change_percent": 1.23,
                    "constituents_count": 60,
                    "output_file": "/data/reports/indexes/isx60_index_2025-10-24.csv"
                },
                "performance": {
                    "calculation_time_ms": 150,
                    "data_points_processed": 60
                }
            },
            {
                "stage_id": "liquidity",
                "stage_name": "Liquidity Metrics Calculation",
                "status": "completed",
                "started_at": (datetime.now() - timedelta(minutes=7)).isoformat(),
                "completed_at": (datetime.now() - timedelta(minutes=6)).isoformat(),
                "duration": "1m 15.456s",
                "progress": 100,
                "message": "Calculated liquidity metrics for all tickers",
                "output": {
                    "tickers_analyzed": 93,
                    "high_liquidity_tickers": 12,
                    "medium_liquidity_tickers": 45,
                    "low_liquidity_tickers": 36,
                    "output_file": "/data/reports/liquidity/liquidity_metrics_2025-10-24.csv"
                },
                "performance": {
                    "calculations_per_second": 1.24,
                    "memory_usage_mb": 45.7
                }
            },
            {
                "stage_id": "indicators",
                "stage_name": "Technical Indicators Pre-calculation",
                "status": "completed",
                "started_at": (datetime.now() - timedelta(minutes=6)).isoformat(),
                "completed_at": (datetime.now() - timedelta(minutes=5)).isoformat(),
                "duration": "1m 45.678s",
                "progress": 100,
                "message": "Pre-calculated 20+ technical indicators for all tickers",
                "output": {
                    "tickers_processed": 93,
                    "indicators_calculated": 22,
                    "total_data_points": 2046,
                    "output_file": "/data/reports/indicators/ticker_indicators_2025-10-24.csv",
                    "indicators_list": [
                        "sma_20", "sma_50", "sma_200", "ema_20",
                        "rsi_14", "macd_12_26_9", "macd_signal", "macd_histogram",
                        "bollinger_upper", "bollinger_middle", "bollinger_lower", "atr_14",
                        "stoch_k", "stoch_d", "williams_r", "mfi",
                        "adx", "cci", "obv", "vwap",
                        "support_level", "resistance_level"
                    ]
                },
                "performance": {
                    "indicators_per_second": 19.5,
                    "cache_hit_rate_percent": 85.2,
                    "memory_usage_mb": 128.4,
                    "batch_size": 100
                }
            }
        ],
        "data_quality": {
            "completeness_score": 97.8,
            "accuracy_score": 99.1,
            "consistency_score": 98.5,
            "validity_score": 96.3,
            "overall_score": 97.9,
            "issues_found": 2,
            "issues_resolved": 0,
            "quality_checks": {
                "price_data_validation": "passed",
                "volume_data_validation": "passed",
                "date_consistency_check": "passed",
                "ticker_format_validation": "passed",
                "duplicate_detection": "passed"
            }
        },
        "performance": {
            "total_duration_seconds": 615.234,
            "average_stage_duration_seconds": 123.047,
            "peak_memory_usage_mb": 128.4,
            "average_cpu_usage_percent": 18.7,
            "total_data_processed_mb": 12.8,
            "processing_rate_mb_per_second": 0.021,
            "pipeline_efficiency_percent": 94.2
        },
        "business_metrics": {
            "market_data_date": "2025-10-24",
            "tickers_processed": 93,
            "market_cap_coverage_usd": 8500000000,
            "trading_volume_processed": 45000000,
            "sectors_covered": 8,
            "data_freshness_minutes": 15,
            "data_completeness_percent": 97.8,
            "processing_latency_minutes": 10,
            "system_health_score": 94.2,
            "user_value_metrics": {
                "data_points_available": 2046,
                "indicators_ready": True,
                "alerts_generated": 0,
                "reports_available": True,
                "api_response_time_ms": 45
            }
        }
    }

    print("Created enhanced test operation data structure")
    return test_operation

def main():
    print("=== Creating Test Completed Operations ===\n")

    # 1. Create a real pipeline operation
    print("1. Starting real full pipeline operation...")
    job_id = create_test_pipeline_operation()

    if job_id:
        print(f"✅ Real operation started with ID: {job_id}")

        # Wait a bit and check status
        time.sleep(2)
        response = requests.get(f"{BASE_URL}/api/operations/jobs/{job_id}")
        if response.status_code == 200:
            job_status = response.json()
            print(f"   Status: {job_status.get('status', 'unknown')}")
            print(f"   Progress: {job_status.get('progress', 0)}%")
            print(f"   Stage: {job_status.get('stage_name', 'unknown')}")
        else:
            print(f"   Failed to get status: {response.status_code}")
    else:
        print("❌ Failed to create real operation")

    print()

    # 2. Show enhanced test operation structure
    print("2. Creating enhanced test operation structure...")
    enhanced_operation = create_enhanced_test_operation()
    print(f"✅ Enhanced operation created with ID: {enhanced_operation['job_id']}")
    print(f"   Status: {enhanced_operation['status']}")
    print(f"   Stages: {len(enhanced_operation['stage_results'])}")
    print(f"   Duration: {enhanced_operation['duration']}")
    print(f"   Data Quality Score: {enhanced_operation['data_quality']['overall_score']}%")
    print(f"   Business Metrics: {len(enhanced_operation['business_metrics'])} categories")

    print()
    print("=== Test Operations Summary ===")
    print(f"✅ Real operation created: {job_id is not None}")
    print(f"✅ Enhanced operation structure defined: {len(enhanced_operation)} fields")
    print(f"✅ Pipeline stages defined: {len(enhanced_operation['stage_results'])}")

    print("\nTo use the enhanced operation structure for testing:")
    print("1. The enhanced operation shows the expected data structure")
    print("2. Use the real operation ID to test the current interface")
    print("3. Modify the backend to return enhanced data for completed operations")
    print("4. Test the frontend components with the enhanced data")

if __name__ == "__main__":
    main()