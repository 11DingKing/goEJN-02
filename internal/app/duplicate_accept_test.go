package app

import (
	"errors"
	"testing"

	"ejina-microgrid/internal/domain"
)

// TestDuplicateAcceptByAnotherTeamIsAlreadyAccepted checks the error reported
// when a second repair team tries to accept a work order that another team has
// already taken: callers must be able to recognise the business conflict with
// errors.Is instead of having to match on the message text.
func TestDuplicateAcceptByAnotherTeamIsAlreadyAccepted(t *testing.T) {
	svc, _ := testService(t)
	svc.RegisterDieselGen(domain.DieselGenerator{ID: "dg-1", Name: "Gen-1"})

	wo, err := svc.ReportAnomaly(domain.EntityTypeDieselGen, "dg-1", "fuel leak")
	if err != nil {
		t.Fatalf("report anomaly: %v", err)
	}
	if _, err := svc.AcceptWorkOrder(wo.ID, "repair-li"); err != nil {
		t.Fatalf("first accept: %v", err)
	}

	_, err = svc.AcceptWorkOrder(wo.ID, "repair-wang")
	if err == nil {
		t.Fatal("a second accept by a different team must fail")
	}
	if !errors.Is(err, domain.ErrAlreadyAccepted) {
		t.Fatalf("errors.Is(err, domain.ErrAlreadyAccepted) = false, err = %v", err)
	}
}
