package expense

func buildSK(date, createdAt string) string {
	return date + "::" + createdAt
}