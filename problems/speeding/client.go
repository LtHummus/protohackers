package speeding

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Client struct {
	system *System
	conn   net.Conn

	camera     *Camera
	dispatcher *Dispatcher

	ticker *time.Ticker

	writeLock *sync.Mutex
}

func NewClient(system *System, conn net.Conn) *Client {
	return &Client{
		system:    system,
		conn:      conn,
		writeLock: &sync.Mutex{},
	}
}

func (c *Client) Run() {
	msgKind := make([]byte, 1)
	defer c.conn.Close()
	
	for {
		var err error
		_, err = c.conn.Read(msgKind)
		if err != nil {
			log.Error().Err(err).Msg("could not read data")
			break
		}

		switch msgKind[0] {
		case MessageKindPlate:
			log.Info().Msg("got plate message")
			err = c.handlePlateMessage()
		case MessageKindWantHeartbeat:
			log.Info().Msg("got heartbeat want")
			err = c.startHeartbeat()
		case MessageKindIAmCamera:
			log.Info().Msg("got IAmCamera")
			err = c.upgradeToCamera()
		case MessageKindIAmDispatch:
			log.Info().Msg("got IAMDispatch")
			err = c.upgradeToDispatch()
		default:
			log.Warn().Hex("message_kind", msgKind).Msg("got invalid message kind")
			c.Write(InvalidMsgPayload)
			err = errors.New("invalid message kind")
		}

		if err != nil {
			log.Error().Err(err).Msg("error in client cycle")

			if err == ErrNotACamera || err == ErrAlreadyAssigned {
				c.Write(InvalidMsgPayload)
			}
			break
		}
	}
}

func (c *Client) upgradeToCamera() error {
	if c.camera != nil || c.dispatcher != nil {
		log.Warn().Msg("tried to assign self again")
		c.Write(InvalidMsgPayload)
		return errors.New("already assigned role")
	}
	var camera Camera
	err := binary.Read(c.conn, binary.BigEndian, &camera)
	if err != nil {
		log.Error().Err(err).Msg("could not read camera info")
		c.Write(InvalidMsgPayload)
		return errors.New("invalid payload")
	}
	log.Info().Uint16("road", camera.Road).Uint16("mile", camera.Mile).Uint16("limit", camera.Limit).Msg("got camera")
	c.camera = &camera

	c.system.RegisterRoad(camera.Road, camera.Limit)
	return nil
}

func (c *Client) upgradeToDispatch() error {
	log.Info().Msg("attempting to upgrade to dispatcher")
	if c.camera != nil || c.dispatcher != nil {
		log.Warn().Msg("tried to assign self again")
		c.Write(InvalidMsgPayload)
		return errors.New("already assigned role")
	}

	numRoadsPayload := make([]byte, 1)
	_, err := c.conn.Read(numRoadsPayload)
	if err != nil {
		log.Error().Err(err).Msg("could not read from client")
		c.Write(InvalidMsgPayload)
		return errors.New("invalid dispatch payload")
	}

	roads := make([]uint16, numRoadsPayload[0])
	err = binary.Read(c.conn, binary.BigEndian, roads)
	if err != nil {
		log.Error().Err(err).Msg("could not decode roads payload")
		c.Write(InvalidMsgPayload)
		return errors.New("invalid roads payload")
	}

	log.Info().Uints16("roads", roads).Msg("read dispatcher roads")
	c.dispatcher = &Dispatcher{
		Roads:         roads,
		TicketChannel: make(chan []byte),
	}

	for _, curr := range roads {
		c.system.RegisterRoad(curr, 0) // 0 limit so we don't blow another limit away
	}

	c.dispatchChannelCreator()
	go c.dispatchSendChannel()
	return nil
}

func (c *Client) startHeartbeat() error {
	if c.ticker != nil {
		log.Warn().Msg("already has a ticker")
		c.Write(InvalidMsgPayload)
		return errors.New("already has ticker")
	}

	var interval uint32
	err := binary.Read(c.conn, binary.BigEndian, &interval)
	if err != nil {
		log.Error().Err(err).Msg("could not heartbeat time")
		c.Write(InvalidMsgPayload)
		return errors.New("invalid heartbeat payload")
	}

	log.Info().Uint32("interval", interval).Msg("read interval")

	if interval == 0 {
		log.Warn().Msg("interval of 0 recieved, ignoring")
		return nil
	}

	c.ticker = time.NewTicker(time.Duration(interval) * 100 * time.Millisecond)

	go func() {
		for {
			<-c.ticker.C
			_, err := c.Write(HeartbeatPayload)
			if err != nil {
				log.Error().Err(err).Msg("error writing heartbeat")
				c.ticker.Stop()
			}
		}
	}()

	return nil
}

func (c *Client) Write(payload []byte) (int, error) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	return c.conn.Write(payload)
}

func (c *Client) handlePlateMessage() error {
	if c.camera == nil || c.dispatcher != nil {
		log.Warn().Msg("plate message from non-camera")
		return ErrNotACamera
	}

	plate, err := ReadString(c.conn)
	if err != nil {
		log.Warn().Msg("could not read string")
		return ErrCouldNotRead
	}

	var timestamp uint32
	err = binary.Read(c.conn, binary.BigEndian, &timestamp)
	if err != nil {
		log.Warn().Msg("could not read timestamp")
		return ErrCouldNotRead
	}

	log.Info().Str("plate", plate).Uint32("timestamp", timestamp).Msg("read plate")

	c.system.RecordObservation(plate, c.camera.Road, c.camera.Mile, timestamp)

	return nil
}
