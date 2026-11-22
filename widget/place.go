package widget

// A layout pass without drawing: the widgets that place children as they
// draw place them here too, so a UI can be laid out before its first frame
// (UI.Layout). For ebitenui/ebitenui#233.

// ChildPlacer is a widget that places its children, as it does before it
// draws them, without drawing; UI.Layout calls it through the tree. A
// game's widget that holds children of its own implements it to take part.
type ChildPlacer interface {
	PlaceChildren()
}

// PlaceChildren places w's children and theirs, if w is a ChildPlacer.
func PlaceChildren(w any) {
	if p, ok := w.(ChildPlacer); ok {
		p.PlaceChildren()
	}
}

// PlaceChildren lays the container's children out, and theirs, as it
// does before it draws them; a hidden container, as one it doesn't draw,
// keeps them where they are.
func (c *Container) PlaceChildren() {
	c.init.Do()
	if !c.widget.IsVisible() {
		return
	}
	c.doLayout()
	for _, ch := range c.children {
		if ch.GetWidget().IsVisible() {
			PlaceChildren(ch)
		}
	}
}

// PlaceChildren places the content where it is scrolled to, and its
// children.
func (s *ScrollContainer) PlaceChildren() {
	s.init.Do()
	if s.content == nil {
		return
	}
	s.placeContent()
	PlaceChildren(s.content)
}

func (f *FlipBook) PlaceChildren() {
	f.init.Do()
	f.container.PlaceChildren()
}

func (t *TabBook) PlaceChildren() {
	t.init.Do()
	t.container.PlaceChildren()
}

func (t *TextArea) PlaceChildren() {
	t.init.Do()
	t.container.PlaceChildren()
}

// PlaceChildren places the list's sliders and rows, and the rows in view's
// entries.
func (l *List) PlaceChildren() {
	l.init.Do()
	l.container.PlaceChildren()
}

func (r *listRows) PlaceChildren() {
	r.layOut()
	for i := r.first; i < min(r.last, len(r.list.entries)); i++ {
		PlaceChildren(r.shownWidget(i))
	}
}

// PlaceChildren sizes a dynamic window to its contents and places them.
func (w *Window) PlaceChildren() {
	w.init.Do()
	w.fitContents()
	PlaceChildren(w.container)
}
