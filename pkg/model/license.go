package model

import (
	"github.com/satori/go.uuid"
	"time"
)

type (
	License struct {
		Model
		ExternalContext     string
		OrderID             string
		ActivationCode      string
		ActivationPolicy    string
		StartDate           time.Time
		EndDate             time.Time
		Package             Package
		PackageID           uuid.UUID
	}

)
