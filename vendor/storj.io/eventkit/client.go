package eventkit

import (
	"bytes"
	"compress/zlib"
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"storj.io/eventkit/pb"
	"storj.io/eventkit/utils"
	"storj.io/picobuf"
)

const (
	defaultQueueDepth           = 100
	defaultMaxUncompressedBytes = 1000
	defaultCompressionLevel     = zlib.BestCompression
	defaultFlushInterval        = 15 * time.Second
)

// this is the size of a zlib compressed, serialized pb.Packet with SendOffset
// set to a reasonable value.
const trailerSize = 24

type UDPClient struct {
	Application string
	Version     string
	Instance    string
	Addr        string

	QueueDepth           int
	MaxUncompressedBytes int
	CompressionLevel     int
	FlushInterval        time.Duration

	initOnce    sync.Once
	submitQueue chan *Event

	writerPool    *zlib.Writer
	droppedEvents atomic.Int64
}

var _ Destination = &UDPClient{}

func NewUDPClient(application, version, instance, addr string) *UDPClient {
	c := &UDPClient{
		Application: application,
		Version:     version,
		Instance:    instance,
		Addr:        addr,

		QueueDepth:           defaultQueueDepth,
		MaxUncompressedBytes: defaultMaxUncompressedBytes,
		CompressionLevel:     defaultCompressionLevel,
		FlushInterval:        defaultFlushInterval,
	}
	return c
}

func (c *UDPClient) init() {
	c.initOnce.Do(func() {
		c.submitQueue = make(chan *Event, c.QueueDepth)
	})
}

type outgoingPacket struct {
	buf                      bytes.Buffer
	zl                       *zlib.Writer
	written, maxUncompressed int
	events                   int
	startTime                time.Time

	client *UDPClient
}

func (c *UDPClient) newOutgoingPacket() *outgoingPacket {
	op := &outgoingPacket{
		startTime:       time.Now(),
		maxUncompressed: c.MaxUncompressedBytes,
		client:          c,
	}
	op.buf.Grow(c.MaxUncompressedBytes)

	_, err := op.buf.Write([]byte("EK"))
	if err != nil {
		panic(err)
	}

	// grab a zlib writer from pool, when one exists.
	op.zl, c.writerPool = c.writerPool, nil
	if op.zl == nil {
		op.zl, err = zlib.NewWriterLevel(&op.buf, c.CompressionLevel)
		if err != nil {
			panic(err)
		}
	} else {
		op.zl.Reset(&op.buf)
	}

	data, err := picobuf.Marshal(&pb.Packet{
		Application:        c.Application,
		ApplicationVersion: c.Version,
		Instance:           c.Instance,
		StartTimestamp:     pb.AsTimestamp(op.startTime),
	})
	if err != nil {
		panic(err)
	}

	op.written += 2 + len(data)

	_, err = op.zl.Write(data)
	if err != nil {
		panic(err)
	}

	return op
}

func (op *outgoingPacket) finalize() []byte {
	data, err := picobuf.Marshal(&pb.Packet{
		SendOffsetNs: int64(time.Since(op.startTime)),
	})
	if err != nil {
		panic(err)
	}

	_, err = op.zl.Write(data)
	if err != nil {
		panic(err)
	}

	err = op.zl.Close()
	if err != nil {
		panic(err)
	}

	// put zlib writer back to the pool.
	op.client.writerPool, op.zl = op.zl, nil

	return op.buf.Bytes()
}

func (op *outgoingPacket) addEvent(ev *Event) (full bool) {
	var v pb.Event

	v.Name = ev.Name
	v.Scope = ev.Scope
	v.TimestampOffsetNs = int64(ev.Timestamp.Sub(op.startTime))
	v.Tags = ev.Tags

	data, err := picobuf.Marshal(&pb.Packet{Events: []*pb.Event{&v}})
	if err != nil {
		panic(err)
	}

	op.written += len(data)

	_, err = op.zl.Write(data)
	if err != nil {
		panic(err)
	}

	err = op.zl.Flush()
	if err != nil {
		panic(err)
	}

	op.events += 1
	return (op.written + trailerSize) > op.maxUncompressed
}

func (c *UDPClient) Run(ctx context.Context) {
	c.init()

	ticker := utils.NewJitteredTicker(c.FlushInterval)
	var background errgroup.Group
	defer func() { _ = background.Wait() }()
	background.Go(func() error {
		ticker.Run(ctx)
		return nil
	})

	p := c.newOutgoingPacket()

	sendAndReset := func() {
		_ = c.send(p, c.Addr)
		p = c.newOutgoingPacket()
	}

	for {
		if drops := c.droppedEvents.Load(); drops > 0 {
			c.droppedEvents.Add(-drops)
			if p.addEvent(&Event{
				Name:      "drops",
				Scope:     []string{"storj.io/eventkit"},
				Timestamp: time.Now(),
				Tags:      []Tag{Int64("events", drops)},
			}) {
				sendAndReset()
			}
		}

		select {
		case em := <-c.submitQueue:
			if p.addEvent(em) {
				sendAndReset()
			}
		case <-ticker.C:
			if p.events > 0 {
				sendAndReset()
			}
		case <-ctx.Done():
			left := len(c.submitQueue)
			for i := 0; i < left; i++ {
				if p.addEvent(<-c.submitQueue) {
					sendAndReset()
				}
			}
			if p.events > 0 {
				_ = c.send(p, c.Addr)
			}
			return
		}
	}
}

func (c *UDPClient) send(packet *outgoingPacket, addr string) (err error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, laddr)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := conn.Close(); err == nil && errClose != nil {
			err = errClose
		}
	}()

	_, _, err = conn.WriteMsgUDP(packet.finalize(), nil, nil)
	return err
}

func (c *UDPClient) Submit(events ...*Event) {
	c.init()

	for _, event := range events {
		select {
		case c.submitQueue <- event:
			return
		default:
			c.droppedEvents.Add(1)
		}
	}
}
