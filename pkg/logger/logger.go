package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	//"github.com/spf13/viper"
)

var (
	Log        *logrus.Logger
	AccessLog  *logrus.Logger
	RuntimeLog *logrus.Logger
	MetricsLog *logrus.Logger
)


//var (
//	LogRemain = viper.GetInt("logger.logRemain") // Total days of log files to remain
//	LogDir    = viper.GetString("logger.logDir")
//	ReqLogger = logFieldKey(viper.GetString("logger.reqLogger"))
//)
const (
	LogRemain int = 10
	LogDir    = "/var/log/yager"
	ReqLogger logFieldKey = "rqlog"
)

type logFieldKey string

type MetricsJSONFormatter struct{}

func (f *MetricsJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry.
	entry.Data["time"] = entry.Time.Format(time.RFC3339)
	serialized, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func FromCtx(ctx context.Context) *logrus.Entry {
	fmt.Printf("log from ctx.")
	if log, ok := ctx.Value(ReqLogger).(*logrus.Entry); ok {
		return log
	}
	return logrus.NewEntry(Log)
}

func MetricsEmit(method, reqId string, latency float32, success bool) {
	MetricsLog.WithFields(logrus.Fields{
		"topic":   "trace",
		"method":  method,
		"reqId":   reqId,
		"latency": latency,
		"success": success,
	}).Info()
}
