// multiplex multiple streams into a single stream that
// output json-encoded data i.e.
//
// {Type:"stdout",Data: []byte("hello")}
// {Type:"stderr",Data: []byte("error")}
//
// So it's suitable to transport stdout/stderr elsewhere
package stdstream

import (
	"fmt"
	"io"
)

type StdStream struct {
	output  io.Writer
	streams map[string]io.Reader
}

func (c *StdStream) Write(p []byte) (int, error) {
	b := append(p, '\n')
	if _, err := c.output.Write(b); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *StdStream) Flush() error {
	errDone := make(chan error, len(c.streams))
	for name, reader := range c.streams {
		outR, outW := io.Pipe()
		tw := newTypeWriter(name, outW)

		go func(reader io.Reader, tw io.Writer) {
			_, err := io.Copy(tw, reader)
			outW.CloseWithError(err)
		}(reader, tw)

		go func(outR io.Reader) {
			_, err := io.Copy(c, outR)
			errDone <- err
		}(outR)
	}

	errs := []error{}
	for i := 0; i < len(c.streams); i++ {
		err := <-errDone
		if err != nil && err != io.EOF {
			errs = append(errs)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("Got error %v", errs)
	}
	return nil
}

func NewStdStream(output io.Writer, streams map[string]io.Reader) *StdStream {
	return &StdStream{
		output:  output,
		streams: streams,
	}
}
