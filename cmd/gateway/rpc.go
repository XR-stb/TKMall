package main

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"TKMall/common/log"

	"github.com/gin-gonic/gin"
)

func NewRPCWrapper(serviceCtx *ServiceContext) *RPCWrapper {
	return &RPCWrapper{
		serviceCtx: serviceCtx,
	}
}

type RPCWrapper struct {
	serviceCtx *ServiceContext
}

func (w *RPCWrapper) Call(serviceName string, fn interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 打印请求方法和URL
		log.Infof("Received request: %s %s", c.Request.Method, c.Request.URL)

		// 打印请求体
		bodyBytes, err := c.GetRawData()
		if err != nil {
			log.Errorf("Failed to read request body: %v", err)
		} else {
			log.Infof("Request body: %s", string(bodyBytes))
		}

		// 通过反射动态调用方法
		fnValue := reflect.ValueOf(fn)
		if fnValue.Kind() != reflect.Func {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  http.StatusInternalServerError,
				"error": "Invalid handler function",
			})
			return
		}

		// 获取客户端实例
		var client interface{}
		if err := w.serviceCtx.GetClient(serviceName, &client); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  http.StatusInternalServerError,
				"error": fmt.Sprintf("%s service unavailable", serviceName),
			})
			return
		}

		// 获取方法
		methodName := runtime.FuncForPC(fnValue.Pointer()).Name()
		// 从最后一个点号后面取方法名
		if lastDot := strings.LastIndex(methodName, "."); lastDot >= 0 {
			methodName = methodName[lastDot+1:]
		}
		method := reflect.ValueOf(client).MethodByName(methodName)
		if !method.IsValid() {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  http.StatusInternalServerError,
				"error": fmt.Sprintf("Method %s not found", methodName),
			})
			return
		}

		// 构造请求参数
		reqType := method.Type().In(1) // 方法的第二个参数是请求参数
		req := reflect.New(reqType.Elem()).Interface()

		// 解析查询参数到 gRPC 请求对象
		if err := c.ShouldBindQuery(req); err != nil {
			log.Errorf("Failed to bind query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  http.StatusBadRequest,
				"error": err.Error(),
			})
			return
		}

		// 打印解析后的请求参数
		log.Infof("Parsed request: %+v", req)

		// 调用方法
		results := method.Call([]reflect.Value{
			reflect.ValueOf(c.Request.Context()),
			reflect.ValueOf(req),
		})

		// 处理结果
		if !results[1].IsNil() {
			err := results[1].Interface().(error)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  http.StatusInternalServerError,
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, results[0].Interface())
	}
}
