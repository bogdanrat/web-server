# Microservices POC with gRPC & Event-Driven, running in Docker Containers, allowing CRUD operations on files.
### Backend Services
- **Core Service**: 
  - dispatches incoming requests to **Auth Service** and to **Storage Service**.
  - emits events (e.g., <em>UserSignUpEvent</em>, <em>NewKeyValuePairEvent</em>) using a configurable Message Broker, either **RabbitMQ** or **Amazon SQS**.
  - features a **Pub/Sub** mechanism with **Redis**.
  - includes a Key-Value Store with **Amazon DynamoDB**, internally used by an **i18n** system which reloads key-value pairs when <em>NewKeyValuePairEvent</em> was triggered
  - sends welcome emails when <em>UserSignUpEvent</em> was triggered
- **Auth Service**:
  - authorizes access inside the application 
  - generates **JWT** Access Tokens & Refresh Tokens and **QR Codes** if MFA is enabled.
- **Storage Service**: 
  - responsible for storing files, either in a local filesystem, or to a configured **Amazon S3 Bucket**.
  ### Frontend
- **React** minimalistic app
#### Monitoring
- Prometheus
- Grafana
#### Reverse Proxy
- NGINX