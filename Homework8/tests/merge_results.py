#!/usr/bin/env python3
"""
Merge MySQL and DynamoDB test results into a single combined dataset.
This creates combined_results.json as the single source of truth for all analysis.
"""

import json
from pathlib import Path

def merge_test_results():
    """Merge MySQL and DynamoDB test results into a combined dataset."""
    
    # Load both result files
    with open('mysql_test_results.json', 'r') as f:
        mysql_data = json.load(f)
    
    with open('dynamoDB_test_results.json', 'r') as f:
        dynamodb_data = json.load(f)
    
    # Create combined structure
    combined = {
        "mysql": {
            "test_info": mysql_data["test_info"],
            "results": mysql_data["results"]
        },
        "dynamodb": {
            "test_info": dynamodb_data["test_info"],
            "results": dynamodb_data["results"]
        },
        "metadata": {
            "description": "Combined test results for MySQL and DynamoDB implementations",
            "mysql_operations": mysql_data["test_info"]["total_operations"],
            "dynamodb_operations": dynamodb_data["test_info"]["total_operations"],
            "total_operations": mysql_data["test_info"]["total_operations"] + dynamodb_data["test_info"]["total_operations"]
        }
    }
    
    # Write combined results
    with open('combined_results.json', 'w') as f:
        json.dump(combined, f, indent=2)
    
    print("✅ Created combined_results.json")
    print(f"   MySQL operations: {combined['metadata']['mysql_operations']}")
    print(f"   DynamoDB operations: {combined['metadata']['dynamodb_operations']}")
    print(f"   Total operations: {combined['metadata']['total_operations']}")

if __name__ == "__main__":
    merge_test_results()
