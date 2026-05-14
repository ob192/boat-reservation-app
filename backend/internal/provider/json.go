package provider

import "encoding/json"

// jsonUnmarshal is wrapped so the stub can swap to a different decoder
// later (e.g. with strict mode) without touching call sites.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
