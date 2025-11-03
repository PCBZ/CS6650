#!/usr/bin/env python3
"""
Calculate operation-level comparison between MySQL and DynamoDB from combined_results.json
"""

import json
import numpy as np
from typing import Dict, List
from collections import defaultdict

def calculate_operation_metrics(results: List[Dict]) -> Dict:
    """Calculate average response time for each operation type."""
    operation_times = defaultdict(list)
    
    for result in results:
        operation = result['operation']
        response_time = result['response_time']
        operation_times[operation].append(response_time)
    
    operation_metrics = {}
    for operation, times in operation_times.items():
        operation_metrics[operation] = {
            'avg': np.mean(times),
            'count': len(times)
        }
    
    return operation_metrics

def main():
    # Load combined results
    with open('combined_results.json', 'r') as f:
        data = json.load(f)
    
    # Calculate operation metrics for both databases
    mysql_ops = calculate_operation_metrics(data['mysql']['results'])
    dynamo_ops = calculate_operation_metrics(data['dynamodb']['results'])
    
    print("\n" + "="*80)
    print("OPERATION-LEVEL PERFORMANCE COMPARISON: MySQL vs DynamoDB")
    print("="*80)
    
    # Print comparison table
    print("\n{:<20} {:>18} {:>18} {:>18}".format(
        "Operation", "MySQL Avg (ms)", "DynamoDB Avg (ms)", "Faster By"
    ))
    print("-" * 80)
    
    # Get all unique operations
    all_operations = set(mysql_ops.keys()) | set(dynamo_ops.keys())
    
    for operation in sorted(all_operations):
        mysql_avg = mysql_ops.get(operation, {}).get('avg', 0)
        dynamo_avg = dynamo_ops.get(operation, {}).get('avg', 0)
        
        if mysql_avg < dynamo_avg:
            faster_by = f"MySQL ({((dynamo_avg - mysql_avg) / dynamo_avg * 100):.2f}%)"
        elif dynamo_avg < mysql_avg:
            faster_by = f"DynamoDB ({((mysql_avg - dynamo_avg) / mysql_avg * 100):.2f}%)"
        else:
            faster_by = "Tie"
        
        print("{:<20} {:>18.2f} {:>18.2f} {:>18}".format(
            operation.upper(), mysql_avg, dynamo_avg, faster_by
        ))
    
    print("="*80)
    
    # Generate markdown table
    markdown = []
    markdown.append("\n# Operation-Level Performance Comparison\n")
    markdown.append("| Operation | MySQL Avg (ms) | DynamoDB Avg (ms) | Faster By |")
    markdown.append("|-----------|----------------|-------------------|-----------|")
    
    for operation in sorted(all_operations):
        mysql_avg = mysql_ops.get(operation, {}).get('avg', 0)
        dynamo_avg = dynamo_ops.get(operation, {}).get('avg', 0)
        
        if mysql_avg < dynamo_avg:
            faster_by = f"MySQL ({((dynamo_avg - mysql_avg) / dynamo_avg * 100):.2f}%)"
        elif dynamo_avg < mysql_avg:
            faster_by = f"DynamoDB ({((mysql_avg - dynamo_avg) / mysql_avg * 100):.2f}%)"
        else:
            faster_by = "Tie"
        
        markdown.append(f"| {operation.upper()} | {mysql_avg:.2f} | {dynamo_avg:.2f} | {faster_by} |")
    
    # Save markdown
    markdown_output = "\n".join(markdown)
    print(markdown_output)
    
    with open('operation_comparison.md', 'w') as f:
        f.write(markdown_output)
    
    print("\n✅ Markdown table saved to operation_comparison.md\n")
    
    # Save detailed JSON
    comparison_data = {
        'mysql': mysql_ops,
        'dynamodb': dynamo_ops
    }
    
    with open('operation_comparison.json', 'w') as f:
        json.dump(comparison_data, f, indent=2)
    
    print("✅ Detailed data saved to operation_comparison.json\n")

if __name__ == "__main__":
    main()
