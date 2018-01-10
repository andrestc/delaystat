// +build linux

package netlink_test

import (
	"testing"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nltest"
	"golang.org/x/sys/unix"
)

func TestConnReceiveErrorLinux(t *testing.T) {
	// Note: using *Conn instead of Linux-only *conn, to test
	// error handling logic in *Conn.Receive

	tests := []struct {
		name string
		msgs []netlink.Message
		err  error
	}{
		{
			name: "ENOENT",
			msgs: []netlink.Message{{
				Header: netlink.Header{
					Length:   20,
					Type:     netlink.HeaderTypeError,
					Sequence: 1,
					PID:      1,
				},
				// -2, little endian (ENOENT)
				Data: []byte{0xfe, 0xff, 0xff, 0xff},
			}},
			err: unix.ENOENT,
		},
		{
			name: "multipart done without error attached",
			msgs: []netlink.Message{
				{
					Header: netlink.Header{
						Flags: netlink.HeaderFlagsMulti,
					},
				},
				{
					Header: netlink.Header{
						Type:  netlink.HeaderTypeDone,
						Flags: netlink.HeaderFlagsMulti,
					},
				},
			},
		},
		{
			name: "multipart done with error attached",
			msgs: []netlink.Message{
				{
					Header: netlink.Header{
						Flags: netlink.HeaderFlagsMulti,
					},
				},
				{
					Header: netlink.Header{
						Type:  netlink.HeaderTypeDone,
						Flags: netlink.HeaderFlagsMulti,
					},
					Data: []byte{0xfc, 0xff, 0xff, 0xff},
				},
			},
			// -4, little endian (EINTR)
			err: unix.EINTR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := nltest.Dial(func(_ netlink.Message) ([]netlink.Message, error) {
				return tt.msgs, nil
			})
			defer c.Close()

			// Need to prepopulate nltest's internal buffers by invoking the
			// function once.
			_, _ = c.Send(netlink.Message{})

			_, err := c.Receive()

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v",
					want, got)
			}
		})
	}
}
