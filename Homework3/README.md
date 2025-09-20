# Homework3
## Part 1: Time and Ordering
### What I like?
The paper redefines time in distributed systems by separating causality from physical time and introducing the “happening before” relation that captures when one event can causally affect another. The key insight was shifting from “when did it happen?” to “what could have influenced what?” - recognizing that the only meaningful notion of time in a distributed system is the partial order induced by potential causality, not timestamps from imperfectly synchronized clocks.
### What I do not like?
The paper assumes perfect communication and doesn’t address what happens when processes fail or messages are lost, which are critical issues in real distributed systems.
## Part 2: Thread Experiments Overview
### Atomicity
<img width="595" height="108" alt="image" src="https://github.com/user-attachments/assets/cab7039b-7e44-44e4-af4a-e114853fb788" />
<img width="698" height="469" alt="image" src="https://github.com/user-attachments/assets/a232fade-0204-45f9-98c5-39c7980531a8" />

The different results show that concurrent programming is dangerous without proper protection. Without proper application may result in unexpected result. `-race` helps us locate which lines generate data race.

### Collections
<img width="702" height="186" alt="image" src="https://github.com/user-attachments/assets/fec52b47-5ad6-44e8-aba7-adf1aa287f63" />

**Reason**  
Maps in Go are not safe for multiple goroutines to write at the same time.

### Mutex vs. RWMutex vs. SyncMap
<img width="1400" height="800" alt="Performance-Comparison" src="https://github.com/user-attachments/assets/cdea8e0b-16c1-48e8-b921-ea41d592e9b8" />
The figure compares 3 different ways to set value in hashmap data structures, experimenting with 100 times in each test group.

**Mutex** and **RWMutex** performances are similar. The task is write-heavy, while RWMutex take advantages on more read operation scenario. RWMutex will do better on read-heavy tasks.

### File Access
<img width="722" height="62" alt="image" src="https://github.com/user-attachments/assets/61ca5db1-c4b9-44e9-a144-a56e8640bb11" />

Each ```file.Write``` ask OS to write to disk, while bufio write data to memory buffer, only write to disk when buffer is full.

### Context Switching
<img width="743" height="113" alt="image" src="https://github.com/user-attachments/assets/0ca2e901-788e-4163-bfae-a90852e7403f" />

**Single thread**: Go scheduler handles everything internally (like a manager organizing workers in one room)  
**Multi thread**: OS has to get involved (like calling security to move workers between buildings)  
Performance comparison: Go routine < Threads < Process < Container < VM

## Part 3: Making Threads work hard with Load-Testing
### Locust
<img width="1498" height="414" alt="image" src="https://github.com/user-attachments/assets/55693967-e91a-4466-beaf-ff788ec6e497" />

Locust with GET & POST requests.  
### Discussion
In real world scenario, there are more POST requests. Therefore, use HashMap to save searching time and read-write to increase performance.  
In experiments scenario, it uses memory data with more bytes in GET and less bytes in POST. So it costs more time to JSON serialization. 

### Load Test
#### 1 worker
<img width="1496" height="731" alt="image" src="https://github.com/user-attachments/assets/511f1507-0986-4939-b993-e17203d05f3a" />
<img width="924" height="72" alt="image" src="https://github.com/user-attachments/assets/dfb9864d-4cfc-4db8-9d1b-1f055f30d874" />

CPU usage is about 16%

#### 4 workers
<img width="1509" height="764" alt="image" src="https://github.com/user-attachments/assets/44b94cd3-1d11-4127-8c1a-4b2b1b068d7e" />
<img width="918" height="110" alt="image" src="https://github.com/user-attachments/assets/6164ac9a-d099-4ea1-a173-61f0377d2d41" />

CPU usage is about 19%

The 2 Figures above shows 1 worker with 5.9 RPS and 4 worker with 7.8 RPS, according to **Amdahl's Law**:
```latex
Speedup = 1.39x
Speedup = 1 / (S + (1-S)/N)

S ≈ 0.72
```

#### Discussion
The service uses a global slice to store albums. Multiple workers access this slice at the same time. GET requests read and POST requests modify it. When doing ```append()```, Go's runtime detects these data races. It forces operations to run one at a time for safety. This is why 4 workers only gave 1.39x improvement instead of 4x.

### Context Switching
#### 1 worker
<img width="1507" height="799" alt="image" src="https://github.com/user-attachments/assets/83defac5-ffe6-41a2-9b6e-1137c7e7dc9f" />
<img width="909" height="86" alt="image" src="https://github.com/user-attachments/assets/06c4489d-4380-40f0-a080-4642fd76e5bd" />

CPU usage is about 19%

#### 4 worker
<img width="1493" height="797" alt="image" src="https://github.com/user-attachments/assets/2cec36e3-d8a5-4bb9-ac28-205bf2ef4a9e" />
<img width="920" height="121" alt="image" src="https://github.com/user-attachments/assets/b413d3a8-7347-4d16-82c1-500265a3331f" />

CPU usage is about 19%

#### Comparison
| Metric | HttpUser | FastHttpUser |
|--------|----------|--------------|
| **1 worker RPS** | 5.9 | 7.1 |
| **4 workers RPS** | 7.8 | 9.8 |
| **1 worker Docker CPU** | 16%  | 19% |
| **4 workers Docker CPU** | 19%  | 19% |

#### Discussion
`FastHttpUser` can promote RPS performance, while do not increase on CPU usage. It is probably because the bottleneck is the network or server side limit.

