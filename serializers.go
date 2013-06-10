package pathways

import (
	"code.google.com/p/vitess/go/bson"
	"encoding/json"
	"errors"
	"github.com/vmihailenco/msgpack"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	Serializers = SerializerMap{
		"application/json":      &JsonSerializer{},
		"application/x-msgpack": &MsgpackSerializer{},
		"application/bson":      &BsonSerializer{},
	}
	UnsupportedContentType = errors.New("unsupported content type")
)

type SerializerMap map[string]Serializer

func (s SerializerMap) DecodeRequest(req *http.Request, contentType string, v interface{}) error {
	return s.Decode(contentType, req.Body, v)
}

func (s SerializerMap) Decode(ct string, r io.Reader, v interface{}) error {
	if ser, ok := s[ct]; ok {
		decoder := ser.NewDecoder(r)
		return decoder.Decode(v)
	}
	return UnsupportedContentType
}

type ApiError struct {
	Status int
	Error  string
}

func (s SerializerMap) EncodeResponse(w http.ResponseWriter, code int, contentType string, response interface{}) error {
	// TODO: Figure out ordering here that isn't shit.
	if ser, ok := s[contentType]; ok {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(code)
		s.rawEncode(ser, w, response)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		ser := s["application/json"]
		err := "Invalid content type " + contentType
		s.rawEncode(ser, w, &ApiError{
			Status: code,
			Error:  err,
		})
		return errors.New(err)
	}
	return nil
}

func (s SerializerMap) Encode(ct string, w io.Writer, v interface{}) error {
	if ser, ok := s[ct]; ok {
		return s.rawEncode(ser, w, v)
	}
	return UnsupportedContentType
}

func (s SerializerMap) rawEncode(ser Serializer, w io.Writer, v interface{}) error {
	encoder := ser.NewEncoder(w)
	return encoder.Encode(v)
}

type ContentTypeDecoder interface {
	Decode(v interface{}) error
}

type ContentTypeEncoder interface {
	Encode(v interface{}) error
}

type Serializer interface {
	NewEncoder(w io.Writer) ContentTypeEncoder
	NewDecoder(r io.Reader) ContentTypeDecoder
}

type JsonSerializer struct{}

func (j *JsonSerializer) NewEncoder(w io.Writer) ContentTypeEncoder {
	return json.NewEncoder(w)
}

func (j *JsonSerializer) NewDecoder(r io.Reader) ContentTypeDecoder {
	return json.NewDecoder(r)
}

type MsgpackSerializer struct{}

func (j *MsgpackSerializer) NewEncoder(w io.Writer) ContentTypeEncoder {
	return msgpack.NewEncoder(w)
}

func (j *MsgpackSerializer) NewDecoder(r io.Reader) ContentTypeDecoder {
	return msgpack.NewDecoder(r)
}

type BsonSerializer struct{}

type bsonEncoder struct {
	w io.Writer
}

func (b *bsonEncoder) Encode(v interface{}) error {
	bytes, err := bson.Marshal(v)
	if err != nil {
		return err
	}
	_, err = b.w.Write(bytes)
	return err
}

func (j *BsonSerializer) NewEncoder(w io.Writer) ContentTypeEncoder {
	return &bsonEncoder{w}
}

type bsonDecoder struct {
	r io.Reader
}

func (b *bsonDecoder) Decode(v interface{}) error {
	bytes, err := ioutil.ReadAll(b.r)
	if err != nil {
		return err
	}
	return bson.Unmarshal(bytes, v)
}

func (j *BsonSerializer) NewDecoder(r io.Reader) ContentTypeDecoder {
	return &bsonDecoder{r}
}
