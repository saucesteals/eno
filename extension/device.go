package extension

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Device struct {
	AuthTransactionId string
	ExtensionId       string
}

func GenerateExtensionId() string {
	h := func() string {
		b := make([]byte, 2)
		rand.Read(b)
		return hex.EncodeToString(b)
	}

	return fmt.Sprintf("%d-%s%s-%s-%s-%s-%s%s%s",
		time.Now().UnixMilli(),
		h(), h(),
		h(),
		h(),
		h(),
		h(), h(), h(),
	)
}

func GenerateAuthTransactionId() string {
	return "GO_" + uuid.NewString()
}

func GenerateDevice() Device {
	return Device{
		ExtensionId:       GenerateExtensionId(),
		AuthTransactionId: GenerateAuthTransactionId(),
	}
}
