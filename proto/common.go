package proto

import (
	"encoding/json"
	"math"
	"strconv"
	"time"
)

// FrameID uniquely identifies a frame.
type FrameID string

// String returns the FrameID as a string.
func (f FrameID) String() string {
	return string(f)
}

// LoaderID uniquely identifies a loader.
type LoaderID string

// String returns the LoaderID as a string.
func (l LoaderID) String() string {
	return string(l)
}

// RequestID uniquely identifies a network request.
type RequestID string

// String returns the RequestID as a string.
func (r RequestID) String() string {
	return string(r)
}

// RemoteObjectID uniquely identifies a remote JavaScript object.
type RemoteObjectID string

// String returns the RemoteObjectID as a string.
func (r RemoteObjectID) String() string {
	return string(r)
}

// BrowserContextID uniquely identifies a browser context.
type BrowserContextID string

// String returns the BrowserContextID as a string.
func (b BrowserContextID) String() string {
	return string(b)
}

// NodeID is a unique DOM node identifier.
type NodeID int64

// Int64 returns the NodeID as an int64.
func (n NodeID) Int64() int64 {
	return int64(n)
}

// UnmarshalJSON handles both string and number representations from Chrome.
func (n *NodeID) UnmarshalJSON(data []byte) error {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}
	v, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*n = NodeID(v)
	return nil
}

// BackendNodeID is a unique backend DOM node identifier used to reference
// a node that may not have been pushed to the front-end.
type BackendNodeID int64

// Int64 returns the BackendNodeID as an int64.
func (b BackendNodeID) Int64() int64 {
	return int64(b)
}

// UnmarshalJSON handles both string and number representations from Chrome.
func (b *BackendNodeID) UnmarshalJSON(data []byte) error {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}
	v, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*b = BackendNodeID(v)
	return nil
}

// TimeSinceEpoch represents a UTC time as seconds since the Unix epoch.
type TimeSinceEpoch float64

// Time converts the TimeSinceEpoch to a time.Time.
func (t TimeSinceEpoch) Time() time.Time {
	sec, frac := math.Modf(float64(t))
	return time.Unix(int64(sec), int64(frac*1e9))
}

// MarshalJSON encodes TimeSinceEpoch as a float64.
func (t TimeSinceEpoch) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(t))
}

// UnmarshalJSON decodes a float64 into TimeSinceEpoch.
func (t *TimeSinceEpoch) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*t = TimeSinceEpoch(v)
	return nil
}

// MonotonicTime represents a monotonic time as seconds since an unspecified epoch.
type MonotonicTime float64

// MarshalJSON encodes MonotonicTime as a float64.
func (m MonotonicTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(m))
}

// UnmarshalJSON decodes a float64 into MonotonicTime.
func (m *MonotonicTime) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = MonotonicTime(v)
	return nil
}
