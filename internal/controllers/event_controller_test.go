package controllers

import (
	"testing"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/controllers/utils/testutil"
	"Sixth_world_Sunday/internal/event"
)

func newEventHarness(t *testing.T) (*testutil.Harness, *event.MockService) {
	h := testutil.NewHarness(t)
	eventMock := event.NewMockService(t)

	s := &Service{
		EventService: eventMock,
		AuthSession:  h.SessionManager,
		AuthzService: h.AuthzService,
	}
	for _, setup := range s.getAllEventRoutes() {
		setup(h.App)
	}

	return h, eventMock
}

func eventFactory(t *testing.T) (*testutil.Harness, *event.MockService) {
	return newEventHarness(t)
}

func TestCreateEvent_PermissionFailures(t *testing.T) {
	testutil.RunPermissionFailureSuite(
		t,
		eventFactory,
		"POST",
		"/events",
		map[string]any{"title": "x"},
		authz.PermManageEvents,
	)
}

func TestEventCover_PermissionFailures(t *testing.T) {
	testutil.RunPermissionFailureSuite(t, eventFactory, "POST", "/events/cover", nil, authz.PermManageEvents)
}

func TestListEvents_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, eventFactory, "GET", "/events", nil)
}
