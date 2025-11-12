```mermaid
flowchart TB
    A[Smart Hybrid Fan-out Algorithm] --> B{User Follower Count}
    B -->|< 50K| C[Push Model<br/>Pre-computed Timeline<br/>Fast Reads]
    B -->|≥ 50K| D[Pull Model<br/>On-demand Generation<br/>Fast Writes]
    
    E[85% Regular Users] --> C
    F[15% Influencers/Celebrities] --> D
```