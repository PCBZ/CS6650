# Homework 1
## Part 1: Go Web Service on local and Google Cloud Shell
**Google Platform**  
https://shell.cloud.google.com/?walkthrough_tutorial_url=https%3A%2F%2Fraw.githubusercontent.com%2Fgolang%2Ftour%2Fmaster%2Ftutorial%2Fweb-service-gin.md&show=ide&environment_deployment=ide

**Run the following commands in the terminal**
```bash
curl http://localhost:8080/albums
```
**Get result**
```json
[
    {
        "id": "1",
        "title": "Blue Train",
        "artist": "John Coltrane",
        "price": 56.99
    },
    {
        "id": "2",
        "title": "Jeru",
        "artist": "Gerry Mulligan",
        "price": 17.99
    },
    {
        "id": "3",
        "title": "Sarah Vaughan and Clifford Brown",
        "artist": "Sarah Vaughan",
        "price": 39.99
    }
]
```

## Part 2: Go Web Service on EC2 
curl output in server side
<img width="882" height="376" alt="image" src="https://github.com/user-attachments/assets/63d76022-c37a-4f22-b586-f7dde9a12181" />

## Part 3: Performance Testing and Understanding Response Times

### Result  
<img width="1189" height="790" alt="image" src="https://github.com/user-attachments/assets/23e1ec0b-c7a0-4178-aa13-20c3360c43b9" />

```bash
Statistics:
Total requests: 140
Average response time: 71.30ms
Median response time: 69.61ms
95th percentile: 88.24ms
99th percentile: 107.58ms
Max response time: 126.57ms
```

### Key Observations

**📊 Response Time Distribution:**
- Most requests: 60-75ms (good baseline performance)
- 95th percentile: ~90-95ms
- Occasional spikes: 100-130ms (5% of requests)
- Pattern: Right-skewed distribution with long tail

**⏱️ Consistency:**
- Generally consistent performance throughout test
- Random spikes, no time-based degradation
- Moderate variability (20-25ms gap between median and 95th percentile)

**🔍 Performance Issues:**
- Single EC2 container shows resource constraints
- Spikes likely due to CPU/memory competition or network latency
- Would struggle with 100+ concurrent users

## Part 4: Reading 
What I found most interesting is how the author reduces all the complexity of distributed systems to just two simple facts:

Information has speed limits
Different components fail independently

This approach is brilliant - whether it's consensus algorithms or consistency models, all the complex solutions are really just ways to deal with these two basic physical constraints. Tracing complex problems back to simple root causes - that's what good technical explanation should look like.

