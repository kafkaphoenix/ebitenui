package widget

// Scroll containers and text inputs draw through images of their own, a
// MaskedRenderBuffer the size of what they show. They give them back as
// they stop showing: hidden, or a widget holding them hidden, removed from
// its container, or in a window gone from its UI. Shown and drawn again,
// they make them anew. Each widget holding other widgets passes it on.

// bufferHolder is a widget holding images it draws through, or widgets
// that do.
type bufferHolder interface {
	releaseBuffers()
}

// releaseBuffers gives back the images w and the widgets in it draw
// through.
func releaseBuffers(w any) {
	if h, ok := w.(bufferHolder); ok {
		h.releaseBuffers()
	}
}

func (c *Container) releaseBuffers() {
	for _, ch := range c.children {
		releaseBuffers(ch)
	}
}

func (s *ScrollContainer) releaseBuffers() {
	s.renderBuf.Release()
	releaseBuffers(s.content)
}

func (t *TextInput) releaseBuffers() { t.renderBuf.Release() }

func (l *List) releaseBuffers() { releaseBuffers(l.container) }

func (t *TextArea) releaseBuffers() { releaseBuffers(t.container) }

func (t *TabBook) releaseBuffers() { releaseBuffers(t.container) }

func (f *FlipBook) releaseBuffers() { releaseBuffers(f.container) }

func (t *Table) releaseBuffers() { releaseBuffers(t.container) }

func (s *SearchBar) releaseBuffers() { releaseBuffers(s.container) }

func (c *Carousel) releaseBuffers() {
	for _, p := range c.pages {
		releaseBuffers(p)
	}
}

// ReleaseBuffers gives back the images the window's widgets draw through,
// a scroll container's or a text input's; its UI calls it as the window
// goes. A widget drawn again makes its own anew.
func (w *Window) ReleaseBuffers() { releaseBuffers(w.Contents) }
