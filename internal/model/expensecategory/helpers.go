package expensecategory

func buildPK(vaultID string) string {
	return pkPrefix + "::" + vaultID
}
