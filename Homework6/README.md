# Homework 6
## Search API single node performance
### Baseline 5 users - 2 min
<img width="2928" height="1800" alt="total_requests_per_second_1761078531 749" src="https://github.com/user-attachments/assets/47c16366-b4f8-4a75-b9a2-31d39be18ba8" />

<img width="1079" height="503" alt="image" src="https://github.com/user-attachments/assets/80f9baa4-ac82-4b85-898d-4b112e818ce9" />

### Breaking Point 20 users - 3 min
<img width="2928" height="1800" alt="total_requests_per_second_1761078072 891" src="https://github.com/user-attachments/assets/7e927857-5d3c-44a3-8af6-2d01fdc284f4" />
<img width="1225" height="498" alt="image" src="https://github.com/user-attachments/assets/4f71c219-4540-4352-af60-9221d28b7a2c" />

With load increasing, both CPU and memory utilization increase.
| Metric | 5 Users | 20 Users |
|--------|---------|----------|
| RPS    |  142    |  500  |
| **CPU Utilization** | 14% | 44% |
| **Memory Utilization** | 17% | 10% |
| 50% response time (ms) | 32 | 35 |
| 95% response time (ms) | 50 | 80 |

**Key Observations:**
- Memory usage remains stable
- CPU becomes the primary bottleneck
 
Increase CPU allocation from 256 → 512 CPU units to handle higher load.
<img width="2928" height="1800" alt="total_requests_per_second_1761026159 267" src="https://github.com/user-attachments/assets/05305d1f-58e7-444c-a97f-7074ffb187d9" />

<img width="1263" height="504" alt="image" src="https://github.com/user-attachments/assets/d4d68272-f40e-42f2-aa2b-35c46f6b47c4" />


Switching to 512v CPU and 1G memory and keeping test with 20 users:
| Metric (20 users) | 256v CPU | 512v CPU |
|--------|---------| -------- |
| **CPU Utilization** | 44% | 23% |
| 50% response time (ms) | 35 | 40 |
| 95% response time (ms) | 80 | 80 |

## Horizontal Scaling Infrastructure
### Architecture with ALB
```mermaid
graph TD
    A[Client] --> B[ALB]
    B --> C[Target Group]
    C --> D[ECS Task]
    C --> E[ECS Task]
    C --> F[ECS Task]
    C --> G[ECS Task]
    
    H[Auto Scaling]
    H -.-> D
    H -.-> E
    H -.-> F
    H -.-> G
    
    style B fill:#ff9999
    style C fill:#99ff99
    style H fill:#99ccff
```

### Core Functions
#### ALB
```tf
resource "aws_lb" "this" {
  name               = "${var.service_name}-alb"
  internal           = false
  load_balancer_type = "application"

resource "aws_lb_target_group" "this" {
  ***
  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = var.health_check_path
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }
```

### Auto Scaling
```tf
variable "min_capacity" {
  type        = number
  default     = 2
  description = "Minimum number of ECS tasks"
}

variable "max_capacity" {
  type        = number
  default     = 4
  description = "Maximum number of ECS tasks"
}

variable "target_cpu_utilization" {
  type        = number
  default     = 70
  description = "Target CPU utilization percentage for auto scaling"
}

variable "health_check_path" {
  type        = string
  default     = "/health"
  description = "Health check path for ALB"
}
```


### Applying 20 users
<img width="2928" height="1800" alt="total_requests_per_second_1761116983 614" src="https://github.com/user-attachments/assets/061fe4cb-de63-4ffb-8c94-5b263ffdce68" />
<img width="1295" height="461" alt="image" src="https://github.com/user-attachments/assets/816df708-25b3-49c1-9625-a402aef7c86d" />


### Upgrading to 40 users
<img width="2928" height="1800" alt="total_requests_per_second_1761117397 901" src="https://github.com/user-attachments/assets/c1f2bcd5-07b6-4023-ae93-70c1f5e489ef" />
<img width="1339" height="494" alt="image" src="https://github.com/user-attachments/assets/35cc0aaa-8466-4e35-ab38-9966e58fc712" />


### Upgrading to 120 users
<img width="1348" height="448" alt="image" src="https://github.com/user-attachments/assets/a17808ae-a2d2-4bf9-8863-29c45008eaee" />
<img width="2928" height="1800" alt="total_requests_per_second_1761028248 855" src="https://github.com/user-attachments/assets/40fbb0c9-cff9-445a-80e3-931a1dc25f90" />



#### Performance Test Results Comparison
| Users | Task Count | CPU Usage (%) | Memory Usage (%) | RPS | 50% Response Time (ms) | 95% Response Time (ms) |
|-------|------------|---------------|------------------|-----|------------------------|------------------------|
| 20    | 2          | 60            | 10               | 120 | 150                    | 300                    |
| 40    | 3          | 45            | 10               | 130 | 260                    | 470                    |
| 80    | 4          | 75            | 18               | 34  | 47                     | 340                    |

### Key Observations
**Auto-Scaling is Responsive**  
- CPU-based scaling (70% threshold) triggers appropriately
- New instances come online before system failure
- Load rebalancing happens automatically
**Performance Improves with Scale**
- 95th percentile response times improve: 440ms → 430ms → 340ms
- This demonstrates the power of distributing load across multiple instances

## Resilience Test
<img width="1178" height="331" alt="image" src="https://github.com/user-attachments/assets/d5c3f817-f1ac-40f7-bd4d-8da1f702341d" />
<img width="1157" height="252" alt="image" src="https://github.com/user-attachments/assets/fb4e96e8-c113-4155-8781-c0328cb41b69" />
<img width="1164" height="279" alt="image" src="https://github.com/user-attachments/assets/6ca39f63-46ee-4986-a65b-095172921ed3" />

### Key Observation
The system had self-healing mechanism, auto detecting "unhealthy", then created a new task, maintaining healthy status.

## Exploration
### 90 default CPU utilization
<img width="1162" height="635" alt="image" src="https://github.com/user-attachments/assets/bb79900d-f58d-4786-8982-03a0110ced00" />
<img width="2928" height="1800" alt="total_requests_per_second_1760160011 663" src="https://github.com/user-attachments/assets/16605e56-9ed8-4f24-8401-c62f29d407ec" />

### 50 default CPU utilization
<img width="1161" height="655" alt="image" src="https://github.com/user-attachments/assets/764fd7b5-d6d9-43cb-be9a-a36c104c14bc" />
<img width="2928" height="1800" alt="total_requests_per_second_1760164237 933" src="https://github.com/user-attachments/assets/93a7d698-01a8-4e3e-a9ca-f147bbbab5ac" />

| Description | Task Count | CPU Usage (%) | Memory Usage (%) | RPS | 50% Response Time (ms) | 95% Response Time (ms) |
|-------------|------------|---------------|------------------|-----|------------------------|------------------------|
| 90 CPU      | 3          | 100           | 18               | 30  | 150                    | 980                    |
| 50 CPU      | 6          | 50            | 11               | 35  | 41                     | 200                    |

### Stop multiple instances
<img width="1153" height="635" alt="image" src="https://github.com/user-attachments/assets/52f8362f-5d23-40b3-a46d-5c288c33b0a4" />
<img width="1172" height="456" alt="image" src="https://github.com/user-attachments/assets/1d0aab3e-2bc4-4e29-b5ea-329e58753477" />
<img width="1181" height="623" alt="image" src="https://github.com/user-attachments/assets/38b512ee-1a2b-43e9-a9b9-f77250816eff" />
<img width="2928" height="1800" alt="total_requests_per_second_1760224647 546" src="https://github.com/user-attachments/assets/bd070c04-f3c7-4481-b63f-dc2ed185a81b" />

Stopped 3 instances, then the service recovered to 6 instances. Found a jitter on Locust test and CPU utilization as well.

### Stop all instances
<img width="1166" height="621" alt="image" src="https://github.com/user-attachments/assets/9adad0ad-fbc8-48e8-912f-b858b03a119b" />
<img width="2928" height="1800" alt="total_requests_per_second_1760227106 183" src="https://github.com/user-attachments/assets/b67fd978-c83e-4917-a655-fcc1b9ae43b1" />

Stopped all instances, then the service recoverd to 6 instances. Found some failed requests on Locust tests.

## Result:
### How the system solved your Part II bottleneck
**Part II bottleneck**
- **Resource exhaustion**: CPU/memory limits on single instance
- **No fault tolerance**: Instance failure = whole service down
- **No load distribution**: Single point handles all concurrent requests

**Part III Solution**
- **Redundancy**  
Load balancer provides multiple paths to service
- **Load Distribution**  
ALB distributes requests across healthy instances preventing any single instance from becoming bottleneck
- **Elastic Scaling**  
Auto Scaling automatically adds capacity when needed
CPU-based scaling prevents problems from overload such as increasing request latency
- **Fault Isolation**  
Individual instance failures don't cascade to system failure
Failed instances automatically replaced without manual intervention

### The role of each component
**ALB**: Traffic dispatcher
**Target Group**: Instance manager and health monitor
**Auto Scaling**: Monitor performance and manage instance

### Advantages of Horizontal scalling
1. **Fault Tolerance**
2. **Cost Efficiency**, scalling up and down according to the requirement, does not waste resource.
3. **Load Distribution**, prevants single instance performance bottleneck, such as CPU core, network I/O.









