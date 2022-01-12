package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
)

func ReadFile(path string) ([]byte, error) {
	// pass in aws credentials by cli flag
	// from cli:  -cloud=<"filepath">
	// go run main.go -cloud="/Users/emilymcmullan/.aws/credentials"
	// cloud := flag.String("cloud", "", "file path for aws credentials")
	// flag.Parse()
	// save passed in cred file as []byteq
	file, err := ioutil.ReadFile(path)
	return file, err
}

func GenNameUuid(prefix string) string {
	uid, _ := uuid.NewUUID()
	return fmt.Sprintf("%s-%s", prefix, uid.String())
}
