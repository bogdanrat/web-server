# Microservices POC with gRPC & Event-Driven, running in Docker Containers, allowing CRUD operations on files.
### Backend Services
- **Core Service**: 
  - dispatches incoming requests to **Auth Service** and to **Storage Service**.
  - features a **Pub/Sub** mechanism with **Redis**.
  - includes a Key-Value Store with **Amazon DynamoDB**, internally used by an **i18n** system which reloads key-value pairs when <em>NewKeyValuePairEvent</em> was triggered
  - sends welcome emails when <em>UserSignUpEvent</em> was triggered
- **Auth Service**:
  - authorizes access inside the application 
  - generates **JWT** Access Tokens & Refresh Tokens and **QR Codes** if MFA is enabled.
- **Storage Service**: 
  - responsible for storing files, either in a local filesystem, or to a configured **Amazon S3 Bucket**.
- **Queue Service**:
  - a message broker interface, with **RabbitMQ** and **Amazon SQS** implementations.
  - provides logic for emitting and listening to events (e.g., <em>UserSignUpEvent</em>, <em>NewKeyValuePairEvent</em>)
### Frontend
- **React** minimalistic app
#### Monitoring
- Prometheus
- Grafana
- OpenCensus
#### Reverse Proxy
- NGINX

Services Architecture\
![ServicesArchitecture](https://files.fm/thumb_show.php?i=7bw6pmfsv)
\
AWS Architecture\
![AWSArchitecture](https://files.fm/thumb_show.php?i=bs8xf9xgm)
