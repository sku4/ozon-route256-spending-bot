package server

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler/telegram"
)

func TracingMiddleware(next telegram.IHandler) telegram.IHandler {
	return telegram.Func(func(ctx context.Context, upd tgbotapi.Update) (err error) {
		incomingTrace, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(map[string][]string{}),
		)

		span, _ := opentracing.StartSpanFromContext(
			ctx,
			"RequestTelegramBot",
			ext.RPCServerOption(incomingTrace),
		)
		defer span.Finish()

		err = next.IncomingMessage(ctx, upd)
		if err != nil {
			ext.Error.Set(span, true)
		}

		return
	})
}
