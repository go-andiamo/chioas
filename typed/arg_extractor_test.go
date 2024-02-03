package typed

import (
	"errors"
	"github.com/go-andiamo/chioas"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type myString1 string
type myString2 string
type myString3 string
type myString4 string

func TestIsArgExtractor(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	ax, err := isArgExtractor(nil)
	assert.Nil(t, ax)
	assert.NoError(t, err)

	ab := ArgExtractor[myString1]{}
	ax, err = isArgExtractor(ab)
	assert.Nil(t, ax)
	assert.Error(t, err)
	ab2 := &ArgExtractor[myString2]{}
	ax, err = isArgExtractor(ab2)
	assert.Nil(t, ax)
	assert.Error(t, err)

	ab = ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo", nil
		},
		ReadsBody: true,
	}
	ax, err = isArgExtractor(ab)
	assert.NoError(t, err)
	assert.NotNil(t, ax)
	assert.True(t, ax.readsBody)
	assert.Equal(t, reflect.TypeOf(myString1("")), ax.forType)
	assert.NotNil(t, ax.fn)
	assert.True(t, ax.fn.IsValid())
	v, err := ax.extract(req)
	assert.NoError(t, err)
	assert.Equal(t, myString1("foo"), v.Interface())

	ab2 = &ArgExtractor[myString2]{
		Extract: func(r *http.Request) (myString2, error) {
			return "foo", errors.New("fooey")
		},
		ReadsBody: true,
	}
	ax, err = isArgExtractor(ab2)
	assert.NoError(t, err)
	assert.NotNil(t, ax)
	assert.True(t, ax.readsBody)
	assert.Equal(t, reflect.TypeOf(myString2("")), ax.forType)
	assert.NotNil(t, ax.fn)
	assert.True(t, ax.fn.IsValid())
	v, err = ax.extract(req)
	assert.Error(t, err)
	assert.Equal(t, "fooey", err.Error())
	assert.Equal(t, myString2("foo"), v.Interface())
}

func TestIsArgExtractor_Func(t *testing.T) {
	fn := func(req *http.Request) (myString4, error) {
		return "foo4", nil
	}
	ax, err := isArgExtractor(fn)
	assert.NoError(t, err)
	assert.NotNil(t, ax)
	assert.False(t, ax.readsBody)
	assert.Equal(t, ax.forType, reflect.TypeOf(myString4("")))
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := ax.extract(req)
	assert.NoError(t, err)
	assert.Equal(t, myString4("foo4"), v.Interface())

	fn2 := func() {}
	_, err = isArgExtractor(fn2)
	assert.Error(t, err)
	assert.Equal(t, "arg extractor func must have signature 'func(*http.Request) (T, error)'", err.Error())
}

func TestArgExtractors_Add(t *testing.T) {
	axs := argExtractors{}
	assert.Equal(t, 0, len(axs))
	ax, _ := isArgExtractor(ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo", nil
		},
		ReadsBody: true,
	})
	err := axs.add(ax)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(axs))
	err = axs.add(ax)
	assert.Error(t, err)
}

func TestArgExtractors_IsApplicable(t *testing.T) {
	ax1, _ := isArgExtractor(ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo1", nil
		},
		ReadsBody: true,
	})
	ax2, _ := isArgExtractor(ArgExtractor[myString2]{
		Extract: func(r *http.Request) (myString2, error) {
			return "", errors.New("fooey")
		},
	})
	axs := argExtractors{}
	_ = axs.add(ax1)
	_ = axs.add(ax2)

	is, rb := axs.IsApplicable(reflect.TypeOf(myString1("")), "", "")
	assert.True(t, is)
	assert.True(t, rb)
	is, rb = axs.IsApplicable(reflect.TypeOf(myString2("")), "", "")
	assert.True(t, is)
	assert.False(t, rb)
	is, _ = axs.IsApplicable(reflect.TypeOf(myString3("")), "", "")
	assert.False(t, is)
}

func TestArgExtractors_BuildValue(t *testing.T) {
	ax1, _ := isArgExtractor(ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo1", nil
		},
		ReadsBody: true,
	})
	ax2, _ := isArgExtractor(ArgExtractor[myString2]{
		Extract: func(r *http.Request) (myString2, error) {
			return "", errors.New("fooey")
		},
	})
	axs := argExtractors{}
	_ = axs.add(ax1)
	_ = axs.add(ax2)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	rv, err := axs.BuildValue(reflect.TypeOf(myString1("")), req, nil)
	assert.NoError(t, err)
	assert.Equal(t, myString1("foo1"), rv.Interface())
	_, err = axs.BuildValue(reflect.TypeOf(myString2("")), req, nil)
	assert.Error(t, err)
	_, err = axs.BuildValue(reflect.TypeOf(myString3("")), req, nil)
	assert.Error(t, err)
}

func TestNewTypedMethodsHandlerBuilder_WithExtractors(t *testing.T) {
	mhb := NewTypedMethodsHandlerBuilder(ArgExtractor[myString1]{})
	assert.NotNil(t, mhb)
	raw, ok := mhb.(*builder)
	assert.True(t, ok)
	assert.Error(t, raw.initErr)
	assert.Equal(t, "ArgExtractor[T] missing 'Extract' function", raw.initErr.Error())

	mhb = NewTypedMethodsHandlerBuilder(ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo1", nil
		},
	})
	assert.NotNil(t, mhb)
	raw, ok = mhb.(*builder)
	assert.True(t, ok)
	assert.NoError(t, raw.initErr)

	rcvd := ""
	hf, err := mhb.BuildHandler("/", http.MethodGet, chioas.Method{
		Handler: func(my myString1) {
			rcvd = string(my)
		},
	}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()

	hf.ServeHTTP(res, req)
	assert.Equal(t, "foo1", rcvd)
}

func TestNewTypedMethodsHandlerBuilder_WithDuplicateExtractors(t *testing.T) {
	mhb := NewTypedMethodsHandlerBuilder(ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo1", nil
		},
	}, ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "foo1", nil
		},
	})
	assert.NotNil(t, mhb)
	raw, ok := mhb.(*builder)
	assert.True(t, ok)
	assert.Error(t, raw.initErr)
	assert.Contains(t, raw.initErr.Error(), "multiple mappings for arg type")
}

func TestTestNewTypedMethodsHandlerBuilder_WithMultipleBodyReadExtractors(t *testing.T) {
	mhb := NewTypedMethodsHandlerBuilder(ArgExtractor[myString1]{
		Extract: func(r *http.Request) (myString1, error) {
			return "", nil
		},
		ReadsBody: true,
	}, ArgExtractor[myString2]{
		Extract: func(r *http.Request) (myString2, error) {
			return "", nil
		},
		ReadsBody: true,
	})
	assert.NotNil(t, mhb)
	raw, ok := mhb.(*builder)
	assert.True(t, ok)
	assert.NoError(t, raw.initErr)
	mdef := chioas.Method{
		Handler: func(my1 myString1, my2 myString2) error {
			return nil
		},
	}
	_, err := mhb.BuildHandler("/", http.MethodGet, mdef, nil)
	assert.Error(t, err)
	assert.Equal(t, "error building in args (path: /, method: GET) - multiple args could be from request.Body", err.Error())
}

type Id string

func TestArgExtractor_Example(t *testing.T) {
	idExtractor := &ArgExtractor[Id]{
		Extract: func(r *http.Request) (Id, error) {
			return Id(chi.URLParam(r, "id")), nil
		},
	}
	mhb := NewTypedMethodsHandlerBuilder(idExtractor)
	gotPersonId := ""
	mdef := chioas.Method{
		Handler: func(personId Id) {
			gotPersonId = string(personId)
		},
	}
	hf, err := mhb.BuildHandler("/people/{id}", http.MethodGet, mdef, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)
	r := chi.NewRouter()
	r.Delete("/people/{id}", hf)
	req, _ := http.NewRequest(http.MethodDelete, "/people/123", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, "123", gotPersonId)
}
