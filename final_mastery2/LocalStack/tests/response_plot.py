import matplotlib.pyplot as plt
import numpy as np

# Deployment times in seconds
localstack_time = 83  # 1:23
aws_time = 241  # 4:01

environments = ['LocalStack', 'AWS Production']
times = [localstack_time, aws_time]
colors = ['#2E86AB', '#F18F01']

# Create the plot
plt.figure(figsize=(10, 6))

bars = plt.bar(environments, times, color=colors, alpha=0.8, 
               edgecolor='black', linewidth=1.5, width=0.6)

# Add time labels on bars
for i, (bar, time) in enumerate(zip(bars, times)):
    minutes = time // 60
    seconds = time % 60
    time_label = f"{minutes}:{seconds:02d}"
    plt.text(bar.get_x() + bar.get_width()/2., bar.get_height() + 5,
             time_label, ha='center', va='bottom', 
             fontweight='bold', fontsize=14)

# Customize the plot
plt.ylabel('Deployment Time (seconds)', fontsize=12, fontweight='bold')
plt.title('Deployment Speed Comparison\nLocalStack vs AWS Production', 
          fontsize=14, fontweight='bold', pad=20)
plt.ylim(0, 280)

# Add efficiency annotation
efficiency_gain = (aws_time - localstack_time) / aws_time * 100
plt.text(0.5, 200, f'{efficiency_gain:.0f}% Faster\nDeployment', 
         ha='center', va='center', fontsize=12, fontweight='bold',
         bbox=dict(boxstyle="round,pad=0.5", facecolor='lightgreen', alpha=0.8))

# Add grid for better readability
plt.grid(True, alpha=0.3, axis='y')

# Convert y-axis to minutes:seconds format
y_ticks = plt.yticks()[0]
y_labels = [f"{int(t//60)}:{int(t%60):02d}" for t in y_ticks if t >= 0]
plt.yticks(y_ticks[y_ticks >= 0], y_labels)

plt.tight_layout()
plt.show()

print("Deployment Time Analysis:")
print(f"LocalStack: {localstack_time//60}:{localstack_time%60:02d}")
print(f"AWS: {aws_time//60}:{aws_time%60:02d}")
print(f"LocalStack is {efficiency_gain:.0f}% faster")
print(f"Time saved per deployment: {aws_time - localstack_time} seconds")