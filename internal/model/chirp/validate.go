package chirp

func ValidateMessage(msg string) (valid bool, error_msg string) {
	if len(msg) == 0 {
		return false, "Chirp not informed"
	}
	if len(msg) > 140 {
		return false, "Chirp is too long"
	}
	return true, ""
}
