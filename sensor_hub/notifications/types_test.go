package notifications

import "testing"

func TestNotification_Validate_Valid(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "User added",
		Message:  "User john was added to the system",
	}
	if err := n.Validate(); err != nil {
		t.Errorf("expected valid notification, got error: %v", err)
	}
}

func TestNotification_Validate_InvalidCategory(t *testing.T) {
	n := Notification{
		Category: "invalid_category",
		Severity: SeverityInfo,
		Title:    "Test",
		Message:  "Test message",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestNotification_Validate_EmptyTitle(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "",
		Message:  "Test message",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for empty title")
	}
}

func TestNotification_Validate_EmptyMessage(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "Test",
		Message:  "",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for empty message")
	}
}

func TestNotification_Validate_InvalidSeverity(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: "critical",
		Title:    "Test",
		Message:  "Test message",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for invalid severity")
	}
}

func TestNotification_MetadataJSON(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "Test",
		Message:  "Test",
		Metadata: map[string]interface{}{"user_id": 123},
	}
	data, err := n.MetadataJSON()
	if err != nil {
		t.Fatalf("MetadataJSON failed: %v", err)
	}
	if string(data) != `{"user_id":123}` {
		t.Errorf("unexpected JSON: %s", string(data))
	}
}

func TestNotification_MetadataJSON_Nil(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "Test",
		Message:  "Test",
		Metadata: nil,
	}
	data, err := n.MetadataJSON()
	if err != nil {
		t.Fatalf("MetadataJSON failed: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("expected null, got: %s", string(data))
	}
}

func TestParseMetadataJSON(t *testing.T) {
	data := []byte(`{"user_id":123,"action":"created"}`)
	m, err := ParseMetadataJSON(data)
	if err != nil {
		t.Fatalf("ParseMetadataJSON failed: %v", err)
	}
	if m["user_id"] != float64(123) {
		t.Errorf("expected user_id 123, got %v", m["user_id"])
	}
}

func TestParseMetadataJSON_Null(t *testing.T) {
	m, err := ParseMetadataJSON([]byte("null"))
	if err != nil {
		t.Fatalf("ParseMetadataJSON failed: %v", err)
	}
	if m != nil {
		t.Errorf("expected nil, got %v", m)
	}
}
