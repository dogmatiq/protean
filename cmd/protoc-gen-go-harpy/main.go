package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dogmatiq/harpy/internal/generator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	gen := &generator.Generator{}

	if err := generate(gen, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generate reads a code generation request from r, invokes gen(r) and writes
// its responses to w.
func generate(
	gen *generator.Generator,
	r io.Reader,
	w io.Writer,
) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read request: %w", err)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return fmt.Errorf("unable to unmarshal request: %w", err)
	}

	if len(req.FileToGenerate) == 0 {
		return fmt.Errorf("no files to generate")
	}

	res, err := gen.Generate(req)
	if err != nil {
		return fmt.Errorf("unable to generate response: %w", err)
	}

	data, err = proto.Marshal(res)
	if err != nil {
		return fmt.Errorf("unable to marshal response: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("unable to write response: %w", err)
	}

	return nil
}
