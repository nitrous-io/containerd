package dockerlxc

import (
	"io"
	"os"
	"os/exec"

	"github.com/kr/pty"
	"github.com/nitrous-io/containerd/agent"
	"github.com/nitrous-io/containerd/pkg/term"
)

type shell struct {
	username string
	id       string
	term     io.ReadWriter
	// we use lxc-attach by default, but user can switch this to a more
	// lightweight command like bash for testing purpose
	cmd *exec.Cmd
}

func (s *shell) Close() {
	p := s.term.(*os.File)
	p.Close()
}

func (s *shell) Open(term string, w, h float64) error {
	if s.cmd == nil {
		args := []string{
			"--name", s.id,
			"--clear-env",
			"--set-var", "TERM=" + term,
			"--",
			"sudo", "login", "-f", s.username,
		}
		s.cmd = exec.Command("lxc-attach", args...)
	}
	sh, err := pty.Start(s.cmd)
	if err != nil {
		return err
	}
	s.term = sh
	return s.Resize(w, h)
}

func (s *shell) Attach(stdin io.Reader, stdout io.Writer) chan error {
	errors := make(chan error, 2)

	// stdin
	go func() {
		_, err := io.Copy(s.term, stdin)
		if err == io.ErrClosedPipe {
			err = nil
		}
		errors <- err
	}()

	// stdout
	go func() {
		_, err := io.Copy(stdout, s.term)
		if err == io.ErrClosedPipe {
			err = nil
		}
		errors <- err
	}()

	return promise(func() error {
		defer s.Close()
		for i := 0; i < 2; i += 1 {
			if err := <-errors; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *shell) Resize(w, h float64) error {
	return term.SetWinsize(s.fd(), &term.Winsize{
		Width:  uint16(w),
		Height: uint16(h),
	})
}

func (s *shell) fd() uintptr {
	p := s.term.(*os.File)
	return p.Fd()
}

func newShell(id, username string) (agent.Shell, error) {
	return &shell{
		id:       id,
		username: username,
	}, nil
}

// Go is a basic promise implementation: it wraps calls a function in a goroutine,
// and returns a channel which will later return the function's return value.
func promise(f func() error) chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- f()
	}()
	return ch
}
