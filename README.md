# Alert Service (SEMS)

Microservicio de alertas para SEMS. Gestiona umbrales, reglas de inactividad, alertas y notificaciones (email/SMS), y consume eventos de multiples microservicios por Kafka.

## Configuracion centralizada

Este servicio ahora prioriza configuracion desde un Config Service usando:

- `GET /api/v1/config/{service-name}`
- `GET /api/v1/config/services` (fallback)
- `GET /api/v1/config/kafka`

Variable principal:

- `CONFIG_SERVICE_URL` (ej. URL interna del Config Service)

Si el Config Service no devuelve valores, el servicio usa fallback local seguro para no romper compatibilidad.

## Variables requeridas

Mantener en `.env` solo variables sensibles o propias del despliegue:

- `PORT` (recomendado para contenedores/Azure)
- `SERVER_PORT` (compatibilidad legacy local)
- `SERVICE_NAME`
- `CONFIG_SERVICE_URL`
- `DATABASE_URL`
- `KAFKA_ENABLED`
- `KAFKA_BROKERS`
- `KAFKA_CONSUMPTION_TOPICS`
- `KAFKA_ALERTS_TOPIC`
- `ALERT_DEFAULT_STATUS`
- `KAFKA_SECURITY_PROTOCOL`
- `KAFKA_SASL_MECHANISM`
- `KAFKA_USERNAME`
- `KAFKA_PASSWORD`
- `GIN_MODE`
- `TWILIO_ACCOUNT_SID`
- `TWILIO_API_KEY`
- `TWILIO_API_SECRET`
- `TWILIO_PHONE_NUMBER`
- `MAIL_PORT`
- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- `MAIL_FROM`

Opcional para desarrollo local sin Config Service:

- `KAFKA_ENABLED` (`true` para consumir/publicar; `false` para levantar solo HTTP)
- `KAFKA_BROKERS` (ej. `kafka:9092` en Docker local o Event Hubs en Azure)
- `KAFKA_CONSUMER_GROUP`
- `KAFKA_CONSUMPTION_TOPICS`
- `KAFKA_CONSUMPTION_TOPIC`
- `KAFKA_ALERTS_TOPIC`
- `ALERT_DEFAULT_STATUS` (por defecto `open`; el micro normaliza `pending` y `active` a `open`)
- `KAFKA_TOPICS` (topics a autocrear en Docker, separados por comas)
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
3. Si no tienes Config Service, define tambien `KAFKA_ENABLED`, `KAFKA_BROKERS` y opcionalmente `KAFKA_CONSUMER_GROUP`/`KAFKA_CONSUMPTION_TOPICS`. `KAFKA_CONSUMPTION_TOPIC` sigue disponible como fallback legacy de un solo topic.
4. Ejecuta:

```bash
go mod tidy
go run main.go
```

Topics agrupados usados por este microservicio:

- Consume `energy.events`
- Consume `analytics.events`
- Publica `alerts.events` con `eventType=alert.created`

## Docker Compose local

El proyecto incluye `docker-compose.yml` con PostgreSQL y Kafka locales:

```bash
docker compose up --build
```

Al levantar Docker Compose, el servicio `kafka-topics-init` espera a Kafka y crea automaticamente los topics definidos en `KAFKA_TOPICS`.

API local:

```text
http://localhost:8085
```

## Docker

Build:

```bash
docker build -t sems-alert-service:latest .
```

Run (ejemplo local):

```bash
docker run --rm -p 8080:8080 ^
  -e PORT=8080 ^
  -e GIN_MODE=release ^
  -e CONFIG_SERVICE_URL=https://config-service.internal ^
  -e KAFKA_BROKERS=kafka:9092 ^
  -e DATABASE_URL="postgres://USER:PASSWORD@HOST:5432/DB_NAME?sslmode=disable" ^
  sems-alert-service:latest
```

## Azure Container Apps (recomendado)

Para Azure Container Apps, configurar variables de entorno (sin localhost):

- `PORT=8080`
- `SERVICE_NAME=alert-service`
- `GIN_MODE=release`
- `CONFIG_SERVICE_URL` (URL interna/privada del Config Service)
- `KAFKA_ENABLED=true`
- `KAFKA_BROKERS` (broker privado, ej. `broker:9092`)
- `KAFKA_CONSUMPTION_TOPICS=energy.events,analytics.events`
- `KAFKA_ALERTS_TOPIC=alerts.events`
- `ALERT_DEFAULT_STATUS=open`
- `KAFKA_SECURITY_PROTOCOL`
- `KAFKA_SASL_MECHANISM`
- `KAFKA_USERNAME`
- `KAFKA_PASSWORD`
- `DATABASE_URL`

Ejemplo de creacion/actualizacion (referencial):

```bash
az containerapp update \
  --name sems-alert-service \
  --resource-group <RESOURCE_GROUP> \
  --set-env-vars PORT=8080 GIN_MODE=release SERVICE_NAME=alert-service \
  --set-env-vars CONFIG_SERVICE_URL=https://<config-service-interno> \
  --set-env-vars KAFKA_BROKERS=<broker-privado>:9092 \
  --set-env-vars KAFKA_CONSUMPTION_TOPICS=energy.events,analytics.events \
  --set-env-vars KAFKA_ALERTS_TOPIC=alerts.events \
  --set-env-vars KAFKA_SECURITY_PROTOCOL=SASL_SSL \
  --set-env-vars KAFKA_SASL_MECHANISM=PLAIN \
  --set-env-vars KAFKA_USERNAME=<kafka-username> \
  --set-env-vars KAFKA_PASSWORD=<kafka-password> \
  --set-env-vars DATABASE_URL=<database-url>
```

- Guardar secretos en Azure Key Vault o secretos de ACA:
  - `KAFKA_USERNAME`
  - `KAFKA_PASSWORD`
  - `DATABASE_URL`
  - `TWILIO_*`
  - `MAIL_USERNAME`
  - `MAIL_PASSWORD`
- Evitar hardcodear topics/brokers en la app; centralizarlos en Config Service.
- Exponer solo puertos necesarios y usar networking privado entre microservicios cuando sea posible.

## Notas tecnicas

- Las migraciones GORM se ejecutan al iniciar.
- El dominio mantiene independencia de frameworks e infraestructura (DDD).
- El consumidor Kafka se desactiva automaticamente si faltan brokers o topics.
- `Alerts` ahora fuerza el set minimo de topics de su flujo: `energy.events` y `analytics.events`.
- Dentro de esos topics agrupados, el micro detecta por `eventType` al menos `energy.consumption.recorded`, `analytics.anomaly.detected` y `analytics.recommendation.generated`.
- Cuando genera una alerta, publica en `alerts.events` usando `eventType=alert.created`.
