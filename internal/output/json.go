package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func JSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("cannot encode JSON: %w", err)
	}
	return nil
}
