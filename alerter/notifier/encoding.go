package notifier

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/eleme/banshee/models"
)

const (
	// PrefixCountPS is the prefix of call counter metric per time unit metric name
	PrefixCountPS = "timer.count_ps."
	// PrefixMean90 is the prefix ofprefix of mean response time metric name
	PrefixMean90 = "timer.mean_90."
	// PrefixUpper90 is the prefix of upper 90 response time metric name
	PrefixUpper90 = "timer.upper_90."
	// PrefixCounter is the  prefix of counter metric name
	PrefixCounter = "counter."
	// PrefixGauge is the prefix of gauge metric name
	PrefixGauge = "gauge."
)

var (
	// Magnitudes contains six magnitude unit
	Magnitudes = []string{"", "K", "M", "G", "T", "P"}
	// RuleLevels contains  three rule level
	RuleLevels = []string{"低", "中", "高"}
)

func wrapNumber(num float64) string {
	magnitude := 0
	cnum := num
	for math.Abs(cnum) >= 1000 {
		magnitude++
		cnum /= 1000.0
	}
	if magnitude >= len(Magnitudes) {
		cnum = num
		magnitude = 0
	}
	s := fmt.Sprintf("%.2f", cnum)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s + Magnitudes[magnitude]
}

func graphiteName(name string) string {
	if strings.HasPrefix(name, PrefixCounter) {
		return fmt.Sprintf("stats.%s", strings.TrimPrefix(name, PrefixCounter))
	}
	if strings.HasPrefix(name, PrefixCountPS) {
		return fmt.Sprintf("stats.timers.%s.count_ps", strings.TrimPrefix(name, PrefixCountPS))
	}
	if strings.HasPrefix(name, PrefixMean90) {
		return fmt.Sprintf("stats.timers.%s.mean_90", strings.TrimPrefix(name, PrefixMean90))
	}
	if strings.HasPrefix(name, PrefixUpper90) {
		return fmt.Sprintf("stats.timers.%s.upper_90", strings.TrimPrefix(name, PrefixUpper90))
	}
	if strings.HasPrefix(name, PrefixGauge) {
		return fmt.Sprintf("stats.gauges.%s", strings.TrimPrefix(name, PrefixGauge))
	}
	return name
}
func translateComment(pattern string, metricName string, comment string) string {
	patternParts := strings.Split(pattern, ".")
	metricParts := strings.Split(metricName, ".")
	if len(patternParts) != len(metricParts) {
		return comment
	}

	i := 0
	for idx, patternPart := range patternParts {
		if patternPart == "*" {
			i++
			marker := fmt.Sprintf("$%d", i)
			comment = strings.Replace(comment, marker, metricParts[idx], 1)
		}
	}
	return comment
}

func packMessage(ew *models.EventWrapper) string {
	metricName := ew.Event.Metric.Name
	ruleComment := ew.Rule.Comment
	var (
		unit      string
		tp        string
		trend     string
		msg       string
		threshold float64
	)
	translatedComment := metricName
	if ruleComment != "" {
		translatedComment = translateComment(ew.Rule.Pattern, metricName, ruleComment)
	}
	if strings.HasPrefix(metricName, PrefixCountPS) {
		unit = "次/秒"
		tp = "调用次数"
		metricName = strings.TrimPrefix(metricName, PrefixCountPS)
		metricName = strings.TrimPrefix(metricName, "new.")
	} else if strings.HasPrefix(metricName, PrefixMean90) {
		unit = "毫秒"
		tp = "响应时间(mean_90)"
		metricName = strings.TrimPrefix(metricName, PrefixMean90)
		metricName = strings.TrimPrefix(metricName, "new.")
	} else if strings.HasPrefix(metricName, PrefixUpper90) {
		unit = "毫秒"
		tp = "响应时间(upper_90)"
		metricName = strings.TrimPrefix(metricName, PrefixUpper90)
		metricName = strings.TrimPrefix(metricName, "new.")
	} else if strings.HasPrefix(metricName, PrefixCounter) {
		unit = "个/秒"
		metricName = strings.TrimPrefix(metricName, PrefixCounter)
		metricName = strings.TrimPrefix(metricName, "new.")
		if strings.HasSuffix(metricName, ".sick") {
			tp = "熔断错误个数"
		} else if strings.HasSuffix(metricName, ".crit") {
			tp = "严重错误个数"
		} else if strings.HasPrefix(metricName, ".timeout") {
			tp = "超时错误个数"
		} else if strings.HasSuffix(metricName, ".soft_timeout") {
			tp = "软超时个数"
		} else if strings.HasSuffix(metricName, ".unkwn_exc") {
			tp = "未知错误个数"
		} else {
			tp = "个数"
		}
	} else if strings.HasSuffix(metricName, PrefixGauge) {
		unit = ""
		tp = "gauge指标"
		metricName = strings.TrimPrefix(metricName, PrefixGauge)
		metricName = strings.TrimPrefix(metricName, "new.")
	} else {
		unit = ""
		tp = ""
	}
	rule := ew.Event.Rule
	metric := ew.Event.Metric
	project := ew.Project

	ruleID := rule.ID
	eventID := ew.ID
	level := RuleLevels[rule.Level]
	date := time.Unix(int64(metric.Stamp), 0).Format("15:04:05")
	avg := metric.Average
	value := metric.Value
	if rule.TrackIdle && ew.Metric.Value == 0 && ew.Metric.Average == 0 && ew.Metric.Score == 0 {
		msg = fmt.Sprintf("{%s等级 %s %s %d %s} %s 丢失数据(空值检测)",
			level, date, project.Name, ruleID, eventID[:7], translatedComment)
	} else if rule.TrendUp || rule.TrendDown {
		if ew.Index.Score > 0 {
			trend = "增加"
		} else {
			trend = "减少"
		}
		msg = fmt.Sprintf("{%s等级 %s %s %d %s} %s%s较往日同时段%s 当前值%s%s,历史平均值%s%s",
			level, date, project.Name, ruleID, eventID[:7], translatedComment, tp,
			trend, wrapNumber(value), unit, wrapNumber(avg), unit)
	} else {
		if rule.ThresholdMin != 0 && metric.Value <= rule.ThresholdMin {
			trend = "小于"
			threshold = rule.ThresholdMin
		} else {
			trend = "大于"
			threshold = rule.ThresholdMax
		}
		msg = fmt.Sprintf("{%s等级 %s %s %d %s} %s%s%s设定阈值 当前值%s%s,阈值%s%s",
			level, date, project.Name, ruleID, eventID[:7], translatedComment, tp,
			trend, wrapNumber(value), unit, wrapNumber(threshold), unit)
	}
	return msg
}
func encodeMessage(msg string) string {
	return url.QueryEscape(msg)
}

func getGrafanaPanelURL(metricName string) string {
	return fmt.Sprintf(cfg.Webapp.GraphiteURL, metricName)
}

func getRuleURL(teamID, projID, ruleID int) string {
	urlTpl := "%s/#/admin/team/%d/project/%d?rule=%d"
	return fmt.Sprintf(urlTpl, cfg.Webapp.URLPrefix, teamID, projID, ruleID)
}
