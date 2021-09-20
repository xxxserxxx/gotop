package widgets

type Widget interface {
	Update()
}

type Widgets []Widget

func (ws Widgets) Update() {
	for _, wid := range ws {
		wid.Update()
	}
}
