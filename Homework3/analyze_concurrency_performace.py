import matplotlib.pyplot as plt
import pandas as pd
from scipy import stats
import numpy as np
import os

def read_durations_from_csv(filename: str):
    """Read durations from CSV file"""
    current_dir = os.getcwd()
    print(f"Current working directory: {current_dir}")

    try:
        # Read CSV file
        df = pd.read_csv(filename)
        
        print(f"✓ Successfully read {len(df)} records from {filename}")
        print(f"Columns: {list(df.columns)}")
        print(f"Date range: {df['timestamp'].min()} to {df['timestamp'].max()}")
        
        return df['duration_ms'].tolist()
        
    except FileNotFoundError:
        print(f"✗ File {filename} not found. Please run the Go program first.")
        return []
    except Exception as e:
        print(f"✗ Error reading CSV: {e}")
        return []

def plot_multiple_scatters(csv_files, labels=None):
    """Plot scatter plots for multiple CSV files in one figure"""
    plt.figure(figsize=(14, 8))
    
    colors = ['blue', 'red', 'green']
    markers = ['o', 's', 'x']
    
    for i, filename in enumerate(csv_files):
        durations = read_durations_from_csv(filename)
        
        if durations:  # Only plot if data was successfully read
            sequence_numbers = range(1, len(durations) + 1)
            
            # Use provided label or filename
            label = labels[i] if labels and i < len(labels) else filename.replace('.csv', '')
            
            plt.scatter(sequence_numbers, durations, 
                       color=colors[i % len(colors)], 
                       marker=markers[i % len(markers)],
                       alpha=0.7, 
                       s=50,  # size of dots
                       label=label)
            
    
    # Formatting
    plt.xlabel('Run Number (Sequence)', fontsize=12)
    plt.ylabel('Duration (ms)', fontsize=12)
    plt.title('Performance Comparison', fontsize=14, fontweight='bold')
    plt.grid(True, alpha=0.3)
    plt.legend()
    
    plt.tight_layout()
    plt.show()


if __name__ == "__main__":
    csv_files = ['mutex_durations.csv', 'rwmutex_durations.csv', 'syncmap_durations.csv']
    labels = ['Mutex', 'RWMutex', 'SyncMap']
    plot_multiple_scatters(csv_files, labels)