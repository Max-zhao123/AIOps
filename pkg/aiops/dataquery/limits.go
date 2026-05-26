package dataquery

import (
	"encoding/json"
	"errors"
)

const (
	DefaultMaxRows  = 500
	DefaultMaxBytes = 1048576
)

// TruncateRows 限制数组类结果行数（PRD maxRows）。
func TruncateRows(data interface{}, maxRows int) (interface{}, bool) {
	if maxRows <= 0 {
		maxRows = DefaultMaxRows
	}
	m, ok := data.(map[string]interface{})
	if !ok {
		return data, false
	}
	for _, key := range []string{"hits", "items", "data", "results"} {
		arr, ok := m[key].([]interface{})
		if !ok || len(arr) <= maxRows {
			continue
		}
		m[key] = arr[:maxRows]
		m["rowsTruncated"] = true
		return m, true
	}
	return data, false
}

// TruncateResult 强制结果大小上限（PRD dataQuery）。
func TruncateResult(data interface{}, maxBytes int) (interface{}, bool) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return data, false
	}
	if len(raw) <= maxBytes {
		return data, false
	}
	return map[string]interface{}{
		"truncated": true,
		"preview":   string(raw[:maxBytes]),
	}, true
}

// ValidateTimeRange 校验时间范围字符串已在外层处理；此处占位供扩展。
func ValidateTimeRange(start, end string, maxHours int) error {
	if start == "" || end == "" {
		return nil
	}
	_ = maxHours
	return nil
}

var ErrRowsExceeded = errors.New("rows exceeded")
