# logger

Example of using gin-gonic/gin.

```go
import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/logger"
	"github.com/itsubaki/tracer"
	"go.opentelemetry.io/otel"
)

var tr = otel.Tracer("mypackage/myfunc")

func Func(c *gin.Context) {
	traceID := c.GetString("trace_id")
	spanID := c.GetString("span_id")
	traceTrue := c.GetBool("trace_true")

	log := logger.New(c.Request, traceID, spanID)
	parent, err := tracer.Context(c.Request.Context(), traceID, spanID, traceTrue)
	if err != nil {
		log.ErrorReport("new context: %v", err)
		return
	}
	log.Info("new tracer context")

	func() {
		_, s := tr.Start(parent, "something to do")
		defer s.End()

		log.Span(s).Info("something to do")
	}()

	...
}
```
