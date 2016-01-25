package util

func JoinErrors(errs []error) string {
	str := ""
	for i, err := range errs {
		if i > 0 {
			str += ", "
		}
		str += err.Error()
	}
	return str
}
