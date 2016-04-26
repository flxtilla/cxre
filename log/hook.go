package log

type Hook interface {
	On() []Level
	Fire(Entry) error
}

type Hooks interface {
	AddHook(Hook)
	Fire(Level, Entry) error
}

type hooks struct {
	has map[Level][]Hook
}

func (h *hooks) AddHook(hk Hook) {
	for _, lv := range hk.On() {
		h.has[lv] = append(h.has[lv], hk)
	}
}

func (h *hooks) Fire(lv Level, e Entry) error {
	for _, hook := range h.has[lv] {
		if err := hook.Fire(e); err != nil {
			return err
		}
	}
	return nil
}
