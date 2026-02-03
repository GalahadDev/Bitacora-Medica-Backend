package domains

type AdminDashboardStats struct {
	TotalUsers     int64 `json:"total_users"`
	PendingUsers   int64 `json:"pending_users"`
	ActivePatients int64 `json:"active_patients"`
	TotalSessions  int64 `json:"total_sessions"`
}

type UserGrowthStats struct {
	Month string `json:"name"`
	Count int64  `json:"usuarios"`
}

type ProfessionalDashboardStats struct {
	ActivePatients    int64 `json:"active_patients"`
	MonthlySessions   int64 `json:"monthly_sessions"`
	ReportedIncidents int64 `json:"reported_incidents"`
}

type ActivityStats struct {
	Day   string `json:"name"`
	Count int64  `json:"sesiones"`
}
