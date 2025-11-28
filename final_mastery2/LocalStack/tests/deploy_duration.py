import matplotlib.pyplot as plt
import numpy as np

# Data from your updated table
users = [20, 40, 60]

# LocalStack data
localstack_50 = [60, 130, 180]
localstack_95 = [100, 210, 310]

# AWS data
aws_50 = [150, 260, 80]
aws_95 = [300, 470, 320]

# Create response time comparison chart
plt.figure(figsize=(10, 6))

# Plot lines
plt.plot(users, localstack_50, 'o-', linewidth=2, markersize=8, 
         label='LocalStack 50th percentile', color='blue')
plt.plot(users, localstack_95, 's--', linewidth=2, markersize=8, 
         label='LocalStack 95th percentile', color='darkblue')
plt.plot(users, aws_50, 'o-', linewidth=2, markersize=8, 
         label='AWS 50th percentile', color='orange')
plt.plot(users, aws_95, 's--', linewidth=2, markersize=8, 
         label='AWS 95th percentile', color='red')

# Labels and title
plt.xlabel('Concurrent Users')
plt.ylabel('Response Time (ms)')
plt.title('Response Time Comparison: LocalStack vs AWS')
plt.legend()
plt.grid(True, alpha=0.3)

# Add value annotations
for i, user in enumerate(users):
    plt.annotate(f'{localstack_50[i]}ms', (user, localstack_50[i]), 
                textcoords="offset points", xytext=(0,10), ha='center')
    plt.annotate(f'{aws_50[i]}ms', (user, aws_50[i]), 
                textcoords="offset points", xytext=(0,-15), ha='center')

plt.tight_layout()
plt.show()

# Create RPS comparison chart
plt.figure(figsize=(10, 6))

# RPS data (updated from your table)
localstack_rps = [320, 310, 310]
aws_rps = [120, 130, 500]

x = np.arange(len(users))
width = 0.35

plt.bar(x - width/2, localstack_rps, width, label='LocalStack', color='blue', alpha=0.7)
plt.bar(x + width/2, aws_rps, width, label='AWS', color='orange', alpha=0.7)

plt.xlabel('Concurrent Users')
plt.ylabel('Requests Per Second (RPS)')
plt.title('RPS Comparison: LocalStack vs AWS')
plt.xticks(x, users)
plt.legend()
plt.grid(True, alpha=0.3)

# Add value labels
for i, v in enumerate(localstack_rps):
    plt.text(i - width/2, v + 10, str(v), ha='center', fontweight='bold')
for i, v in enumerate(aws_rps):
    plt.text(i + width/2, v + 10, str(v), ha='center', fontweight='bold')

plt.tight_layout()
plt.show()

print("Key Findings:")
print("LocalStack: Linear response time increase (60→130→180ms)")
print("AWS: Auto-scaling improves performance at 60 users (260→80ms)")
print("LocalStack RPS: Stable ~310 RPS")
print("AWS RPS: Scales dramatically to 500 RPS at high load")