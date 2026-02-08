package http

func IsSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
