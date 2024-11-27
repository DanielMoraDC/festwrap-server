package serialization

type FakeDeserializer[T any] struct {
	bytes  []byte
	result T
	err    error
}

func (d *FakeDeserializer[T]) GetArgs() []byte {
	return d.bytes
}

func (d *FakeDeserializer[T]) SetResponse(result T) {
	d.result = result
}

func (d *FakeDeserializer[T]) SetError(err error) {
	d.err = err
}

func (d *FakeDeserializer[T]) Deserialize(bytes []byte, dest *T) error {
	d.bytes = bytes
	*dest = d.result
	if d.err != nil {
		return d.err
	}
	return nil
}
