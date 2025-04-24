package simpleoneof

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/datatrails/go-datatrails-common/logger"
	v2attribute "github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/attribute/v2/attribute"
)

type SnakeMarshaler struct {
	m *protojson.MarshalOptions
	u *protojson.UnmarshalOptions
}

func (s *SnakeMarshaler) Marshal(value any) ([]byte, error) {
	message, ok := value.(proto.Message)
	if !ok {
		return nil, errors.New("unable to marshal non proto field")
	}
	return s.m.Marshal(message)
}

func (s *SnakeMarshaler) Unmarshal(data []byte, value any) error {
	message, ok := value.(proto.Message)
	if !ok {
		return errors.New("unable to unmarshal non proto field")
	}
	return s.u.Unmarshal(data, message)
}

// ContentType for a SnakeMarshaler is `application/octet-stream` because
//
//	the SnakeMarshaler encodes a protobuf message into json, in the format
//	of a byte slice.
func (*SnakeMarshaler) ContentType(_ any) string {
	return "application/octet-stream"
}

// NewDecoder returns a Decoder which reads proto stream from "reader".
func (s *SnakeMarshaler) NewDecoder(reader io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(value any) error {
		buffer, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		return s.Unmarshal(buffer, value)
	})
}

// NewEncoder returns an Encoder which writes proto stream into "writer".
func (s *SnakeMarshaler) NewEncoder(writer io.Writer) runtime.Encoder {
	return runtime.EncoderFunc(func(value any) error {
		buffer, err := s.Marshal(value)
		if err != nil {
			return err
		}
		_, err = writer.Write(buffer)
		if err != nil {
			return err
		}

		return nil
	})
}

func NewSnakeMarshaler() *SnakeMarshaler {
	return &SnakeMarshaler{
		m: &protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true, UseEnumNumbers: false},
		u: &protojson.UnmarshalOptions{},
	}
}

var (
	// ErrUnexpectedType is returned if we got
	// type we can't handle
	ErrUnexpectedType = errors.New("unexpected type in structure")

	// ErrUnexpectedFieldInNestedStruct returned when the nested
	// struct has unexpected fields (other than string,list and dict)
	ErrUnexpectedFieldInNestedStruct = errors.New("unexpected field in nested struct")

	// ErrFailedToDecodeRequest returned when we can't
	// decode request data into proto.Message
	ErrFailedToDecodeRequest = errors.New("failed to decode request data")
)

// Marshaler marshals proto3 oneof into flat json structures
type Marshaler struct {
	SingleTypes []reflect.Type
	NestedTypes []reflect.Type

	m                runtime.Marshaler
	deepStructType   reflect.Type
	attributeNames   map[string]string
	inAttributeNames map[string]string

	ignoreFields []string
}

type writerWithError struct {
	w   *bytes.Buffer
	err error
}

func (w *writerWithError) WriteNoErr(p []byte) {
	_, err := w.w.Write(p)
	if err != nil && w.err == nil {
		w.err = err
	}
}

func (w *writerWithError) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *writerWithError) Bytes() []byte {
	return w.w.Bytes()
}

func (w *writerWithError) Error() error {
	return w.err
}

// NewFlatMarshalerForAssetsV2 creates marshaler for assetsv2
func NewFlatMarshalerForAssetsV2(
	singleTypes []reflect.Type,
	nestedTypes []reflect.Type,
	ignoreFields []string,
) *Marshaler {
	return &Marshaler{
		SingleTypes: singleTypes,
		NestedTypes: nestedTypes,

		m:              NewSnakeMarshaler(),
		deepStructType: reflect.TypeOf(v2attribute.Attribute{}),
		attributeNames: map[string]string{
			"event_attributes": "EventAttributes",
			"asset_attributes": "AssetAttributes",
			"attributes":       "Attributes",
			"access_policy":    "AccessPolicy",
		},
		inAttributeNames: map[string]string{
			"attributes": "Attributes",
		},
		ignoreFields: ignoreFields,
	}
}

// NewFlatMarshalerForLocations creates marshaler
func NewFlatMarshalerForLocations(
	singleTypes []reflect.Type,
	nestedTypes []reflect.Type,
) *Marshaler {
	return &Marshaler{
		SingleTypes: singleTypes,
		NestedTypes: nestedTypes,

		m:              NewSnakeMarshaler(),
		deepStructType: reflect.TypeOf(v2attribute.Attribute{}),
		attributeNames: map[string]string{
			"attributes": "Attributes",
		},
		inAttributeNames: map[string]string{
			"attributes": "Attributes",
		},
	}
}

// NewFlatMarshaler creates marshaler
func NewFlatMarshaler(
	singleTypes []reflect.Type,
	nestedTypes []reflect.Type,
) *Marshaler {
	return &Marshaler{
		SingleTypes: singleTypes,
		NestedTypes: nestedTypes,

		m:              NewSnakeMarshaler(),
		deepStructType: reflect.TypeOf(v2attribute.Attribute{}),
		attributeNames: map[string]string{
			"event_attributes": "EventAttributes",
			"asset_attributes": "AssetAttributes",
			"attributes":       "Attributes",
		},
		inAttributeNames: map[string]string{
			"event_attributes": "EventAttributes",
			"asset_attributes": "AssetAttributes",
		},
	}
}

func (m *Marshaler) isNestedType(t reflect.Type) bool {
	for _, lT := range m.NestedTypes {
		if t == lT {
			return true
		}
	}
	return false
}

func (m *Marshaler) isSingleType(t reflect.Type) bool {
	for _, lT := range m.SingleTypes {
		if t == lT {
			return true
		}
	}
	return false
}

// Marshal marshal v into json bytes
func (m *Marshaler) Marshal(v any) ([]byte, error) {

	s := reflect.ValueOf(v).Elem()
	t := s.Type()
	if m.isNestedType(t) {
		return m.marshalNestedResponse(s.Interface(), t)
	}

	if m.isSingleType(t) {
		return m.marshalSingleResponse(v)
	}

	return m.m.Marshal(v)
}

func (m *Marshaler) skipField(jName string) bool {
	if jName == "-" || jName == "" {
		return true
	}
	for _, f := range m.ignoreFields {
		if jName == f {
			return true
		}
	}
	return false
}

func (m *Marshaler) marshalNestedResponse(resp any, t reflect.Type) ([]byte, error) {
	v := reflect.ValueOf(resp)
	buf := writerWithError{w: &bytes.Buffer{}}
	buf.WriteNoErr([]byte("{"))
	first := true
	for i := 0; i < t.NumField(); i++ {
		prop := t.FieldByIndex([]int{i})
		jTag := prop.Tag.Get("json")
		jName := strings.Split(jTag, ",")[0]
		if m.skipField(jName) {
			continue
		}
		if !first {
			buf.WriteNoErr([]byte(","))
		}
		first = false
		jNameval, merr := json.Marshal(jName)
		if merr != nil {
			return nil, merr
		}
		buf.WriteNoErr(jNameval)
		buf.WriteNoErr([]byte(": "))

		vv := v.Field(i)
		switch vv.Kind() {
		case reflect.String:
			vval, err := json.Marshal(vv.String())
			if err != nil {
				return nil, err
			}
			buf.WriteNoErr(vval)
		case reflect.Slice:
			buf.WriteNoErr([]byte(" ["))
			for j := 0; j < vv.Len(); j++ {
				if j > 0 {
					buf.WriteNoErr([]byte(","))
				}
				bytesData, err := m.marshalSingleResponse(vv.Index(j).Interface())
				if err != nil {
					return nil, err
				}
				buf.WriteNoErr(bytesData)
			}
			buf.WriteNoErr([]byte("]"))
		case reflect.Ptr:
			if vv.IsNil() {
				buf.WriteNoErr([]byte(`null`))
				continue
			}
			ss := vv.Elem()

			bytesData, err := m.marshalSingleResponse(ss.Interface())
			if err != nil {
				return nil, err
			}
			buf.WriteNoErr(bytesData)
		default:
			logger.Sugar.Info("Expected slice or string but didn't get one")
			return nil, ErrUnexpectedType
		}
	}
	buf.WriteNoErr([]byte("}"))
	return buf.Bytes(), buf.Error()
}

func (m *Marshaler) marshalSingleResponse(resp any) ([]byte, error) {

	buf := writerWithError{w: &bytes.Buffer{}}
	buf.WriteNoErr([]byte("{"))

	msg := resp.(proto.Message)
	desc := msg.ProtoReflect().Descriptor()
	firstAttr := true
	for i := 0; i < desc.Fields().Len(); i++ {
		var bts []byte
		var e error
		jName := string(desc.Fields().Get(i).Name())
		if m.skipField(jName) {
			continue
		}

		if !firstAttr {
			buf.WriteNoErr([]byte(", "))
		}
		firstAttr = false
		jNameval, merr := json.Marshal(jName)
		if merr != nil {
			return nil, merr
		}
		buf.WriteNoErr(jNameval)
		buf.WriteNoErr([]byte(": "))
		if desc.Fields().Get(i).Kind() == protoreflect.MessageKind {
			// check if we have custom field here
			if _, ok := m.attributeNames[string(desc.Fields().Get(i).Name())]; ok {
				buf.WriteNoErr([]byte("{"))
				first := true
				msg.ProtoReflect().Get(desc.Fields().Get(i)).Map().Range(
					func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
						if !first {
							buf.WriteNoErr([]byte(", "))
						}
						first = false
						jNameval, merr := json.Marshal(mk.String())
						if merr != nil {
							return false
						}
						buf.WriteNoErr(jNameval)
						buf.WriteNoErr([]byte(": "))
						vvx := reflect.ValueOf(mv.Message().Interface())
						e = m.marshalAttribute(&buf, vvx, false)
						if e != nil {
							logger.Sugar.Infof("failed marshalling attributes %v", e)
							return false
						}
						return true
					},
				)
				buf.WriteNoErr([]byte("}"))
				continue
			}

			// just normal message - let the original marshaler take care of that
			bts, e = m.m.Marshal(msg.ProtoReflect().Get(desc.Fields().Get(i)).Message().Interface())
			if e != nil {
				logger.Sugar.Infof("failed marshalling message %v", e)
				return nil, e
			}
			buf.WriteNoErr(bts)
			continue
		}

		// check for list
		if desc.Fields().Get(i).IsList() {
			lv := msg.ProtoReflect().Get(desc.Fields().Get(i)).List()
			buf.WriteNoErr([]byte("["))
			for j := 0; j < lv.Len(); j++ {
				if j > 0 {
					buf.WriteNoErr([]byte(", "))
				}
				jNameval, merr := json.Marshal(lv.Get(j).String())
				if merr != nil {
					return nil, merr
				}
				buf.WriteNoErr(jNameval)
			}
			buf.WriteNoErr([]byte("] "))
			continue
		}

		// if enum handle here
		if desc.Fields().Get(i).Kind() == protoreflect.EnumKind {
			bts = []byte("\"" +
				desc.Fields().Get(i).Enum().Values().ByNumber(msg.ProtoReflect().Get(desc.Fields().Get(i)).Enum()).Name() +
				"\"")
			buf.WriteNoErr(bts)
			continue
		}

		// everything else we just use json marshal
		bts, e = json.Marshal(msg.ProtoReflect().Get(desc.Fields().Get(i)).Interface())
		if e != nil {
			logger.Sugar.Infof("failed marshalling json response %v", e)
			return nil, e
		}
		buf.WriteNoErr(bts)
	}

	buf.WriteNoErr([]byte("}"))
	return buf.Bytes(), buf.Error()
}

func (m *Marshaler) marshalAttribute(w *writerWithError, v reflect.Value, flatmode bool) error {

	switch v.Kind() {
	case reflect.String:
		strv, err := json.Marshal(v.String())
		if err != nil {
			return err
		}
		w.WriteNoErr(strv)
		return w.Error()
	case reflect.Slice:
		w.WriteNoErr([]byte(" ["))
		for j := 0; j < v.Len(); j++ {
			if j > 0 {
				w.WriteNoErr([]byte(","))
			}
			err := m.marshalAttribute(w, v.Index(j), flatmode)
			if err != nil {
				return err
			}
		}
		w.WriteNoErr([]byte("]"))
		return w.Error()
	case reflect.Map:
		w.WriteNoErr([]byte(" {"))
		for j, kv := range v.MapKeys() {
			if j > 0 {
				w.WriteNoErr([]byte(","))
			}
			kstrv, err := json.Marshal(kv.String())
			if err != nil {
				return err
			}
			vstrv, err := json.Marshal(v.MapIndex(kv).String())
			if err != nil {
				return err
			}
			w.WriteNoErr(kstrv)
			w.WriteNoErr([]byte(" :"))
			w.WriteNoErr(vstrv)
		}
		w.WriteNoErr([]byte("}"))
		return w.Error()
	case reflect.Ptr:
		return m.marshalAttribute(w, v.Elem(), flatmode)
	case reflect.Struct:
		if v.Type() == m.deepStructType || flatmode {
			for j := 0; j < v.NumField(); j++ {
				if v.Field(j).CanSet() {
					return m.marshalAttribute(w, v.Field(j), true)
				}
			}
		}

		data, err := json.Marshal(v.Interface())
		if err != nil {
			logger.Sugar.Info("failed to do original marshalling for a struct")
			return err
		}
		w.WriteNoErr(data)
		return w.Error()
	case reflect.Interface: // assume oneof or nil here
		if !v.IsNil() {
			return m.marshalAttribute(w, v.Elem().Elem(), flatmode)
		}
		nilv, err := json.Marshal(nil)
		if err != nil {
			return err
		}
		w.WriteNoErr(nilv)
		return w.Error()
	default:
		logger.Sugar.Infow("got default Kind", "k", v.Kind())
		// return ErrUnexpectedType
		return ErrUnexpectedType
	}
}

// Unmarshal unmarshal json data into v
func (m *Marshaler) Unmarshal(data []byte, v any) error {
	return m.m.Unmarshal(data, v)
}

// NewDecoder create new WrappedDecoder
func (m *Marshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return &WrappedDecoder{r, m.inAttributeNames}
}

// NewEncoder get underlying proto encoder
func (m *Marshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return m.m.NewEncoder(w)
}

// ContentType returns `application/json` content type.
//
// This is because the Marshaler encodes our proto messages
//
//	as json.
func (m *Marshaler) ContentType(pb any) string {
	return "application/json; charset=utf-8"
}

// WrappedDecoder decoder wrapping json with custom logic handling oneofs
type WrappedDecoder struct {
	r              io.Reader
	attributeNames map[string]string
}

func unmarshalNestedStructs(v any) (any, error) {
	switch v := v.(type) {
	case string:
		return v, nil
	case []any:
		ret := []map[string]string{}
		for _, vv := range v {
			data, err := unmarshalNestedStructs(vv)
			if err != nil {
				return "", err
			}
			dataDict, ok := data.(map[string]string)
			if !ok {
				logger.Sugar.Infof("unexpected field in nested struct")
				return "", ErrUnexpectedFieldInNestedStruct
			}
			ret = append(ret, dataDict)
		}
		return ret, nil
	case map[string]any:
		ret := map[string]string{}
		for k, vv := range v {
			data, err := unmarshalNestedStructs(vv)
			if err != nil {
				return "", err
			}
			dataStr, ok := data.(string)
			if !ok {
				logger.Sugar.Infof("unexpected field in nested struct")
				return "", ErrUnexpectedFieldInNestedStruct
			}
			ret[k] = dataStr
		}
		return ret, nil
	default:
		logger.Sugar.Infof("unexpected field in nested struct: %v", reflect.TypeOf(v))
		return "", ErrUnexpectedFieldInNestedStruct
	}
}

func (m *WrappedDecoder) decodeOneOfAttributes(data json.RawMessage, protomap protoreflect.Map) error {

	stringVals := map[string]string{}
	dictVals := map[string]map[string]string{}
	listVals := map[string][]map[string]string{}

	vals := map[string]any{}
	err := json.Unmarshal(data, &vals)
	if err != nil {
		logger.Sugar.Infow("failed to unmarshal into object")
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}
	for kk, vvv := range vals {
		switch vvv.(type) {
		case string:
			data, err := unmarshalNestedStructs(vvv)
			if err != nil {
				return err
			}
			dataStr, ok := data.(string)
			if !ok {
				logger.Sugar.Infow("unexpected field in nested struct")
				return ErrUnexpectedFieldInNestedStruct
			}
			stringVals[kk] = dataStr
		case []any:
			data, err := unmarshalNestedStructs(vvv)
			if err != nil {
				return err
			}
			dataList, ok := data.([]map[string]string)
			if !ok {
				logger.Sugar.Infow("unexpected field in nested struct")
				return ErrUnexpectedFieldInNestedStruct
			}
			listVals[kk] = dataList
		case map[string]any:
			data, err := unmarshalNestedStructs(vvv)
			if err != nil {
				return err
			}
			dataDict, ok := data.(map[string]string)
			if !ok {
				logger.Sugar.Infow("unexpected field in nested struct")
				return ErrUnexpectedFieldInNestedStruct
			}
			dictVals[kk] = dataDict
		default:
			logger.Sugar.Infow("unexpected type in the complex oneof", "t", reflect.TypeOf(vvv))
			return ErrUnexpectedFieldInNestedStruct
		}
	}

	complexAttr := map[string]*v2attribute.Attribute{}

	// slot in the lists dicts and strings into the oneofs
	for key, val := range stringVals {
		complexAttr[key] = &v2attribute.Attribute{
			Value: &v2attribute.Attribute_StrVal{
				StrVal: val,
			},
		}
	}

	for key, val := range dictVals {
		complexAttr[key] = &v2attribute.Attribute{
			Value: &v2attribute.Attribute_DictVal{
				DictVal: &v2attribute.DictAttr{
					Value: val,
				},
			},
		}
	}

	for key, val := range listVals {
		listAttr := []*v2attribute.DictAttr{}
		for _, dval := range val {
			listAttr = append(listAttr, &v2attribute.DictAttr{
				Value: dval,
			})
		}
		complexAttr[key] = &v2attribute.Attribute{
			Value: &v2attribute.Attribute_ListVal{
				ListVal: &v2attribute.ListAttr{
					Value: listAttr,
				},
			},
		}
	}

	for k, v := range complexAttr {
		protomap.Set(protoreflect.MapKey(protoreflect.ValueOf(k)), protoreflect.ValueOf(v.ProtoReflect()))
	}

	return nil
}

// Decode bytes into obj
func (m *WrappedDecoder) Decode(obj any) error {
	bytesData, err := io.ReadAll(m.r)
	if err != nil {
		logger.Sugar.Infow("failed reading", "err", err)
		return fmt.Errorf("could not read the content: %w", err)
	}

	if reflect.ValueOf(obj).IsNil() {
		valueType := reflect.ValueOf(obj).Elem().Type()
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
			objvalueptr := reflect.New(valueType)
			if reflect.ValueOf(obj).Elem().CanSet() {
				reflect.ValueOf(obj).Elem().Set(objvalueptr)
			}
		}
	}

	objvalue := reflect.ValueOf(obj).Elem()
	var valueType reflect.Type
	valueType = reflect.ValueOf(obj).Elem().Type()

	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
		objvalueptr := reflect.New(valueType)
		if reflect.ValueOf(obj).Elem().CanSet() {
			reflect.ValueOf(obj).Elem().Set(objvalueptr)
		}
		obj = objvalueptr.Elem().Interface()
	}

	var nn proto.Message
	nn, ok := obj.(proto.Message)
	if !ok {
		logger.Sugar.Infof("obj not a proto.Message")
		nn, ok = objvalue.Interface().(proto.Message)
		if !ok {
			logger.Sugar.Infof("objvalue not a proto.Message 2")
			return ErrFailedToDecodeRequest
		}
	}

	logger.Sugar.Debugf("Object is: %#v", obj)
	_, ok = obj.(protoreflect.Message)
	if !ok {
		logger.Sugar.Infof("warn: Not a protoreflect.Message: %T", obj)
	}

	data := map[string]json.RawMessage{}
	//unmarshal here so we have attr names
	//and raw json we can unmarshal piecemeal
	err = json.Unmarshal(bytesData, &data)
	if err != nil {
		logger.Sugar.Infow("failed to unmarshal into raw map")
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}

	newMsg := nn.ProtoReflect()

	lenFields := newMsg.Descriptor().Fields().Len()
	for a := 0; a < lenFields; a++ {
		f := newMsg.Descriptor().Fields().Get(a)

		if _, ok := data[string(f.Name())]; !ok {
			continue
		}
		v := newMsg.Get(f)

		interfaceV := v.Interface()

		switch f.Kind() {
		case protoreflect.MessageKind:
			if _, ok := m.attributeNames[string(f.Name())]; ok {
				vvv := newMsg.Mutable(f).Map()
				e := m.decodeOneOfAttributes(data[string(f.Name())], vvv)
				if e != nil {
					logger.Sugar.Infof("failed decoding custom attr with %v", err)
					return e
				}
				continue

			}

			vvv := newMsg.Mutable(f)

			e := protojson.Unmarshal(data[string(f.Name())], vvv.Message().Interface())
			if e != nil {
				return e
			}
			newMsg.Set(f, vvv)

		case protoreflect.EnumKind:

			ee := json.Unmarshal(data[string(f.Name())], &interfaceV)
			if ee != nil {
				logger.Sugar.Infof("ERR: %v", ee)
				return err
			}

			name := protoreflect.ValueOf(interfaceV).String()

			// Only support the ENUM_NAME form, don't try to parse numbers and look up codes.
			enumFieldValue := f.Enum().Values().ByName(protoreflect.Name(name))
			if enumFieldValue == nil {
				return fmt.Errorf("enum field value invalid. '%s' not found", name)
			}
			enumValue := protoreflect.ValueOfEnum(enumFieldValue.Number())

			newMsg.Set(f, enumValue)

		case protoreflect.Int32Kind:
			{
				// We know to expect an int32 so we can be explicit about it.
				var hopefulInt32 int32
				ee := json.Unmarshal(data[string(f.Name())], &hopefulInt32)
				if ee != nil {
					logger.Sugar.Infof("ERR: %v", ee)
					return ee
				}

				v = protoreflect.ValueOf(hopefulInt32)
				newMsg.Set(f, v)
			}

		default:
			if f.IsList() {
				vv := []json.RawMessage{}
				ees := json.Unmarshal(data[string(f.Name())], &vv)
				if ees != nil {
					logger.Sugar.Infof("failed unmarshalling chunk: %v", ees)
					return ees
				}
				for _, vvv := range vv {
					interfaceVal := v.List().NewElement().Interface()
					eex := json.Unmarshal(vvv, &interfaceVal)
					if eex != nil {
						logger.Sugar.Infof("ERR: %v", eex)
					}
					newMsg.Mutable(f).List().Append(protoreflect.ValueOf(interfaceVal))
				}
				continue
			}

			ee := json.Unmarshal(data[string(f.Name())], &interfaceV)
			if ee != nil {
				logger.Sugar.Infof("ERR: %v", ee)
				return err
			}
			v = protoreflect.ValueOf(interfaceV)
			// Execute an inline function which recovers a panic when trying
			// to convert a string to a float - see #4598
			err = func() (myerr error) {
				defer func() {
					// despite many examples on the interweb the value
					// returned from recover() is **not** an error - its
					// just a plain any
					if r := recover(); r != nil {
						myerr = fmt.Errorf("recover %q", r)
					}
				}()
				newMsg.Set(f, v)
				return
			}()
			if err != nil {
				logger.Sugar.Infof("Error when setting value: %v", err)
				return ErrFailedToDecodeRequest
			}
		}
	}

	return nil
}
