package template

type Templates interface {
	Templater
	SwapTemplater(Templater)
}

type templates struct {
	Templater
}

func New(t Templater) Templates {
	return &templates{
		Templater: t,
	}
}

func (t *templates) SwapTemplater(tr Templater) {
	t.Templater = tr
}
