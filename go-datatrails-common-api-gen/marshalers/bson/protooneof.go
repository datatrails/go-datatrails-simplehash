package bson

import (
	"errors"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"

	"github.com/datatrails/go-datatrails-common/logger"
	v2attribute "github.com/datatrails/go-datatrails-simplehash/go-datatrails-common-api-gen/attribute/v2/attribute"
)

// OneofDecoder decodes attribute oneof Attribute
type OneofDecoder struct{}

// DecodeValue decodes aattribute oneof Attribute
func (*OneofDecoder) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {

	dr, err := vr.ReadDocument()
	if err != nil {
		logger.Sugar.Infow("read doc failed", "err", err)
		return err
	}

	for {
		_, vvr, err := dr.ReadElement()
		if errors.Is(err, bsonrw.ErrEOD) {
			break
		}
		if err != nil {
			logger.Sugar.Infow("failed read element", "err", err)
			return err
		}

		if vvr.Type() != bson.TypeEmbeddedDocument {
			err = vvr.Skip()
			if err != nil {
				logger.Sugar.Infow("failed to skip element", "err", err)
				return err
			}
			continue
		}

		ddr, err := vvr.ReadDocument()
		if err != nil {
			logger.Sugar.Infow("failed second doc read", "err", err)
			return err
		}

		err = decodeDoc(ectx, ddr, val)
		if err != nil {
			logger.Sugar.Infow("failed decoding doc", "err", err)
			return err
		}
	}
	return nil
}

func decodeDoc(ectx bsoncodec.DecodeContext, ddr bsonrw.DocumentReader, val reflect.Value) error {
	for {
		pname, vrr, err := ddr.ReadElement()
		if errors.Is(err, bsonrw.ErrEOD) {
			break
		}
		if err != nil {
			logger.Sugar.Infow("failed second element read out", "err", err)
			return err
		}

		switch pname {
		case "strval":
			strVal := reflect.New(reflect.TypeOf(v2attribute.Attribute_StrVal{}))
			if val.Kind() == reflect.Ptr {
				attrVal := reflect.New(reflect.TypeOf(v2attribute.Attribute{}))
				val.Set(attrVal)
				val.Elem().Field(0).Set(strVal)
			} else {
				for i := 0; i < val.NumField(); i++ {
					if val.Field(i).CanSet() {
						val.Field(i).Set(strVal)
						break
					}
				}

			}
			fval := strVal.Elem().Field(0)
			dec, err := ectx.LookupDecoder(fval.Type())
			if err != nil {
				logger.Sugar.Infow("failed to get decoder", "err", err)
				return err
			}

			if err = dec.DecodeValue(ectx, vrr, fval); err != nil {
				logger.Sugar.Infow("failed decode value")
				return err
			}
		case "listval":
			listVal := reflect.New(reflect.TypeOf(v2attribute.Attribute_ListVal{}))
			if val.Kind() == reflect.Ptr {
				attrVal := reflect.New(reflect.TypeOf(v2attribute.Attribute{}))
				val.Set(attrVal)
				val.Elem().Field(0).Set(listVal)
			} else {
				for i := 0; i < val.NumField(); i++ {
					if val.Field(i).CanSet() {
						val.Field(i).Set(listVal)
						break
					}
				}
			}
			fval := listVal.Elem().Field(0)
			dec, err := ectx.LookupDecoder(fval.Type())
			if err != nil {
				logger.Sugar.Infow("failed to get decoder", "err", err)
				return err
			}

			if err = dec.DecodeValue(ectx, vrr, fval); err != nil {
				logger.Sugar.Infow("failed decode value")
				return err
			}
		case "dictval":
			dictval := reflect.New(reflect.TypeOf(v2attribute.Attribute_DictVal{}))
			if val.Kind() == reflect.Ptr {
				attrVal := reflect.New(reflect.TypeOf(v2attribute.Attribute{}))
				val.Set(attrVal)
				val.Elem().Field(0).Set(dictval)
			} else {
				for i := 0; i < val.NumField(); i++ {
					if val.Field(i).CanSet() {
						val.Field(i).Set(dictval)
						break
					}
				}
			}
			fval := dictval.Elem().Field(0)
			dec, err := ectx.LookupDecoder(fval.Type())
			if err != nil {
				logger.Sugar.Infow("failed to get decoder", "err", err)
				return err
			}

			if err = dec.DecodeValue(ectx, vrr, fval); err != nil {
				logger.Sugar.Infow("failed decode value")
				return err
			}
		default:
			logger.Sugar.Infow("default case, we're not handling that field", "name", pname)
		}
	}
	return nil
}
