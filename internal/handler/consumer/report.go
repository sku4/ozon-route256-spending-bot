package consumer

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/kafka"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
)

func (h *Handler) ReportMessage(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
	var report kafka.Report
	err = json.Unmarshal(msg.Value, &report)
	if err != nil {
		logger.Infos("unmarshal error:", err.Error())
		return
	}

	err = h.services.BuildReport.Build(ctx, report.F1, report.F2, report.UserCurr, report.ChatId)
	if err != nil {
		logger.Infos("report message error:", err.Error())
	}

	return
}
