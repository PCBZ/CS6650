# Homework 9 实现改动分析

## 概述
本文档详细说明了从 Homework 8 到 Homework 9 需要进行的改动。

## 主要改动总结

### ✅ 可以保留的部分
1. **Product Service 基础结构** - 但需要修改 API 端点
2. **Shopping Cart Service 基础结构** - 但需要添加 checkout 端点
3. **基础设施代码** - Docker、Go 项目结构等
4. **数据库模型** - 基本数据结构可以复用

### ❌ 需要删除/替换的部分
1. **SNS/SQS 消息队列** → 替换为 **RabbitMQ**
2. **Order Service** → 不再需要（Homework 9 不需要订单服务）
3. **Lambda 处理器** → 不再需要（改用 RabbitMQ 消费者）
4. **当前 ALB 配置** → 需要重构为多 Target Group 配置

### 🆕 需要新增的部分
1. **Credit Card Authorizer Service** - 全新的微服务
2. **Warehouse Service** - RabbitMQ 消费者（Java 实现）
3. **RabbitMQ 基础设施** - Terraform 配置
4. **多 Target Group ALB** - 基于路径和 HTTP Header 的路由
5. **"坏"版本的 Product Service** - 50% 返回 503 错误

---

## 详细改动清单

### 1. Product Service 改动

#### 当前实现 (Homework 8)
- `GET /v1/products/{productId}` ✅ 保留
- `POST /v1/products/{productId}/details` ❌ 需要删除
- `GET /v1/products/search` ❌ 需要删除（Homework 9 不需要）

#### 新要求 (Homework 9)
- `GET /products/{productId}` ✅ 保留（去掉 /v1 前缀）
- `POST /product` ✅ **新增** - 创建产品，服务器生成 product_id

#### 需要修改的文件
- `src/productAPI.go` - 添加 `CreateProduct` 方法，删除 `AddProductDetails` 和 `SearchProducts`
- `src/router.go` - 更新路由配置
- `src/models.go` - 确保 Product 模型符合 OpenAPI 规范

---

### 2. Shopping Cart Service 改动

#### 当前实现 (Homework 8)
- `POST /v1/shopping-carts` ✅ 保留（需要改为 `/shopping-cart`）
- `GET /v1/shopping-carts/:id` ❌ 删除（Homework 9 不需要）
- `POST /v1/shopping-carts/:id/items` ✅ 保留（需要改为 `/shopping-carts/{shoppingCartId}/addItem`）

#### 新要求 (Homework 9)
- `POST /shopping-cart` ✅ 创建购物车
- `POST /shopping-carts/{shoppingCartId}/addItem` ✅ 添加商品
- `POST /shopping-carts/{shoppingCartId}/checkout` ✅ **新增** - 结账功能

#### Checkout 功能要求
1. 接收 `credit_card_number` 参数（格式：`1234-5678-9012-3456`）
2. 调用 Credit Card Authorizer Service 进行授权
3. 如果授权成功，发送消息到 RabbitMQ 队列（发送给 Warehouse）
4. 返回 `order_id`
5. 使用 Manual Acknowledgements 确保 Warehouse 收到消息

#### 需要修改的文件
- `src/shopping_cart_api.go` - 添加 `CheckoutCart` 方法
- `src/router.go` - 更新路由配置，添加 checkout 端点
- 需要添加 RabbitMQ 客户端代码（替换 SNS 发布者）

---

### 3. Credit Card Authorizer Service (全新服务)

#### 功能要求
- `POST /credit-card-authorizer/authorize`
- 验证信用卡号格式：`^[0-9]{4}-[0-9]{4}-[0-9]{4}-[0-9]{4}$`
- 如果格式错误，返回 `400 Bad Request`
- 如果格式正确：
  - 90% 概率返回 `200 OK` (Authorized)
  - 10% 概率返回 `402 Payment Declined`

#### 需要创建的文件
- `src/credit_card_authorizer/main.go` - 新的微服务
- `src/credit_card_authorizer/Dockerfile` - 独立的 Docker 镜像
- Terraform 配置 - 新的 ECS 服务和 Target Group

---

### 4. Warehouse Service (全新服务 - Java)

#### 功能要求
- 作为 RabbitMQ 消费者
- 多线程处理消息
- 使用 Manual Acknowledgements
- 统计每个 product_id 的总数量
- 统计总订单数
- 服务关闭时打印总订单数

#### 需要创建的文件
- `warehouse/src/main/java/...` - Java RabbitMQ 消费者
- `warehouse/Dockerfile` - Java 应用的 Docker 镜像
- Terraform 配置 - 新的 ECS 服务

---

### 5. RabbitMQ 基础设施

#### 需要替换
- 删除 SNS/SQS 相关代码
- 删除 Lambda 处理器
- 添加 RabbitMQ 消息队列

#### 需要创建的文件
- Terraform 配置 - RabbitMQ 实例（可以使用 AWS MQ 或自托管）
- RabbitMQ 连接配置
- 消息格式定义

#### RabbitMQ 配置要求
- 队列名称：需要定义
- 消息格式：需要定义（包含 product_id 和 quantity）
- 使用 Manual Acknowledgements
- 支持多线程消费者

---

### 6. Application Load Balancer 重构

#### 当前实现 (Homework 8)
- 单个 Target Group
- 简单的转发规则

#### 新要求 (Homework 9)
- 多个 Target Groups：
  - ProductService Target Group
  - ShoppingCartService Target Group
  - CreditCardAuthorizerService Target Group
- 基于 URL 路径的路由规则：
  - URL 包含 "product" → ProductService Target Group
  - URL 包含 "shopping-cart" → ShoppingCartService Target Group
  - URL 包含 "credit-card-authorizer" → CreditCardAuthorizerService Target Group
- HTTP Header Condition 路由（在 Target Group 内部）
- 演示负载均衡器管理"坏"实例的能力

#### 需要修改的文件
- `terraform/modules/alb/main.tf` - 重构为多 Target Group 配置
- 添加 Listener Rules 用于路径匹配
- 配置 HTTP Header Condition 路由

---

### 7. "坏"版本的 Product Service

#### 功能要求
- 50% 的请求返回 `503 Service Unavailable`
- 50% 的请求返回 `201 Created`（正常响应）
- 用于演示 ALB 的自动权重调整功能

#### 需要创建的文件
- `src/productAPI_bad.go` - "坏"版本的实现
- 或者使用环境变量控制行为
- 独立的 Docker 镜像或配置

---

### 8. 路由和 API 端点改动

#### 当前路由结构 (Homework 8)
```
/v1/products/{productId}
/v1/products/{productId}/details
/v1/products/search
/v1/shopping-carts
/v1/shopping-carts/:id
/v1/shopping-carts/:id/items
/v1/orders/sync
/v1/orders/async
```

#### 新路由结构 (Homework 9)
```
/products/{productId}          # GET
/product                       # POST
/shopping-cart                 # POST
/shopping-carts/{shoppingCartId}/addItem  # POST
/shopping-carts/{shoppingCartId}/checkout # POST
/credit-card-authorizer/authorize         # POST
```

#### 注意
- 去掉 `/v1` 前缀
- 路由路径必须匹配 OpenAPI 规范
- 每个服务应该是独立的微服务（独立的容器）

---

### 9. 服务拆分

#### 当前架构 (Homework 8)
- 单一服务包含：Product API、Shopping Cart API、Order API

#### 新架构 (Homework 9)
- **Product Service** - 独立的微服务容器
- **Shopping Cart Service** - 独立的微服务容器
- **Credit Card Authorizer Service** - 独立的微服务容器
- **Warehouse Service** - 独立的 Java 容器（RabbitMQ 消费者）
- **RabbitMQ Broker** - 独立的容器

#### 需要修改
- 每个服务需要独立的 `main.go`
- 每个服务需要独立的 `Dockerfile`
- 每个服务需要独立的 ECS 服务配置
- 每个服务需要独立的 Target Group（除了 Warehouse）

---

### 10. 数据库和存储

#### 可以保留
- MySQL/RDS 配置（用于 Shopping Cart）
- DynamoDB 配置（如果使用）

#### 需要注意
- Product Service 可能需要独立的存储
- 确保每个服务的数据隔离

---

### 11. 测试和负载测试

#### 新要求
- 负载测试：发送 200k checkout 消息
- 优化 RabbitMQ 配置：
  - 客户端线程数
  - Warehouse 消费者线程数
  - 队列长度目标：< 1000 消息
- 监控 RabbitMQ 管理控制台
- 测试 ALB 的"坏"实例管理功能

#### 需要创建的文件
- 新的负载测试脚本
- RabbitMQ 性能测试脚本

---

## 实施建议

### 阶段 1: 基础设施准备
1. 创建 RabbitMQ 基础设施（Terraform）
2. 重构 ALB 配置（多 Target Group）
3. 准备服务拆分的基础结构

### 阶段 2: 服务实现
1. 修改 Product Service（添加 POST /product，删除不需要的端点）
2. 修改 Shopping Cart Service（添加 checkout 端点）
3. 创建 Credit Card Authorizer Service
4. 创建 Warehouse Service（Java）

### 阶段 3: 集成测试
1. 测试各个服务的独立功能
2. 测试服务间的通信（Shopping Cart → Credit Card Authorizer）
3. 测试 RabbitMQ 消息传递
4. 测试 ALB 路由规则

### 阶段 4: 负载测试和优化
1. 实现负载测试脚本
2. 优化 RabbitMQ 配置
3. 测试 ALB 的"坏"实例管理
4. 收集性能指标

---

## 关键注意事项

1. **API 路径匹配**：确保所有 API 路径完全匹配 OpenAPI 规范
2. **服务独立性**：每个服务应该是完全独立的，可以单独部署和扩展
3. **错误处理**：确保所有错误响应符合 OpenAPI 规范
4. **消息格式**：定义清晰的 RabbitMQ 消息格式
5. **线程安全**：Warehouse Service 必须处理并发访问
6. **连接池**：RabbitMQ 连接应该使用连接池，避免频繁创建/关闭连接
7. **监控**：设置适当的监控和日志记录

---

## 文件结构建议

```
Homework9/
├── src/
│   ├── product-service/
│   │   ├── main.go
│   │   ├── productAPI.go
│   │   ├── Dockerfile
│   │   └── ...
│   ├── shopping-cart-service/
│   │   ├── main.go
│   │   ├── shopping_cart_api.go
│   │   ├── rabbitmq_client.go
│   │   ├── Dockerfile
│   │   └── ...
│   ├── credit-card-authorizer/
│   │   ├── main.go
│   │   ├── authorizer.go
│   │   ├── Dockerfile
│   │   └── ...
│   └── product-service-bad/  # "坏"版本
│       ├── main.go
│       ├── Dockerfile
│       └── ...
├── warehouse/
│   ├── src/main/java/...
│   ├── Dockerfile
│   └── pom.xml
├── terraform/
│   ├── main.tf
│   ├── modules/
│   │   ├── alb/          # 重构为多 Target Group
│   │   ├── rabbitmq/     # 新增
│   │   ├── product-service/
│   │   ├── shopping-cart-service/
│   │   ├── credit-card-authorizer/
│   │   └── warehouse/
│   └── ...
└── tests/
    ├── load_test.py
    └── ...
```

---

## 总结

Homework 9 需要对 Homework 8 进行**重大重构**：

1. ✅ **可以复用**：基础代码结构、数据库模型、基础设施模块
2. 🔄 **需要修改**：API 端点、路由配置、服务拆分
3. ❌ **需要删除**：SNS/SQS、Lambda、Order Service、Search 功能
4. 🆕 **需要新增**：Credit Card Authorizer、Warehouse、RabbitMQ、多 Target Group ALB

**建议**：这是一个团队作业，建议分工合作，每人负责一个或两个服务的实现和测试。

