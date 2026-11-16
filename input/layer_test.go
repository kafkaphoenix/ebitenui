package input

import (
	"image"
	"testing"

	"github.com/matryer/is"
	"github.com/stretchr/testify/mock"
)

type layererMock struct {
	mock.Mock
	setupInputLayerCall *mock.Call
}

func TestSetupInputLayersWithDeferred(t *testing.T) {
	l := newLayererMock(nil)
	SetupInputLayersWithDeferred([]Layerer{l})
	l.AssertExpectations(t)
}

func TestSetupInputLayersWithDeferred_Deferred(t *testing.T) {
	is := is.New(t)

	called := false
	l := newLayererMock(func(def DeferredSetupInputLayerFunc) {
		def(func(_ DeferredSetupInputLayerFunc) {
			called = true
		})
	})

	SetupInputLayersWithDeferred([]Layerer{l})

	is.True(called)
}

func TestLayer_ActiveFor_FullScreen(t *testing.T) {
	is := is.New(t)

	l1 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	AddLayer(&l1)

	l2 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	AddLayer(&l2)

	is.True(l1.ActiveFor(100, 100, LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(100, 100, LayerEventTypeMouseButton))
}

func TestLayer_ActiveFor_BlockLower(t *testing.T) {
	is := is.New(t)

	l1 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	AddLayer(&l1)

	l2 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: true,
		FullScreen: true,
	}
	AddLayer(&l2)

	is.True(!l1.ActiveFor(100, 100, LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(100, 100, LayerEventTypeMouseButton))
}

func TestLayer_ActiveFor_Rect(t *testing.T) {
	is := is.New(t)

	l1 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: false,
		RectFunc: func() image.Rectangle {
			return image.Rect(0, 0, 50, 50)
		},
	}
	AddLayer(&l1)

	l2 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: false,
		RectFunc: func() image.Rectangle {
			return image.Rect(20, 20, 70, 70)
		},
	}
	AddLayer(&l2)

	is.True(l1.ActiveFor(10, 10, LayerEventTypeMouseButton))
	is.True(!l2.ActiveFor(10, 10, LayerEventTypeMouseButton))

	is.True(l1.ActiveFor(30, 30, LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(30, 30, LayerEventTypeMouseButton))

	is.True(!l1.ActiveFor(100, 100, LayerEventTypeMouseButton))
	is.True(!l2.ActiveFor(100, 100, LayerEventTypeMouseButton))
}

func TestLayer_ActiveFor_EventType(t *testing.T) {
	is := is.New(t)

	l1 := Layer{
		EventTypes: LayerEventTypeAll,
		BlockLower: false,
		FullScreen: true,
	}
	AddLayer(&l1)

	l2 := Layer{
		EventTypes: LayerEventTypeMouseButton,
		BlockLower: false,
		FullScreen: true,
	}
	AddLayer(&l2)

	is.True(l1.ActiveFor(100, 100, LayerEventTypeMouseButton))
	is.True(l2.ActiveFor(100, 100, LayerEventTypeMouseButton))

	is.True(l1.ActiveFor(100, 100, LayerEventTypeWheel))
	is.True(!l2.ActiveFor(100, 100, LayerEventTypeWheel))
}

func newLayererMock(f SetupInputLayerFunc) *layererMock {
	l := layererMock{}
	l.setupInputLayerCall = l.On("SetupInputLayer", mock.Anything)
	if f != nil {
		l.setupInputLayerCall.Run(func(args mock.Arguments) {
			def := args[0].(DeferredSetupInputLayerFunc)
			f(def)
		})
	}
	return &l
}

func (l *layererMock) SetupInputLayer(def DeferredSetupInputLayerFunc) {
	l.Called(def)
}

// funcLayerer sets its layers up with a function.
type funcLayerer func(def DeferredSetupInputLayerFunc)

func (f funcLayerer) SetupInputLayer(def DeferredSetupInputLayerFunc) { f(def) }

// What the layerers defer runs after every layerer has set its own layers
// up, in the order deferred, so a layer deferred lies above them all.
func TestSetupInputLayersWithDeferred_DeferredAfterAll(t *testing.T) {
	is := is.New(t)

	var order []string
	layerer := func(name string) Layerer {
		return funcLayerer(func(def DeferredSetupInputLayerFunc) {
			order = append(order, name)
			def(func(DeferredSetupInputLayerFunc) { order = append(order, name+" deferred") })
		})
	}
	SetupInputLayersWithDeferred([]Layerer{layerer("a"), layerer("b")})
	is.Equal(order, []string{"a", "b", "a deferred", "b deferred"})
}

// A layer kept by a widget and added again in a later frame is on that
// frame's stack, as a new one would be; one not added again is not.
func TestLayer_KeptFromFrameToFrame(t *testing.T) {
	is := is.New(t)
	t.Cleanup(func() { SetupInputLayersWithDeferred(nil) })

	kept := &Layer{EventTypes: LayerEventTypeAll, FullScreen: true}
	left := &Layer{EventTypes: LayerEventTypeAll, FullScreen: true}
	SetupInputLayersWithDeferred(nil)
	AddLayer(left)
	AddLayer(kept)
	SetupInputLayersWithDeferred(nil) // the next frame
	AddLayer(kept)

	is.True(kept.Valid() && kept.ActiveFor(10, 10, LayerEventTypeMouseButton))
	is.True(!left.Valid() && !left.ActiveFor(10, 10, LayerEventTypeMouseButton))
}
