package vcf

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

type StreamRecord struct {
	Line int
	Raw  string
}

func Stream(ctx context.Context, r io.Reader, maximumLineBytes int) (<-chan StreamRecord, <-chan error) {
	records := make(chan StreamRecord)
	errors := make(chan error, 1)
	go func() {
		defer close(records)
		defer close(errors)
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64<<10), maximumLineBytes)
		line := 0
		for scanner.Scan() {
			line++
			value := scanner.Text()
			if strings.HasPrefix(value, "#") || strings.TrimSpace(value) == "" {
				continue
			}
			select {
			case records <- StreamRecord{Line: line, Raw: value}:
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errors <- fmt.Errorf("stream vcf: %v", err)
		}
	}()
	return records, errors
}
