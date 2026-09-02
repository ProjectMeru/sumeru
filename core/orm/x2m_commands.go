package orm

import "fmt"

// X2M command tuple op codes (ERP ORM compatibility subset).
const (
	X2MCommandCreate = 0
	X2MCommandUpdate = 1
	X2MCommandDelete = 2
	X2MCommandUnlink = 3
	X2MCommandLink   = 4
	X2MCommandClear  = 5
	X2MCommandSet    = 6
)

// X2MCommand is one relational write command.
type X2MCommand struct {
	Op     int
	ID     int64
	Values map[string]interface{}
	IDs    []int64
}

// ParseX2MCommands normalizes RPC x2m command tuples into structured commands.
func ParseX2MCommands(raw []interface{}) ([]X2MCommand, error) {
	var out []X2MCommand
	for _, item := range raw {
		cmd, err := parseOneX2MCommand(item)
		if err != nil {
			return nil, err
		}
		out = append(out, cmd)
	}
	return out, nil
}

func parseOneX2MCommand(item interface{}) (X2MCommand, error) {
	tuple, ok := item.([]interface{})
	if !ok {
		return X2MCommand{}, fmt.Errorf("x2m: expected command tuple, got %T", item)
	}
	if len(tuple) == 0 {
		return X2MCommand{}, fmt.Errorf("x2m: empty command tuple")
	}
	op, ok := asInt64(tuple[0])
	if !ok {
		return X2MCommand{}, fmt.Errorf("x2m: invalid op code %v", tuple[0])
	}
	cmd := X2MCommand{Op: int(op)}
	switch int(op) {
	case X2MCommandCreate:
		if len(tuple) < 3 {
			return X2MCommand{}, fmt.Errorf("x2m create: need (0, 0, values)")
		}
		cmd.Values = asStringMap(tuple[2])
	case X2MCommandUpdate:
		if len(tuple) < 3 {
			return X2MCommand{}, fmt.Errorf("x2m update: need (1, id, values)")
		}
		cmd.ID = opIntID(tuple[1])
		cmd.Values = asStringMap(tuple[2])
	case X2MCommandDelete, X2MCommandUnlink:
		if len(tuple) < 2 {
			return X2MCommand{}, fmt.Errorf("x2m delete: need (2|3, id)")
		}
		cmd.ID = opIntID(tuple[1])
	case X2MCommandLink:
		if len(tuple) < 2 {
			return X2MCommand{}, fmt.Errorf("x2m link: need (4, id)")
		}
		cmd.ID = opIntID(tuple[1])
	case X2MCommandClear:
		// (5,) — no args
	case X2MCommandSet:
		if len(tuple) < 2 {
			return X2MCommand{}, fmt.Errorf("x2m set: need (6, ids)")
		}
		cmd.IDs = asInt64Slice(tuple[1])
	default:
		return X2MCommand{}, fmt.Errorf("x2m: unsupported op %d", op)
	}
	return cmd, nil
}

func opIntID(v interface{}) int64 {
	n, ok := asInt64(v)
	if !ok {
		return 0
	}
	return n
}

func asInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}

func asInt64Slice(v interface{}) []int64 {
	switch ids := v.(type) {
	case []interface{}:
		out := make([]int64, 0, len(ids))
		for _, id := range ids {
			if n, ok := asInt64(id); ok {
				out = append(out, n)
			}
		}
		return out
	case []int64:
		return ids
	default:
		return nil
	}
}

func asStringMap(v interface{}) map[string]interface{} {
	m, ok := v.(map[string]interface{})
	if ok {
		return m
	}
	return map[string]interface{}{}
}
