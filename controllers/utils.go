package controllers

func mergeMaps(new, old map[string]string) map[string]string {
	if old == nil {
		old = make(map[string]string, len(new))
	}
	for k, v := range new {
		old[k] = v
	}
	return old
}
