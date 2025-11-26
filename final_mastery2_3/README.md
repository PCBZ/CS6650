## Short Project Report / 简短项目报告

This report summarizes the repository, architecture, deployment choices, and the metrics to monitor in each environment. It's designed to be handed to an interviewer and kept under 5 pages.

### Repo information / 仓库信息

- Repository path: `final_mastery2_3` (local)
- Project: Order Processing & Product Search microservice (Go / Gin)
- Key technologies: Go (Gin), Docker, Terraform, LocalStack (Pro), ECS (Fargate model), ALB, CloudWatch (simulated), Locust for load testing

Replace with your public repo URL before sharing: https://github.com/<your-username>/final_mastery2_3

---

## Architecture (diagram)

Below is an architecture diagram (Mermaid). Use this when presenting the system topology.

```mermaid
architecture-beta
  group internet(internet)[Internet]

  group alb(logos:aws-elb)[ALB] in internet
  service alb_instance(logos:aws-elb)[ALB / Listener] in alb

  group vpc(cloud)[VPC]
  group public_subnets(cloud)[Public Subnets] in vpc
  group private_subnets(cloud)[Private Subnets] in vpc

  service nat(gateway)[NAT Gateway] in public_subnets
  service ecs_cluster(server)[ECS Cluster] in private_subnets
  service ecs_service(server)[ECS Service (Fargate)] in ecs_cluster
  service tasks(server)[Task (container)] in ecs_service
  service ecr(database)[ECR Repo] in vpc
  service logs(database)[CloudWatch Logs] in vpc

  alb_instance:R -- L:ecs_service
  ecs_service:B -- T:tasks
  tasks:R -- L:logs
  tasks:L -- R:ecr
  nat:R -- L:ecs_service

``` 

Notes:
- ALB terminates HTTP and forwards to container port 8080.
- ECR stores built Docker images (pushed via Terraform/Docker provider to LocalStack in tests).
- CloudWatch receives container logs and metrics (simulated in LocalStack testing).

---

## Deployment environments and when to use each / 部署环境与适用场景

1) Local (developer machine)
   - Tools: local Docker, `go run`, unit tests, small integration tests.
   - When to use: fast iteration; writing features and unit tests; debugging business logic.
   - Pros: instant feedback, cheap, no infra costs.
   - Cons: not representative of network, IAM, and distributed load.

2) LocalStack (full-stack simulation)
   - Tools: LocalStack (Pro), Terraform configured to point to LocalStack endpoints, Docker images pushed to LocalStack ECR, run ECS tasks simulated by LocalStack.
   - When to use: validate IaC (Terraform) and higher-level integration (ECR/ECS/ALB) without cloud cost; smoke-test deployments and infra changes.
   - Pros: near-real AWS API compatibility, good for CI prechecks, fast iteration on infra code.
   - Cons: slight behavioral differences from real AWS; some services may be simulated or behind feature flags.

3) Staging (real cloud, isolated account)
   - Tools: real AWS account (or sandbox), CI pipeline deploys to staging, end-to-end load tests.
   - When to use: final verification before production, non-destructive load tests, security review.
   - Pros: identical to production environment, enables safe scale testing.
   - Cons: cost and IAM/cleanup overhead.

4) Production
   - Tools: real AWS infrastructure, full monitoring and alerting, autoscaling enabled.
   - When to use: serving real traffic.
   - Pros: reliable, scalable.
   - Cons: requires strict change control and monitoring.

Concrete evidence / examples of use:
- LocalStack: This project uses LocalStack to simulate ECS and ALB for CI and manual validation. Example command used during development:

  - Start LocalStack with required services (iam, ec2, ecs, ecr, elb, logs, application-autoscaling).
  - Run `terraform init && terraform apply -auto-approve` pointed at LocalStack endpoint (http://localhost:4566) to create the simulated ECR, ECS cluster, ALB and autoscaling target/policy.

  Observed result in this repo: terraform plan created 35 resources (VPC, subnets, ALB, ECS, auto-scaling target/policy) and after apply the ALB DNS resolved to `order-processing-service-alb.elb.localhost.localstack.cloud` (LocalStack host mapping).

---

## Metrics and which are meaningful in each environment / 各环境重要指标

Core metrics to collect and why:

- Latency (p50 / p95 / p99)
  - Why: Shows user-perceived response time and tail latency.
  - Where: All environments. In Local use simple latency histograms. In Staging/Prod use real APM or CloudWatch.

- Request rate (RPS)
  - Why: Understand throughput and scaling needs.
  - Where: Staging and Prod (drive with load test in Staging). In LocalStack, use to validate autoscaling triggers.

- Error rate (4xx/5xx)
  - Why: Detect regressions and data/contract issues.
  - Where: All environments.

- CPU / Memory per task / container
  - Why: Determines whether Fargate CPU/memory sizing is adequate and informs autoscaling thresholds.
  - Where: Staging/Prod and LocalStack for autoscaling validation.

- ALB Target group healthy/unhealthy counts
  - Why: Ensure load balancer routing health.
  - Where: Staging/Prod and LocalStack.

- Queue lengths / backlog (if present)
  - Why: Backpressure and worker saturation indicators.
  - Where: All environments if background processing exists.

Suggested charts (examples shown as ASCII samples and the data collection commands):

1) Latency p50/p95/p99 over time (line chart)

   Sample (ASCII):

   p50: ──────▇▇▇▇▇▇▇▇▇▇▇
   p95: ──────▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇
   p99: ──────▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇

   How to get it:
   - Locust/ab/wrk for load tests (staging), or built-in metrics from APM (prod).

2) RPS vs Error rate (dual-axis)

   - Use Locust to generate increase in RPS and observe error spikes to identify breaking points.

3) CPU utilization vs Task count (autoscaling demonstration)

   - For autoscaling verification: run a load test and show CPU increases, hitting the target value (e.g., 70%). The app-autoscaling target increases desired count from min (1) to max (4) as load increases.

Data collection snippets (examples):

  - Locust (simple):

    locust -f tests/load_test.py --headless -u 200 -r 10 --run-time 5m --host=http://<ALB_DNS>

  - Observe CloudWatch (or LocalStack logs) for CPU and memory metrics per task.

---

## How to reproduce quick checks (commands)

1) Start LocalStack (Pro) with services: iam,ec2,ecs,ecr,elb,logs,application-autoscaling

2) In `terraform` directory:

  ```
  terraform init
  terraform apply -auto-approve
  ```

3) Test the app endpoints via ALB (LocalStack uses host mapping):

  - Health: `curl -H "Host: order-processing-service-alb.elb.localhost.localstack.cloud" http://localhost:4566/health`
  - Product search: `curl -H "Host: order-processing-service-alb.elb.localhost.localstack.cloud" "http://localhost:4566/v1/products/search?q=Alpha"`
  - Post order: `curl -X POST -H "Host: order-processing-service-alb.elb.localhost.localstack.cloud" -H "Content-Type: application/json" -d '{"customer_id":12345, "items":[{"product_id":45501, "quantity":2}]}' http://localhost:4566/v1/orders/sync`

4) Run a small Locust load test to validate scaling behavior.

---

## Recommendations & next steps (interview talking points)

- Use LocalStack for fast infra iteration and Terraform validation; always validate in staging before production.
- Instrument the app with structured logs and metrics (OpenTelemetry / Prometheus exporter). Push container-level metrics to CloudWatch in production.
- Add CI gates: `terraform fmt`/`validate`, `go test`, `docker build`, `terraform plan` against LocalStack.

---

If you want, I can add pre-made CSV sample data and simple PNG charts (generated from recorded Locust runs) and embed them into this report. Tell me which target metrics you'd like plotted (latency percentiles, RPS, CPU vs tasks) and I will generate the sample charts.

---

Authorship: Generated and validated during development; include your public repository URL before submitting to Canvas.
