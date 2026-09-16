// Package metrics Prometheus 指标：HTTP 中间件埋点 + 关键业务事件计数。
// /metrics 由 handler 注册 promhttp 暴露，供 Prometheus 抓取、Grafana 展示与告警规则消费。
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "pet_api_http_requests_total", Help: "HTTP 请求总数"},
		[]string{"method", "path", "code"},
	)
	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pet_api_http_request_seconds",
			Help:    "HTTP 请求耗时分布",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)
	OrdersCreated = prometheus.NewCounter(prometheus.CounterOpts{Name: "pet_api_orders_created_total", Help: "创建订单数"})
	OrdersPaid    = prometheus.NewCounter(prometheus.CounterOpts{Name: "pet_api_orders_paid_total", Help: "支付成功订单数（含补尾款）"})
	OrdersClosed  = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "pet_api_orders_closed_total", Help: "自动关单数"},
		[]string{"kind"}, // pending=待支付超时 tail=尾款超时
	)
	PayNotifies = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "pet_api_pay_notify_total", Help: "支付回调结果"},
		[]string{"result"}, // ok / amount_mismatch / unknown_order
	)
	NotifyDeliveryFails = prometheus.NewCounter(prometheus.CounterOpts{Name: "pet_api_notify_delivery_fail_total", Help: "通知投递最终失败数"})
	LimitRejected       = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "pet_api_ratelimit_rejected_total", Help: "限流拦截数"},
		[]string{"scope"},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequests, HTTPDuration, OrdersCreated, OrdersPaid,
		OrdersClosed, PayNotifies, NotifyDeliveryFails, LimitRejected)
}

// Handler Prometheus 抓取端点
func Handler() http.Handler { return promhttp.Handler() }

// Middleware HTTP 指标埋点（path 含资源 ID，内部看板可接受该基数）
func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		path := r.URL.Path
		HTTPRequests.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		HTTPDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
