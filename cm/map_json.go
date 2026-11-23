package cm

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"unsafe"
)

func (p pairs[K, V]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}
	for k, v := range p.All() {
		name, err := keyName(k)
		if err != nil {
			return err
		}
		if err := enc.WriteToken(jsontext.String(name)); err != nil {
			return err
		}
		if err := json.MarshalEncode(enc, v); err != nil {
			return err
		}
	}
	return enc.WriteToken(jsontext.EndObject)
}

func keyName[K comparable](k K) (string, error) {
	key, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	if jsontext.Value(key).Kind() != jsontext.KindString {
		return string(key), nil
	}
	var name string
	err = json.Unmarshal(key, &name)
	return name, err
}

func (p *pairs[K, V]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() == jsontext.KindNull {
		_, err := dec.ReadToken()
		return err
	}
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	if tok.Kind() != jsontext.KindBeginObject {
		return fmt.Errorf("cm: map JSON must be an object, not %v", tok.Kind())
	}
	var s []Tuple[K, V]
	for dec.PeekKind() != jsontext.KindEndObject {
		name, err := dec.ReadToken()
		if err != nil {
			return err
		}
		var e Tuple[K, V]
		if err := unmarshalKey(name.String(), &e.F0); err != nil {
			return err
		}
		if err := json.UnmarshalDecode(dec, &e.F1); err != nil {
			return err
		}
		s = append(s, e)
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	p.data = unsafe.SliceData(s)
	p.len = uintptr(len(s))
	return nil
}

func unmarshalKey[K comparable](text string, k *K) error {
	quoted, err := json.Marshal(text)
	if err != nil {
		return err
	}
	if json.Unmarshal(quoted, k) == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(text), k); err != nil {
		return err
	}
	canonical, err := json.Marshal(*k)
	if err != nil {
		return err
	}
	if string(canonical) != text {
		return fmt.Errorf("cm: invalid map key %q", text)
	}
	return nil
}
