package queries

import "github.com/google/uuid"

type AlertsByUserQuery struct {
	UserID uuid.UUID
}
