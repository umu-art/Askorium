package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"ask-parser/internal/amqp"
	amqpimpl "ask-parser/internal/amqp/impl"
	"ask-parser/internal/parser"

	renderermodel "github.com/omo-ri/askorium/go-renderer-api"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

func NewHandler(p parser.Parser, publisher amqp.MessagePublisher, logger *slog.Logger) amqpimpl.Handler {
	return func(ctx context.Context, msg []byte) amqpimpl.AckAction {
		var input renderermodel.RenderOutput
		if err := json.Unmarshal(msg, &input); err != nil {
			logger.Error("failed to unmarshal RenderOutput", "error", err)
			return amqpimpl.NackDiscard
		}

		taskID := input.GetTaskId()

		var resp *scrappermodel.ScrapeResponse

		if !input.GetSuccess() {
			resp = scrappermodel.NewScrapeResponse(taskID, false)
			if input.HasError() {
				renderErr := input.GetError()
				scrapeErr := scrappermodel.NewScrapeError(
					scrappermodel.ErrorCode(renderErr.GetCode()),
					renderErr.GetMessage(),
				)
				resp.SetError(*scrapeErr)
			}
		} else {
			page, err := p.Parse(input.GetHtml(), input.GetUrl())
			if err != nil {
				logger.Error("parse failed", "task_id", taskID, "error", err)
				resp = scrappermodel.NewScrapeResponse(taskID, false)
				resp.SetError(*scrappermodel.NewScrapeError(scrappermodel.ERRORCODE_PARSE_ERROR, err.Error()))
			} else {
				logger.Info("parse succeeded", "task_id", taskID, "url", input.GetUrl())
				resp = scrappermodel.NewScrapeResponse(taskID, true)
				resp.SetPage(*page)
			}
		}

		if input.HasMetadata() {
			resp.SetMetadata(input.GetMetadata())
		}

		body, err := json.Marshal(resp)
		if err != nil {
			logger.Error("failed to marshal ScrapeResponse", "task_id", taskID, "error", err)
			return amqpimpl.NackDiscard
		}

		if err := publisher.Publish(ctx, body); err != nil {
			logger.Error("failed to publish response", "task_id", taskID, "error", err)
			return amqpimpl.NackRequeue
		}

		logger.Info("task completed", "task_id", taskID, "success", resp.GetSuccess())
		return amqpimpl.Ack
	}
}
