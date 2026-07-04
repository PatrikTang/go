package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

type LaunchOptions struct {
	Difficulty   Difficulty
	GNUGoPath    string
	KataGoPath   string
	KataGoConfig string
	KataGoModel  string
}

// Launch starts an engine subprocess based on opts. Beginner/Easy use
// GNU Go; Intermediate/Strong use KataGo (which requires config + model
// paths). Returns an Engine wrapping the subprocess.
func Launch(ctx context.Context, opts LaunchOptions) (Engine, error) {
	switch opts.Difficulty {
	case Beginner, Easy:
		return launchGNUGo(ctx, opts)
	case Intermediate, Strong:
		return launchKataGo(ctx, opts)
	}
	return nil, fmt.Errorf("unknown difficulty: %v", opts.Difficulty)
}

func launchGNUGo(ctx context.Context, opts LaunchOptions) (Engine, error) {
	path := opts.GNUGoPath
	if path == "" {
		path = "gnugo"
	}
	level := 1
	if opts.Difficulty == Easy {
		level = 5
	}
	return spawn(ctx, path, "--mode", "gtp", "--level", fmt.Sprint(level))
}

func launchKataGo(ctx context.Context, opts LaunchOptions) (Engine, error) {
	path := opts.KataGoPath
	if path == "" {
		path = "katago"
	}
	if opts.KataGoConfig == "" || opts.KataGoModel == "" {
		return nil, errors.New("KataGo requires KataGoConfig and KataGoModel paths")
	}
	args := []string{"gtp", "-config", opts.KataGoConfig, "-model", opts.KataGoModel}
	if opts.Difficulty == Intermediate {
		args = append(args, "-override-config", "maxVisits=100")
	}
	return spawn(ctx, path, args...)
}

func spawn(ctx context.Context, name string, args ...string) (Engine, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	closer := &procCloser{cmd: cmd, stdin: stdin}
	client := NewGTPClient(stdout, stdin, closer)
	return NewGTPEngine(client), nil
}

type procCloser struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func (p *procCloser) Close() error {
	_ = p.stdin.Close()
	return p.cmd.Wait()
}
