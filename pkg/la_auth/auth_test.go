package la_auth

import (
	"fmt"
	"testing"
)

func TestBuildClaims(t *testing.T) {
	InitSecret("lalala")
	token, err := NewToken("lalala", 0, 30)
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("token: %s\n", token)
}
