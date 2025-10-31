import json
import matplotlib.pyplot as plt
import os

# 文件路径
RESULT_FILE = "mysql_test_results.json"

# 读取结果
with open(RESULT_FILE, "r") as f:
    data = json.load(f)

results = data["results"]

# 分类收集响应时间
api_types = ["create_cart", "add_items", "get_cart"]
response_times = {api: [] for api in api_types}

for item in results:
    op = item["operation"]
    if op in response_times:
        response_times[op].append(item["response_time"])

# 画图
plt.figure(figsize=(12, 6))
for i, api in enumerate(api_types):
    plt.hist(response_times[api], bins=15, alpha=0.6, label=api)

plt.xlabel("Response Time (ms)")
plt.ylabel("Count")
plt.title("Response Time Distribution for 3 Shopping Cart APIs")
plt.legend()
plt.grid(True, linestyle='--', alpha=0.5)
plt.tight_layout()
plt.savefig("response_distribution.png")
plt.show()

print("图已保存为 response_distribution.png")
