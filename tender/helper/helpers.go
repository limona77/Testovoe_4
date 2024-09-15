package helper

func IsValidTenderStatus(status string) bool {
	validStatuses := map[string]struct{}{"Created": {}, "Published": {}, "Closed": {}}

	if _, ok := validStatuses[status]; !ok {
		return false
	}
	return true
}

func IsValidBidsStatus(status string) bool {
	validStatuses := map[string]struct{}{"Created": {}, "Published": {}, "Canceled": {}}

	if _, ok := validStatuses[status]; !ok {
		return false
	}
	return true
}
