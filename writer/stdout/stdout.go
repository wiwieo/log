package stdout

import (
	"fmt"
	"os"
)

type stdout struct {
}

func New() (*stdout, error) {
	return &stdout{}, nil
}

func (n *stdout) Write(content []byte) error {
	fmt.Fprint(os.Stdout, string(content))
	return nil
}

func (n *stdout) Close() error {
	return nil
}
