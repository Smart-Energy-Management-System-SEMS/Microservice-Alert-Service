# Alert Service (SEMS)

Microservicio de alertas para el Smart Energy Management System (SEMS). Gestiona umbrales, reglas de inactividad, alertas y notificaciones por email/SMS. Incluye listener Kafka para eventos de consumo.

## Requisitos

- Go 1.22+
- PostgreSQL (Neon u otro)
- Kafka

## Configuracion

1. Copia el archivo .env.example a .env y completa los valores.
2. Asegura que DATABASE_URL apunte a tu instancia de PostgreSQL.
3. Configura credenciales de Twilio y Gmail SMTP.

## Ejecutar

```bash
go mod tidy
go run main.go
```

El servicio expone la API REST en el puerto configurado (SERVER_PORT).

## Endpoints

- GET /api/v1/alerts
- GET /api/v1/alerts/:id
- GET /api/v1/users/:userId/alerts
- POST /api/v1/alerts
- PATCH /api/v1/alerts/:id/status
- POST /api/v1/thresholds
- GET /api/v1/users/:userId/thresholds
- POST /api/v1/inactivity-rules
- GET /api/v1/users/:userId/inactivity-rules
- POST /api/v1/notification-preferences
- GET /api/v1/users/:userId/notification-preferences

## Kafka

Topic de consumo: energy.consumption.recorded

Ejemplo de evento:

```json
{
	"user_id": "uuid",
	"device_id": "uuid",
	"metric": "kwh",
	"value": 15.5,
	"recorded_at": "2026-05-26T00:00:00Z"
}
```

## Notas

- El dominio no depende de Gin, GORM, Kafka, Twilio ni SMTP.
- Las migraciones se ejecutan automaticamente al iniciar el servicio.
- Para notificaciones reales se recomienda integrar un proveedor de contactos de usuario.
- Por defecto, el servicio usa MAIL_FROM o MAIL_USERNAME y TWILIO_PHONE_NUMBER como destinatarios de demo.

## Keep-Alive para Render

Si quieres evitar que el servicio entre en reposo por inactividad, puedes ejecutar un ping periodico contra tu URL publicada.

Scripts incluidos:

- `scripts/keep-alive.sh`
- `scripts/keep-alive.ps1`

Variables:

- `TARGET_URL`: URL base del servicio en Render (ej. `https://tu-servicio.onrender.com`)
- `PING_PATH`: ruta a consultar (default: `/api/v1/alerts`)
- `INTERVAL_SECONDS`: intervalo entre pings en segundos (default: `600`)

Ejemplo Bash:

```bash
TARGET_URL="https://tu-servicio.onrender.com" INTERVAL_SECONDS=600 bash scripts/keep-alive.sh
```

Ejemplo PowerShell:

```powershell
$env:TARGET_URL="https://tu-servicio.onrender.com"
$env:INTERVAL_SECONDS="600"
.\scripts\keep-alive.ps1
```

Importante: en Render, los servicios `Free` pueden suspenderse igual por politicas de la plataforma. Para evitar suspension garantizada, se recomienda plan `paid` o un worker/cron externo que haga los pings.
