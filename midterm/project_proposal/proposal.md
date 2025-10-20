# A Distributed Social Media Platform

## Codebase Description

We propose building a **distributed social media platform** that demonstrates 
key distributed systems concepts from CS6650. This project implements a 
microservices architecture addressing classic scalability challenges including 
data partitioning, replication strategies, and the fan-out problem.

**System Architecture:**
The platform consists of 4 core microservices: User Service (authentication, 
profiles), Post Service (content management), Social Graph Service (relationships), 
and Timeline Service (feed generation). Services communicate via REST APIs with 
asynchronous messaging through AWS SQS for write-heavy operations.

**Technology Stack:**
- Go + Gin framework leveraging team's existing experience
- PostgreSQL for relational data, DynamoDB for timeline storage  
- AWS ECS Fargate + Terraform for cloud-native deployment
- Redis ElastiCache for performance optimization
- SQS/SNS for service decoupling

**Core Functionality:**
Users can register, post content, follow other users, and view personalized 
timelines. The system emphasizes scalability challenges typical of social 
platforms, particularly the fan-out problem when high-follower users post content.

**Distributed Systems Focus:**
The project directly implements CS6650 concepts including CAP theorem trade-offs, 
eventual consistency models, horizontal scaling strategies, and fault tolerance 
patterns. We'll compare different timeline generation approaches (push vs pull vs 
hybrid) to demonstrate performance vs consistency trade-offs in real scenarios.