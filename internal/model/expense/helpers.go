package expense

func buildPK(vaultID string) string {
	return pkPrefix + "::" + vaultID
}

func buildSK(date, createdAt string) string {
	return date + "::" + createdAt
}

func buildMonthlySumPK(vaultID string) string {
	return monthlySumPKPrefix + "::" + vaultID
}

func buildMonthlySumSK(month, category string) string {
	return month + "::" + category
}
