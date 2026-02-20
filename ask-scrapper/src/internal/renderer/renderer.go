package renderer

import (
	"context"
	"errors"
	"fmt"

	renderermodel "github.com/omo-ri/askorium/go-renderer-api"
)

type Renderer interface {
	Render(ctx context.Context, url string, timeoutMs int) (html string, elapsedMs int, err error)
}

type httpRenderer struct {
	client *renderermodel.APIClient
}

func NewRenderer(baseURL string) Renderer {
	cfg := renderermodel.NewConfiguration()
	cfg.Servers = renderermodel.ServerConfigurations{
		{
			URL: baseURL,
		},
	}
	client := renderermodel.NewAPIClient(cfg)
	return &httpRenderer{client: client}
}

func (r *httpRenderer) Render(ctx context.Context, url string, timeoutMs int) (html string, elapsedMs int, err error) {
	timeoutMs32 := int32(timeoutMs)
	req := r.client.RendererAPI.RenderPage(ctx).RenderRequest(renderermodel.RenderRequest{
		Url:       url,
		TimeoutMs: &timeoutMs32,
	})
	resp, _, err := req.Execute()
	if err != nil {
		var apiErr *renderermodel.GenericOpenAPIError
		if errors.As(err, &apiErr) {
			if errResp, ok := apiErr.Model().(renderermodel.RenderErrorResponse); ok {
				return "", 0, &RenderError{
					Code:    string(errResp.GetCode()), // RenderErrorCode → string
					Message: errResp.GetMessage(),
				}
			}
		}
		return "", 0, &RenderError{
			Code:    "NETWORK_ERROR",
			Message: err.Error(),
		}
	}
	return resp.GetHtml(), int(resp.GetElapsedMs()), nil
}

type RenderError struct {
	Code    string
	Message string
}

func (e *RenderError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
