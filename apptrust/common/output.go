package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
)

// PrintResponse formats and prints a JSON response body to w.
// Json: pretty-prints the JSON. Table: renders a FIELD/VALUE table using orderedKeys.
// Commands set DefaultFormat: coreformat.Json so the default branch should not be reached;
// it is kept as a no-op safety net.
func PrintResponse(data []byte, outputFormat coreformat.OutputFormat, w io.Writer, orderedKeys []string) error {
	switch outputFormat {
	case coreformat.Json:
		_, err := fmt.Fprintln(w, clientUtils.IndentJson(data))
		return err
	case coreformat.Table:
		return PrintTable(data, w, orderedKeys)
	default:
		return nil
	}
}

// PrintTable renders data as a FIELD/VALUE table using orderedKeys for display order.
// Fields absent from data or with empty/nil values are omitted.
func PrintTable(data []byte, w io.Writer, orderedKeys []string) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "FIELD\tVALUE"); err != nil {
		return err
	}
	for _, key := range orderedKeys {
		val, ok := fields[key]
		if !ok || val == nil {
			continue
		}
		var strVal string
		switch v := val.(type) {
		case string:
			strVal = v
		case []interface{}, map[string]interface{}:
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}
			strVal = string(b)
		default:
			strVal = fmt.Sprintf("%v", v)
		}
		if strVal == "" {
			continue
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", key, strVal); err != nil {
			return err
		}
	}
	return tw.Flush()
}
