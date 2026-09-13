package security

// FieldPolicy describes read redaction and write denial for a model field.
type FieldPolicy struct {
	ReadRedact         bool
	WriteDenyUnlessSys bool
}

// FieldRegistry is the single source of truth for sensitive field handling.
var FieldRegistry = map[string]map[string]FieldPolicy{
	"core.user": {
		"password":     {ReadRedact: true},
		"totp_secret":  {ReadRedact: true, WriteDenyUnlessSys: true},
		"totp_enabled": {WriteDenyUnlessSys: true},
		"active":       {WriteDenyUnlessSys: true},
		"user_type":    {WriteDenyUnlessSys: true},
		"company_id":   {WriteDenyUnlessSys: true},
	},
	"core.user.apikey": {
		"key_hash": {ReadRedact: true},
	},
	"sys.attachment": {
		"datas": {ReadRedact: true},
	},
}

// ReadRedactFields returns fields stripped on read for a model.
func ReadRedactFields(model string) map[string]bool {
	return fieldsMatching(model, func(p FieldPolicy) bool { return p.ReadRedact })
}

// WriteDenyUnlessSysFields returns fields denied on write for non-system admins.
func WriteDenyUnlessSysFields(model string) map[string]bool {
	return fieldsMatching(model, func(p FieldPolicy) bool { return p.WriteDenyUnlessSys })
}

func fieldsMatching(model string, pred func(FieldPolicy) bool) map[string]bool {
	fields := FieldRegistry[model]
	if len(fields) == 0 {
		return nil
	}
	out := make(map[string]bool)
	for name, policy := range fields {
		if pred(policy) {
			out[name] = true
		}
	}
	return out
}

// ScrubKeywords returns log-scrub keyword fragments derived from the registry.
func ScrubKeywords() []string {
	seen := map[string]bool{
		"sid": true, "password": true, "token": true, "secret": true,
		"authorization": true, "cookie": true, "bearer": true, "session": true,
		"api_key": true, "apikey": true, "key_hash": true, "csrf": true,
		"totp": true, "credential": true, "private_key": true, "refresh": true,
	}
	for _, fields := range FieldRegistry {
		for name, policy := range fields {
			if policy.ReadRedact || policy.WriteDenyUnlessSys {
				seen[name] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	return out
}
