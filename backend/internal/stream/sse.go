package stream

import (
	"fmt"
	"net/http"
)

// WriteSSE writes a single Server-Sent Event to the response writer.
// Format: "event: <event>\ndata: <data>\n\n"
// It calls Flush() if the writer implements http.Flusher.
func WriteSSE(w http.ResponseWriter, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
