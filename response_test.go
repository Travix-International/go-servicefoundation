package servicefoundation_test

import (
	"net/http"
	"testing"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testObj struct {
	Name string
	Age  int
}

func TestCreateWrappedResponseWriter(t *testing.T) {
	w := &mockResponseWriter{}
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Twice()
	w.
		On("WriteHeader", mock.AnythingOfType("int")).
		Twice()
	sut := sf.NewWrappedResponseWriter(w)

	// Act
	sut.Write([]byte("okay"))
	sut.Write([]byte("then"))
	sut.WriteHeader(http.StatusInternalServerError)

	assert.Equal(t, http.StatusOK, sut.Status())
	w.AssertExpectations(t)
}

func TestWrappedResponseWriterImpl_JSON(t *testing.T) {
	const status = http.StatusGatewayTimeout
	w := &mockResponseWriter{}
	w.
		On("Header").
		Return(http.Header{}).
		Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Once()
	w.
		On("WriteHeader", mock.AnythingOfType("int")).
		Once()
	obj := testObj{"Fifi", 22}
	sut := sf.NewWrappedResponseWriter(w)

	sut.JSON(status, obj)

	assert.Equal(t, status, sut.Status())
	w.AssertExpectations(t)
}

func TestWrappedResponseWriterImpl_XML(t *testing.T) {
	const status = http.StatusAlreadyReported
	w := &mockResponseWriter{}
	w.
		On("Header").
		Return(http.Header{}).
		Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Twice()
	w.
		On("WriteHeader", mock.AnythingOfType("int")).
		Once()
	obj := testObj{"Yoyo", 33}
	sut := sf.NewWrappedResponseWriter(w)

	sut.XML(status, obj)

	assert.Equal(t, status, sut.Status())
	w.AssertExpectations(t)
}

func TestWrappedResponseWriterImpl_AcceptsXML(t *testing.T) {
}

func TestWrappedResponseWriterImpl_AcceptsJSON(t *testing.T) {
	rdr := &mockReader{}
	rdr.On("Read", mock.Anything).
		Return(0).
		Once()

	r, _ := http.NewRequest("GET", "https://www.site.com/some/url", rdr)
	r.Header.Add(sf.AcceptHeader, sf.ContentTypeJSON)

	w := &mockResponseWriter{}
	sut := sf.NewWrappedResponseWriter(w)

	actual := sut.AcceptsXML(r)

	assert.False(t, actual)
	w.AssertExpectations(t)
}

func TestWrappedResponseWriterImpl_WriteResponse_AsXML(t *testing.T) {
	const status = http.StatusAlreadyReported
	rdr := &mockReader{}
	rdr.On("Read", mock.Anything).
		Return(0).
		Once()
	h := make(http.Header)

	r, _ := http.NewRequest("GET", "https://www.site.com/some/url", rdr)
	r.Header.Add(sf.AcceptHeader, sf.ContentTypeXML)

	w := &mockResponseWriter{}
	w.
		On("Header").
		Return(h).
		Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Twice()
	w.
		On("WriteHeader", mock.AnythingOfType("int")).
		Once()
	sut := sf.NewWrappedResponseWriter(w)
	obj := testObj{"Gaga", 44}

	sut.WriteResponse(r, status, obj)

	assert.Equal(t, status, sut.Status())
	assert.Equal(t, sf.ContentTypeXML, h.Get(sf.ContentTypeHeader))
	w.AssertExpectations(t)
}

func TestWrappedResponseWriterImpl_WriteResponse_AsJSON(t *testing.T) {
	const status = http.StatusContinue
	rdr := &mockReader{}
	rdr.On("Read", mock.Anything).
		Return(0).
		Once()
	h := make(http.Header)

	r, _ := http.NewRequest("GET", "https://www.site.com/some/url", rdr)
	r.Header.Add(sf.AcceptHeader, sf.ContentTypeJSON)

	w := &mockResponseWriter{}
	w.
		On("Header").
		Return(h).
		Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Once()
	w.
		On("WriteHeader", mock.AnythingOfType("int")).
		Once()
	sut := sf.NewWrappedResponseWriter(w)
	obj := testObj{"Mumu", 55}

	sut.WriteResponse(r, status, obj)

	assert.Equal(t, status, sut.Status())
	assert.Equal(t, sf.ContentTypeJSON, h.Get(sf.ContentTypeHeader))
	w.AssertExpectations(t)
}

func TestWrappedResponseWriterImpl_SetCaching(t *testing.T) {
	h := make(http.Header)
	w := &mockResponseWriter{}
	w.
		On("Header").
		Return(h).
		Twice()
	sut := sf.NewWrappedResponseWriter(w)

	sut.SetCaching(66)

	assert.Equal(t, "Accept, Origin", h.Get("Vary"))
	assert.Equal(t, "public, max-age=66", h.Get("Cache-Control"))
	w.AssertExpectations(t)
}
