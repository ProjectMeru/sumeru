package web

import (
	"context"
	"fmt"
	"strconv"

	"sumeru/core/orm"
)

// PublishSwcBusEvent persists and fan-outs a bus event to subscribers (all processes via NOTIFY).
func PublishSwcBusEvent(ctx context.Context, channel string, payload map[string]interface{}) error {
	if channel == "" {
		return fmt.Errorf("bus channel required")
	}
	eventID, err := orm.InsertBusEvent(ctx, channel, payload)
	if err != nil {
		return err
	}
	if 0 >= eventID {
		return nil
	}
	if err := orm.NotifyBusEvent(eventID); err != nil {
		return err
	}
	dispatchLocalBusEvent(eventID, channel, payload)
	return nil
}

// PublishRecordBusEvent emits record.created|updated|deleted to record and model channels.
func PublishRecordBusEvent(ctx context.Context, eventName, model string, id int) {
	if model == "" || 0 >= id {
		return
	}
	payload := map[string]interface{}{
		"event": eventName,
		"model": model,
		"id":    id,
	}
	ch := RecordBusChannel(model, id)
	_ = PublishSwcBusEvent(ctx, ch, payload)
	modelCh := fmt.Sprintf("model/%s", model)
	_ = PublishSwcBusEvent(ctx, modelCh, payload)
}

func dispatchLocalBusEvent(eventID int64, channel string, payload map[string]interface{}) {
	hub := ensureBusHub()
	frame, err := marshalBusEventFrame(eventID, channel, payload)
	if err != nil {
		return
	}
	hub.publishChannel(channel, frame)
}

func dispatchBusEventByID(ctx context.Context, eventID int64) {
	if 0 >= eventID {
		return
	}
	channel, payload, err := orm.LoadBusEventByID(ctx, eventID)
	if err != nil || channel == "" {
		return
	}
	dispatchLocalBusEvent(eventID, channel, payload)
}

func marshalBusEventFrame(eventID int64, channel string, payload map[string]interface{}) ([]byte, error) {
	return marshalBusFrame(map[string]interface{}{
		"type":    "event",
		"id":      eventID,
		"channel": channel,
		"payload": payload,
	})
}

func parseBusEventIDNotify(payload string) int64 {
	id, _ := strconv.ParseInt(payload, 10, 64)
	return id
}
