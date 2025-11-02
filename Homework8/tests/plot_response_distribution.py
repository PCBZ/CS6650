import json
import matplotlib.pyplot as plt

RESULT_FILE = "dynamoDB_test_results.json"

with open(RESULT_FILE, "r") as f:
    data = json.load(f)

results = data["results"]

api_types = ["create_cart", "add_items", "get_cart"]
response_times = {api: [] for api in api_types}

for item in results:
    op = item["operation"]
    if op in response_times:
        response_times[op].append(item["response_time"])

plt.figure(figsize=(12, 6))
for i, api in enumerate(api_types):
    plt.hist(response_times[api], bins=15, alpha=0.6, label=api)

plt.xlabel("Response Time (ms)")
plt.ylabel("Count")
plt.title("Response Time Distribution for 3 Shopping Cart APIs")
plt.legend()
plt.grid(True, linestyle='--', alpha=0.5)
plt.tight_layout()
plt.savefig("dynamodb_response_distribution.png")
plt.show()
