package audit_trail

type ListRequest struct {
	TableName  string `json:"table_name,omitempty"`
	RecordCode string `json:"record_code,omitempty"` // search by entity code (e.g., VEN005)
	Action     string `json:"action,omitempty"`      // CREATE, UPDATE, DELETE
	ChangedBy  string `json:"changed_by,omitempty"`  // user code (e.g., USR001)
	DateFrom   string `json:"date_from,omitempty"`   // YYYY-MM-DD
	DateTo     string `json:"date_to,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

type ListResponse struct {
	Data  []AuditTrail `json:"data"`
	Total int           `json:"total"`
}