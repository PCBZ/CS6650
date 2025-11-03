#!/usr/bin/env python3
"""
Calculate comparison metrics between MySQL and DynamoDB from combined_results.json
"""

import json
import numpy as np
from typing import Dict, List, Tuple

def calculate_percentile(data: List[float], percentile: float) -> float:
    """Calculate the specified percentile of the data."""
    return np.percentile(data, percentile)

def calculate_metrics(results: List[Dict]) -> Dict:
    """Calculate all metrics for a dataset."""
    response_times = [r['response_time'] for r in results]
    successes = [r['success'] for r in results]
    
    metrics = {
        'avg_response_time': np.mean(response_times),
        'p50_response_time': calculate_percentile(response_times, 50),
        'p95_response_time': calculate_percentile(response_times, 95),
        'p99_response_time': calculate_percentile(response_times, 99),
        'success_rate': (sum(successes) / len(successes)) * 100 if successes else 0
    }
    
    return metrics

def determine_winner(mysql_val: float, dynamo_val: float, lower_is_better: bool = True) -> Tuple[str, float]:
    """Determine the winner and calculate the margin."""
    if lower_is_better:
        if mysql_val < dynamo_val:
            margin = ((dynamo_val - mysql_val) / dynamo_val) * 100
            return "MySQL", margin
        else:
            margin = ((mysql_val - dynamo_val) / mysql_val) * 100
            return "DynamoDB", margin
    else:  # higher is better (for success rate)
        if mysql_val > dynamo_val:
            margin = mysql_val - dynamo_val
            return "MySQL", margin
        else:
            margin = dynamo_val - mysql_val
            return "DynamoDB", margin

def main():
    # Load combined results
    with open('combined_results.json', 'r') as f:
        data = json.load(f)
    
    # Calculate metrics for both databases
    mysql_metrics = calculate_metrics(data['mysql']['results'])
    dynamo_metrics = calculate_metrics(data['dynamodb']['results'])
    
    print("\n" + "="*80)
    print("PERFORMANCE COMPARISON: MySQL vs DynamoDB")
    print("="*80)
    
    # Print detailed comparison table
    print("\n{:<30} {:>12} {:>12} {:>12} {:>12}".format(
        "Metric", "MySQL", "DynamoDB", "Winner", "Margin"
    ))
    print("-" * 80)
    
    metrics_to_compare = [
        ('Avg Response Time (ms)', 'avg_response_time', True),
        ('P50 Response Time (ms)', 'p50_response_time', True),
        ('P95 Response Time (ms)', 'p95_response_time', True),
        ('P99 Response Time (ms)', 'p99_response_time', True),
        ('Success Rate (%)', 'success_rate', False),
    ]
    
    for metric_name, metric_key, lower_is_better in metrics_to_compare:
        mysql_val = mysql_metrics[metric_key]
        dynamo_val = dynamo_metrics[metric_key]
        
        winner, margin = determine_winner(mysql_val, dynamo_val, lower_is_better)
        
        if metric_key == 'success_rate':
            print("{:<30} {:>11.2f}% {:>11.2f}% {:>12} {:>10.2f}%".format(
                metric_name, mysql_val, dynamo_val, winner, margin
            ))
        else:
            print("{:<30} {:>11.2f} {:>11.2f} {:>12} {:>10.2f}%".format(
                metric_name, mysql_val, dynamo_val, winner, margin
            ))
    
    print("="*80)
    
    # Summary
    print("\n📊 SUMMARY")
    print("-" * 80)
    
    mysql_wins = sum(1 for _, key, lower in metrics_to_compare 
                     if (mysql_metrics[key] < dynamo_metrics[key] if lower 
                         else mysql_metrics[key] > dynamo_metrics[key]))
    dynamo_wins = len(metrics_to_compare) - mysql_wins
    
    print(f"MySQL wins: {mysql_wins}/{len(metrics_to_compare)} metrics")
    print(f"DynamoDB wins: {dynamo_wins}/{len(metrics_to_compare)} metrics")
    
    # Calculate overall average response time improvement
    avg_improvement = ((mysql_metrics['avg_response_time'] - dynamo_metrics['avg_response_time']) 
                      / mysql_metrics['avg_response_time'] * 100)
    
    if avg_improvement > 0:
        print(f"\n✅ DynamoDB is {abs(avg_improvement):.2f}% faster on average")
    else:
        print(f"\n✅ MySQL is {abs(avg_improvement):.2f}% faster on average")
    
    print("\n" + "="*80 + "\n")
    
    # Save comparison results
    comparison_results = {
        'mysql': mysql_metrics,
        'dynamodb': dynamo_metrics,
        'comparison': {}
    }
    
    for metric_name, metric_key, lower_is_better in metrics_to_compare:
        winner, margin = determine_winner(
            mysql_metrics[metric_key], 
            dynamo_metrics[metric_key], 
            lower_is_better
        )
        comparison_results['comparison'][metric_key] = {
            'mysql': mysql_metrics[metric_key],
            'dynamodb': dynamo_metrics[metric_key],
            'winner': winner,
            'margin_percent': margin
        }
    
    with open('comparison_results.json', 'w') as f:
        json.dump(comparison_results, f, indent=2)
    
    print("✅ Detailed comparison saved to comparison_results.json\n")

if __name__ == "__main__":
    main()
