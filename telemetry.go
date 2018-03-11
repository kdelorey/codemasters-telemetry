package cmtelemetry

import (
	"net"
)

// Telemetry is the basic
type Telemetry struct {
	dataAccessor TelemetryAccessor

	udp    *net.UDPConn
	buffer []byte
	quit   chan struct{}
	rec    chan TelemetryAccessor
}

// TelemetryAccessor is an interface that is responsible for accessing the current
// telemetry values.
type TelemetryAccessor interface {
	GetFieldValue(field TelemetryField) (float32, error)
}

// GatherDefaultTelemetry will open the default udp port and begin processing
// telemetry from Codemasters' games.
func GatherDefaultTelemetry(frames chan TelemetryAccessor) (*Telemetry, error) {
	return GatherTelemetry(":20777", frames)
}

// GatherTelemetry will open a udp port up at the specified location and begin
// processing telemetry from Codemasters' games.
func GatherTelemetry(address string, frames chan TelemetryAccessor) (t *Telemetry, err error) {
	addr, err := net.ResolveUDPAddr("udp", address)

	if err != nil {
		return
	}

	udp, err := net.ListenUDP("udp", addr)

	if err != nil {
		return
	}

	acc, buff := createMode3Accessor()

	t = &Telemetry{
		dataAccessor: acc,
		udp:          udp,
		buffer:       buff,
		quit:         make(chan struct{}),
		rec:          frames,
	}

	go t.telemetryRoutine()

	return
}

// Close will stop the connection.
func (t *Telemetry) Close() {
	close(t.rec)
	close(t.quit)
	t.udp.Close()
}

func (t *Telemetry) telemetryRoutine() {
	for {
		select {
		case <-t.quit:
			return

		default:
			t.udp.ReadFromUDP(t.buffer)
			t.rec <- t.dataAccessor
			break
		}
	}
}
