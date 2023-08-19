package deck

import (
	"reflect"
	"testing"
)

func TestEncrypt(t *testing.T) {
	key := []byte("hello world")
	c := NewCard(Spades, 1)

	encOutput, err := EncryptCard(key, c)
	if err != nil {
		t.Errorf("enc error %s\n", err)
	}

	decOutput, err := DecryptCard(key, encOutput)
	if err != nil {
		t.Errorf("dec error %s\n", err)
	}

	if !reflect.DeepEqual(c, decOutput) {
		t.Errorf("want %+v, got %+v\n", c, decOutput)
	}
}
