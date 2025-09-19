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
The different results show that concurrent programming is dangerous without proper protection. Without proper application may result in unexpected result.

### Collections
<img width="702" height="186" alt="image" src="https://github.com/user-attachments/assets/fec52b47-5ad6-44e8-aba7-adf1aa287f63" />
**Reason**
Maps in Go are not safe for multiple goroutines to write to at the same time.

### Mutex vs. RWMutex vs. SyncMap
<img width="1400" height="800" alt="Performance-Comparison" src="https://github.com/user-attachments/assets/cdea8e0b-16c1-48e8-b921-ea41d592e9b8" />
Mutex and RWMutex performances are similar. The task is more on write operation, while RWMutex take advantages on more read operation scenario. RWMutex will do better on read-heavy tasks.

### File Access
<img width="722" height="62" alt="image" src="https://github.com/user-attachments/assets/61ca5db1-c4b9-44e9-a144-a56e8640bb11" />
Each file.Write ask OS to write to disk, while bufio write data to memory buffer, only write to disk when buffer is full.

### Context Switching
<img width="743" height="113" alt="image" src="https://github.com/user-attachments/assets/0ca2e901-788e-4163-bfae-a90852e7403f" />
Single thread: Go scheduler handles everything internally (like a manager organizing workers in one room)
Multi thread: OS has to get involved (like calling security to move workers between buildings)
Performance comparison: Go routine < Threads < Process < Container < VM
