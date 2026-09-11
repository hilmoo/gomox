package mcpx

import (
	"log/slog"
	"net/http"

	mlog "github.com/hilmoo/gomox/transport/middleware/log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewStreamableHTTPHandler wraps s in an [mcp.StreamableHTTPHandler] that initializes
// request-scoped logging attributes (via [mlog.InitMcp]) for every incoming request.
func NewStreamableHTTPHandler(s *mcp.Server) *mcp.StreamableHTTPHandler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		mlog.InitMcp(r)
		return s
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})
}

// NewServer creates an [mcp.Server] with structured request logging (via [mlog.Mcp])
// installed as a receiving middleware.
func NewServer(impl *mcp.Implementation, log *slog.Logger) *mcp.Server {
	s := mcp.NewServer(impl, nil)
	s.AddReceivingMiddleware(mlog.Mcp(log))
	return s
}
