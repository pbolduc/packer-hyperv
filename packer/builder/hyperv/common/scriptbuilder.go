package common

import (
	"bytes"
)

type ScriptBuilder struct {
	buffer bytes.Buffer
}

func (b *ScriptBuilder) WriteLine(s string) (n int, err error) {
	n, err = b.buffer.WriteString(s);
	b.buffer.WriteString("\n")

	return n+1, err
}

func (b *ScriptBuilder) String() string {
	return b.buffer.String()
}

func (b *ScriptBuilder) Bytes() []byte {
	return b.buffer.Bytes()
}

func (b *ScriptBuilder) Reset() {
	b.buffer.Reset()
}

