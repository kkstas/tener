package expense

func buildPK(vaultID string) string {
	return pkPrefix + "::" + vaultID
}

func buildSK(date, createdAt string) string {
	return date + "::" + createdAt
}
