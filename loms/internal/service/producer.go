package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/model"
)

// ProduceFromOutbox ...
func (s *Service) ProduceFromOutbox(ctx context.Context) {
	messages, err := s.repository.GetNewMsgOutbox(ctx)
	if err != nil {
		logger.Errorw(fmt.Sprintf("GetNewMsgOutbox: %v", err))
	}

	for _, msg := range messages {
		event := &model.OrderEvent{}

		if err = json.Unmarshal(msg.Payload, event); err != nil {
			logger.Errorw(fmt.Sprintf("Unmarshal : %v", err))
			continue
		}

		_, _, err = s.producer.SendMsg(event)
		if err != nil {
			logger.Errorw(fmt.Sprintf("SendMsg id=%d: %v", msg.ID, err))
			continue
		}

		if err = s.repository.UpdateStatusMsgOutbox(ctx, msg.ID, model.StatusMsgSent); err != nil {
			logger.Errorw(fmt.Sprintf("UpdateStatusMsgOutbox : %d, err : %v", msg.ID, err))
		}
	}

}
