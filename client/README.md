# Remnant go client

// To use:
func handleService(w http.ResponseWriter, r *http.Request) {
	cl := NewRemnantClient('http://remnant-server', req)
	defer cl.EndSpan()

	// ... more work ...
	cl.Get('http://another-service/url')
	// ... more work ...
}
