package apinotify

import "fmt"

var ErrManagerClosed = fmt.Errorf("notify manager closed")

type SendError struct {
	Channel string
	Err     error
}

func (e SendError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: notify send failed", e.Channel)
	}
	return fmt.Sprintf("%s: notify send failed: %v", e.Channel, e.Err)
}

func (e SendError) Unwrap() error {
	return e.Err
}
