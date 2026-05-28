# Alert Service (SEMS)

Microservicio de alertas para SEMS. Gestiona umbrales, reglas de inactividad, alertas y notificaciones (email/SMS), y consume eventos de consumo por Kafka.

## Configuracion centralizada

Este servicio ahora prioriza configuracion desde un Config Service usando:

- `GET /api/v1/config/{service-name}`
- `GET /api/v1/config/services` (fallback)
- `GET /api/v1/config/kafka`

Variable principal:

- `CONFIG_SERVICE_URL` (local: `http://localhost:8090`)

Si el Config Service no devuelve valores, el servicio usa fallback local seguro para no romper compatibilidad.

## Variables locales del microservicio

Mantener en `.env` solo variables sensibles o propias del despliegue:

- `SERVER_PORT`
- `SERVICE_NAME`
- `CONFIG_SERVICE_URL`
- `DATABASE_URL`
- `TWILIO_ACCOUNT_SID`
- `TWILIO_API_KEY`
- `TWILIO_API_SECRET`
- `TWILIO_PHONE_NUMBER`
- `MAIL_PORT`
- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- `MAIL_FROM`

Opcional para desarrollo local sin Config Service:

- `KAFKA_BROKERS` (ej. `localhost:9092`)
- `KAFKA_CONSUMER_GROUP`
- `KAFKA_CONSUMPTION_TOPIC`
- `MAIL_HOST`

## Endpoints

- `GET /api/v1/health`
- `GET /api/v1/alerts`
- `GET /api/v1/alerts/:id`
- `GET /api/v1/users/:userId/alerts`
- `POST /api/v1/alerts`
- `PATCH /api/v1/alerts/:id/status`
- `POST /api/v1/thresholds`
- `GET /api/v1/users/:userId/thresholds`
- `POST /api/v1/inactivity-rules`
- `GET /api/v1/users/:userId/inactivity-rules`
- `POST /api/v1/notification-preferences`
- `GET /api/v1/users/:userId/notification-preferences`
- `POST /api/v1/kafka/publish-test` (publica un evento de prueba en Kafka)

## Ejecucion local

1. Copia `.env.example` a `.env`.
2. Define credenciales reales (DB, Twilio, correo).
3. Si no tienes Config Service local, define tambien `KAFKA_BROKERS` y opcionalmente `KAFKA_CONSUMER_GROUP`/`KAFKA_CONSUMPTION_TOPIC`.
4. Ejecuta:

```bash
go mod tidy
go run main.go
```

## Docker Compose local

El proyecto incluye `docker-compose.yml` con PostgreSQL y Kafka locales:

```bash
docker compose up --build
```

API local:

```text
http://localhost:8086
```

## Azure Container Apps (recomendado)

Para Azure Container Apps:

- Configurar en variables de entorno:
  - `SERVER_PORT`
  - `SERVICE_NAME=alert-service`
  - `CONFIG_SERVICE_URL` (URL interna/privada del Config Service)
- Guardar secretos en Azure Key Vault o secretos de ACA:
  - `DATABASE_URL`
  - `TWILIO_*`
  - `MAIL_USERNAME`
  - `MAIL_PASSWORD`
- Evitar hardcodear topics/brokers en la app; centralizarlos en Config Service.
- Exponer solo puertos necesarios y usar networking privado entre microservicios cuando sea posible.

## Notas tecnicas

- Las migraciones GORM se ejecutan al iniciar.
- El dominio mantiene independencia de frameworks e infraestructura (DDD).
- El consumidor Kafka se desactiva automaticamente si faltan brokers o topic.
