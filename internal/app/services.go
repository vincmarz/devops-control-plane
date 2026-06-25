package app

import "github.com/vincmarz/devops-control-plane/internal/database"

type Services struct {
	Applications *ApplicationService
	Changes      *ChangeService
	Evidence     *EvidenceService
	DB           database.Pinger
}
