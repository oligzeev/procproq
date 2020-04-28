package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/opentracing/opentracing-go"
	"net/http"
)

func BalanceMiddleware(requestUrl string, retryMax int) gin.HandlerFunc {
	tracer := opentracing.GlobalTracer()
	client := retryablehttp.NewClient()
	client.RetryMax = retryMax
	return func(c *gin.Context) {
		// Get current context and tracing
		span, spanCtx := opentracing.StartSpanFromContext(c.Request.Context(), "balanced request: "+c.Request.RequestURI)
		defer span.Finish()

		// Prepare http (with retries) request
		request, err := retryablehttp.NewRequest(c.Request.Method, requestUrl+c.Request.RequestURI, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewError(err))
			return
		}
		request.WithContext(spanCtx)

		// Propagate content type
		// TBD Headers have to be propagated as is
		request.Header.Set(domain.HeaderContentType, c.Request.Header.Get(domain.HeaderContentType))

		// Propagate tracing
		if err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header)); err != nil {
			c.JSON(http.StatusInternalServerError, NewError(err))
			return
		}

		// Perform request
		response, err := client.Do(request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewError(err))
			return
		}

		// Propagate response
		// Example at https://github.com/gin-gonic/gin#serving-data-from-reader
		c.DataFromReader(response.StatusCode, response.ContentLength, response.Header.Get(domain.HeaderContentType),
			response.Body, nil)
	}
}
