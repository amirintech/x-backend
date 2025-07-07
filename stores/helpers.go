package stores

import "time"


func toStringPtr(val any) *string {
	if val == nil {
		return nil
	}
	if s, ok := val.(string); ok {
		return &s
	}
	return nil
}

func toTimePtr(val any) *time.Time {
	if val == nil {
		return nil
	}
	if t, ok := val.(time.Time); ok {
		return &t
	}
	return nil
}
