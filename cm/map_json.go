package cm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"
)

func (p pairs[K, V]) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	first := true
	var err error
	for k, v := range p.All() {
		if !first {
			b.WriteByte(',')
		}
		first = false
		var key, value []byte
		key, err = marshalKey(k)
		if err != nil {
			return nil, err
		}
		value, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')
		b.Write(value)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func marshalKey[K comparable](k K) ([]byte, error) {
	key, err := json.Marshal(k)
	if err != nil || key[0] == '"' {
		return key, err
	}
	return json.Marshal(string(key))
}

func (p *pairs[K, V]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('{') {
		return errors.New("cm: map JSON must be an object")
	}
	var s []Tuple[K, V]
	for dec.More() {
		tok, err = dec.Token()
		if err != nil {
			return err
		}
		var e Tuple[K, V]
		if err = unmarshalKey(tok.(string), &e.F0); err != nil {
			return err
		}
		if err = dec.Decode(&e.F1); err != nil {
			return err
		}
		s = append(s, e)
	}
	if _, err = dec.Token(); err != nil {
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
