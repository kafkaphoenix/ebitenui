package input_test

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/input"

	"github.com/matryer/is"
)

func TestSetupInputLayersWithDeferred(t *testing.T) {
	calls := 0
	l := funcLayerer(func(input.DeferredSetupInputLayerFunc) { calls++ })
	input.SetupInputLayersWithDeferred([]input.Layerer{l})
	if calls != 1 {
		t.Errorf("the layerer set its layers up %d times, want once", calls)
	}
}

func TestSetupInputLayersWithDeferred_Deferred(t *testing.T) {
	is := is.New(t)

	called := false
	l := funcLayerer(func(def input.DeferredSetupInputLayerFunc) {
		def(func(_ input.DeferredSetupInputLayerFunc) {
			called = true
		})
	})

	input.SetupInputLayersWithDeferred([]input.Layerer{l})

	is.True(called)
}

func TestLayer_ActiveFor_FullScreen(t *testing.T) {
	is := is.New(t)

	l1 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	input.AddLayer(&l1)

	l2 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	input.AddLayer(&l2)

	is.True(l1.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
}

func TestLayer_ActiveFor_BlockLower(t *testing.T) {
	is := is.New(t)

	l1 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	input.AddLayer(&l1)

	l2 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: true,
		FullScreen: true,
	}
	input.AddLayer(&l2)

	is.True(!l1.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
}

func TestLayer_ActiveFor_Rect(t *testing.T) {
	is := is.New(t)

	l1 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: false,
		RectFunc: func() image.Rectangle {
			return image.Rect(0, 0, 50, 50)
		},
	}
	input.AddLayer(&l1)

	l2 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: false,
		RectFunc: func() image.Rectangle {
			return image.Rect(20, 20, 70, 70)
		},
	}
	input.AddLayer(&l2)

	is.True(l1.ActiveFor(10, 10, input.LayerEventTypeMouseButton))
	is.True(!l2.ActiveFor(10, 10, input.LayerEventTypeMouseButton))

	is.True(l1.ActiveFor(30, 30, input.LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(30, 30, input.LayerEventTypeMouseButton))

	is.True(!l1.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
	is.True(!l2.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
}

func TestLayer_ActiveFor_EventType(t *testing.T) {
	is := is.New(t)

	l1 := input.Layer{
		EventTypes: input.LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	input.AddLayer(&l1)

	l2 := input.Layer{
		EventTypes: input.LayerEventTypeMouseButton,
		BlockLower: false,
		FullScreen: true,
	}
	input.AddLayer(&l2)

	is.True(l1.ActiveFor(100, 100, input.LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(100, 100, input.LayerEventTypeMouseButton))

	is.True(l1.ActiveFor(100, 100, input.LayerEventTypeWheel))
	is.True(!l2.ActiveFor(100, 100, input.LayerEventTypeWheel))
}

// funcLayerer sets its layers up with a function.
type funcLayerer func(def input.DeferredSetupInputLayerFunc)

func (f funcLayerer) SetupInputLayer(def input.DeferredSetupInputLayerFunc) { f(def) }

// What the layerers defer runs after every layerer has set its own layers
// up, in the order deferred, so a layer deferred lies above them all.
func TestSetupInputLayersWithDeferred_DeferredAfterAll(t *testing.T) {
	is := is.New(t)

	var order []string
	layerer := func(name string) input.Layerer {
		return funcLayerer(func(def input.DeferredSetupInputLayerFunc) {
			order = append(order, name)
			def(func(input.DeferredSetupInputLayerFunc) { order = append(order, name+" deferred") })
		})
	}
	input.SetupInputLayersWithDeferred([]input.Layerer{layerer("a"), layerer("b")})
	is.Equal(order, []string{"a", "b", "a deferred", "b deferred"})
}

// A layer kept by a widget and added again in a later frame is on that
// frame's stack, as a new one would be; one not added again is not.
func TestLayer_KeptFromFrameToFrame(t *testing.T) {
	is := is.New(t)
	t.Cleanup(func() { input.SetupInputLayersWithDeferred(nil) })

	kept := &input.Layer{EventTypes: input.LayerEventTypeAll, FullScreen: true}
	left := &input.Layer{EventTypes: input.LayerEventTypeAll, FullScreen: true}
	input.SetupInputLayersWithDeferred(nil)
	input.AddLayer(left)
	input.AddLayer(kept)
	input.SetupInputLayersWithDeferred(nil) // the next frame
	input.AddLayer(kept)

	is.True(kept.Valid() && kept.ActiveFor(10, 10, input.LayerEventTypeMouseButton))
	is.True(!left.Valid() && !left.ActiveFor(10, 10, input.LayerEventTypeMouseButton))
}
