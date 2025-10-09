# Homework 6
## Reading
### What I like
The criteria mentioned can be mapped to the SOLID principles of OOP design: “Information Hidden” -> Encapsulation, “characterized by its knowledge of a design decision which it hides from all others”
-> Single Responsiblity, "Abstract Interfaces" -> Dependency Inversion, "Hierarchical Structure" -> Open/Closed Principle, "Design Decision Isolation" -> Interface Segregation.

It describes the theoretical foundation of microservices architecture. “information hiding” principle maps directly to how microservices encapsulate their data stores and business logic behind API boundaries. Abstract interfaces like are essentially REST/gRPC APIs that provide service contracts without exposing implementation details. The changeability analysis showing how modifications stay isolated to single modules - you can change one service without affecting other services. His emphasis on independent development through abstract interfaces directly enables the DevOps model where different teams can develop, deploy, and scale their services autonomously.

## Search API performance
### Baseline 5 users - 2 min
<img width="2928" height="1800" alt="total_requests_per_second_1759955371 654" src="https://github.com/user-attachments/assets/d900045a-97df-49a4-a3b1-05969ea16978" />

<img width="1118" height="344" alt="image" src="https://github.com/user-attachments/assets/b3dd5ef8-dc85-48b9-8d7b-a57928982f5c" />

### Breaking Point 20 users - 3 min
<img width="2928" height="1800" alt="total_requests_per_second_1759966171 661" src="https://github.com/user-attachments/assets/4f349579-617c-4956-9f97-577c18a55901" />

<img width="1141" height="356" alt="image" src="https://github.com/user-attachments/assets/1f385f69-6d8b-402b-b132-6fbad4a4e885" />

With load increasing, both CPU and memory utilization increase.
| Metric | 5 Users | 20 Users | Change |
|--------|---------|----------|---------|
| **CPU Utilization** | 20% | 80% | +60% ⬆️ |
| **Memory Utilization** | 10% | 11% | +1% ➡️ |

**Evidence for Hardware Scaling (not Code Optimization):**
- Fixed computation workload (exactly 100 product checks per search)
- Memory usage remains stable
- CPU becomes the primary bottleneck
 
Increase CPU allocation from 256 → 512 CPU units to handle higher load.


