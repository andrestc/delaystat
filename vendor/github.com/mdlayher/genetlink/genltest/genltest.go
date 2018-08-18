// Package genltest provides utilities for generic netlink testing.
package genltest

import (
	"fmt"

	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nltest"
)

// Error returns a netlink error to the caller with the specified error
// number.
func Error(number int) error {
	return &errnoError{number: number}
}

type errnoError struct {
	number int
}

func (err *errnoError) Error() string {
	return fmt.Sprintf("genltest errno: %d", err.number)
}

// A Func is a function that can be used to test genetlink.Conn interactions.
// The function can choose to return zero or more generic netlink messages,
// or an error if needed.
//
// For a netlink request/response interaction, the requests greq and nreq are
// populated by genetlink.Conn.Send and passed to the function.  greq is created
// from the body of nreq.
//
// For multicast interactions, both greq and nreq are empty when passed to the function
// when genetlink.Conn.Receive is called.
//
// If a Func returns an error, the error will be returned as-is to the caller.
// If no messages and io.EOF are returned, no messages and no error will be
// returned to the caller, simulating a multi-part message with no data.
type Func func(greq genetlink.Message, nreq netlink.Message) ([]genetlink.Message, error)

// Dial sets up a genetlink.Conn for testing using the specified Func. All requests
// sent from the connection will be passed to the Func.  The connection should be
// closed as usual when it is no longer needed.
func Dial(fn Func) *genetlink.Conn {
	return genetlink.NewConn(nltest.Dial(adapt(fn)))
}

// ServeFamily returns a Func that intercepts "get family" commands to the
// generic netlink controller, verifies that the requested family name matches
// the provided one, and then returns family information specified by f.
//
// Requests which are not related to requesting a family are passed through to fn.
//
// ServeFamily is primarily useful in tests for packages which interact with
// a specific generic netlink family.
func ServeFamily(f genetlink.Family, fn Func) Func {
	return serveFamily(f, fn)
}

// CheckRequest returns a Func that verifies that an incoming request message
// has the specified generic netlink family, command, and netlink header flags,
// and then passes the request through to fn.
//
// If family, command, or flags are set to the zero value, the specific check
// for that value will be skipped for request message.
func CheckRequest(family uint16, command uint8, flags netlink.HeaderFlags, fn Func) Func {
	base := nltest.CheckRequest(
		// Expect genetlink family in header type.
		[]netlink.HeaderType{netlink.HeaderType(family)},
		// Expect specified netlink flags.
		[]netlink.HeaderFlags{flags},
		// Make the next nltest function a noop.
		// TODO(mdlayher): modify nltest to eliminate the need for this?
		nltest.Func(func(_ []netlink.Message) ([]netlink.Message, error) {
			return nil, nil
		}),
	)

	return func(greq genetlink.Message, nreq netlink.Message) ([]genetlink.Message, error) {
		if _, err := base([]netlink.Message{nreq}); err != nil {
			return nil, fmt.Errorf("genltest: netlink header validation failed: %v", err)
		}

		if want, got := command, greq.Header.Command; command != 0 && want != got {
			return nil, fmt.Errorf("genltest: unexpected generic netlink header command: %d, want: %d", got, want)
		}

		return fn(greq, nreq)
	}
}

var _ nltest.Func = adapt(nil)

// adapt is an adapter function for a Func to be used as a nltest.Func.  adapt
// handles marshaling and unmarshaling of generic netlink messages.
func adapt(fn Func) nltest.Func {
	return func(reqs []netlink.Message) ([]netlink.Message, error) {
		var req netlink.Message
		l := len(reqs)
		switch l {
		case 0:
			// No messages.
		case 1:
			// Use the first message.
			req = reqs[0]
		default:
			// Multiple messages; doesn't seem to occur with genetlink?
			return nil, fmt.Errorf("genltest: expected zero or one request, but got: %d", l)
		}

		var gm genetlink.Message

		// Populate message if some data has been passed in req.
		if len(req.Data) > 0 {
			if err := gm.UnmarshalBinary(req.Data); err != nil {
				return nil, err
			}
		}

		gmsgs, err := fn(gm, req)
		if err != nil {
			// An error was returned with an error number by the Func.
			// Pass this to the caller as a netlink message error.
			nerr, ok := err.(*errnoError)
			if !ok {
				return nil, err
			}

			return nltest.Error(nerr.number, reqs)
		}

		nmsgs := make([]netlink.Message, 0, len(gmsgs))
		for _, msg := range gmsgs {
			b, err := msg.MarshalBinary()
			if err != nil {
				return nil, err
			}

			nmsgs = append(nmsgs, netlink.Message{
				// Mimic the sequence and PID of the request for validation.
				Header: netlink.Header{
					Sequence: req.Header.Sequence,
					PID:      req.Header.PID,
				},
				Data: b,
			})
		}

		return nmsgs, nil
	}
}
