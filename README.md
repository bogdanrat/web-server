# A Web App POC, designed as Microservices with gRPC & Event-Driven, running in Docker Containers.
## The app allows CRUD operations on files & images. Image processing (e.g., thumbnail generation) with AWS Lambda could follow.
### Services
- **Core Service**: the first entry point inside the application. Dispatches requests to **Auth Service** and to **Storage Service**. Sends **User SignUp Event** and reacts to it in parallel by sending a welcome email.
  The Message Broker can be configured to be either **RabbitMQ** or **Amazon SQS**.
  It also features a Pub/Sub pattern with Redis.
- **Auth Service**: responsible for authorization inside the application. Generates **JWT** Access Tokens & Refresh Tokens and **QR Codes** if configured.
- **Storage Service**: responsible for storing files, either in a local filesystem, or to a configured **Amazon S3 Bucket**.