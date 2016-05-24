/*@ngInject*/
module.exports = function() {
  var exports = {};
  exports.buildRepr = function(rule) {
    var parts = [];

    var trendUp = rule.trendUp || false;
    var trendDown = rule.trendDown || false;
    var thresholdMax = rule.thresholdMax || 0;
    var thresholdMin = rule.thresholdMin || 0;

    if (trendUp && thresholdMax === 0) {
      parts.push('trend ↑');
    }
    if (trendUp && thresholdMax !== 0) {
      parts.push('(trend ↑ && value >= ' + parseFloat(thresholdMax.toFixed(3)) +
                 ')');
    }
    if (!trendUp && thresholdMax !== 0) {
      parts.push('value >= ' + parseFloat(thresholdMax.toFixed(3)));
    }
    if (trendDown && thresholdMin === 0) {
      parts.push('trend ↓');
    }
    if (trendDown && thresholdMin !== 0) {
      parts.push('(trend ↓ && value <= ' + parseFloat(thresholdMin.toFixed(3)) +
                 ')');
    }
    if (!trendDown && thresholdMin !== 0) {
      parts.push('value <= ' + parseFloat(thresholdMin.toFixed(3)));
    }
    return parts.join(' || ');
  };

  exports.startsWith = function(s, p) { return s.indexOf(p) === 0; };

  exports.endsWith =
      function(s, p) { return s.indexOf(p, s.length - p.length) !== -1; };

  exports.isGraphiteName =
      function(name) { return exports.startsWith(name, 'stats.'); };

  exports.translateGraphiteName = function(name) {
    var slug;
    if (exports.startsWith(name, 'stats.timers.')) {
      var arr = name.split('.');
      slug = arr.slice(2, arr.length - 1).join('.');
      return 'timer.' + arr[arr.length - 1] + '.' + slug;
    }
    if (exports.startsWith(name, 'stats.gauges.')) {
      slug = name.slice('stats.gauges.'.length);
      return 'gauge.' + slug;
    }
    if (exports.startsWith(name, 'stats.')) {
      return 'counter.' + name.slice('stats.'.length);
    }
  };

  exports.ruleCheck = function(rule) {
    if (exports.isGraphiteName(rule.pattern) && rule.numMetrics === 0) {
      return 1;  // Graphite name.
    }
    if (!exports.startsWith(rule.pattern, 'timer.count_ps.') &&
        !exports.startsWith(rule.pattern, 'timer.mean.') &&
        !exports.startsWith(rule.pattern, 'timer.mean_90.') &&
        !exports.startsWith(rule.pattern, 'timer.mean_95.') &&
        !exports.startsWith(rule.pattern, 'timer.upper.') &&
        !exports.startsWith(rule.pattern, 'timer.upper_90.') &&
        !exports.startsWith(rule.pattern, 'timer.upper_95.') &&
        !exports.startsWith(rule.pattern, 'timer.count.') &&
        !exports.startsWith(rule.pattern, 'timer.count_90.') &&
        !exports.startsWith(rule.pattern, 'timer.count_95.') &&
        !exports.startsWith(rule.pattern, 'timer.median.') &&
        !exports.startsWith(rule.pattern, 'timer.std.') &&
        !exports.startsWith(rule.pattern, 'timer.sum.') &&
        !exports.startsWith(rule.pattern, 'timer.sum_90.') &&
        !exports.startsWith(rule.pattern, 'timer.sum_95.') &&
        !exports.startsWith(rule.pattern, 'timer.sum_suqares.') &&
        !exports.startsWith(rule.pattern, 'timer.lower.') &&
        !exports.startsWith(rule.pattern, 'counter.') &&
        !exports.startsWith(rule.pattern, 'gauge.')) {
      return 2;  // Unsupported metric.
    }
    return 0;  // OK
  };

  // Translate rule repr to readable string.
  exports.translateRuleRepr = function(rule, config, $translate) {
    var parts = [];

    var trendUp = rule.trendUp || false;
    var trendDown = rule.trendDown || false;
    var thresholdMax = rule.thresholdMax || 0;
    var thresholdMin = rule.thresholdMin || 0;

    if (trendUp && thresholdMax === 0) {  // trendUp
      parts.push($translate.instant('ADMIN_RULE_TRANS_TRENDUP'));
    }
    if (trendUp && thresholdMax !== 0) {  // trendUp && value >= thresholdMax
      parts.push($translate.instant('ADMIN_RULE_TRANS_TRENDUP_AND_THRESHOLDMAX',
                                    {'thresholdMax': thresholdMax}));
    }
    if (!trendUp && thresholdMax !== 0) {  // value >= thresholdMax
      parts.push($translate.instant('ADMIN_RULE_TRANS_TRESHOLDMAX',
                                    {'thresholdMax': thresholdMax}));
    }
    if (trendDown && thresholdMin === 0) {  // trendDown
      parts.push($translate.instant('ADMIN_RULE_TRANS_TRENDDOWN'));
    }
    if (trendDown &&
        thresholdMin !== 0) {  // trendDown && value <= thresholdMin
      parts.push(
          $translate.instant('ADMIN_RULE_TRANS_TRENDDOWN_AND_THRESHOLDMIN',
                             {'thresholdMin': thresholdMin}));
    }
    if (!trendDown && thresholdMin !== 0) {  // value <= thresholdMin
      parts.push($translate.instant('ADMIN_RULE_TRANS_TRESHOLDMIN',
                                    {'thresholdMin': thresholdMin}));
    }

    var s = parts.join($translate.instant('ADMIN_RULE_TRANS_OR'));
    return $translate.instant('ADMIN_RULE_TRANS_TPL', {'text': s});
  };

  // Translate Date object to human readable string.
  exports.translateDate = function(date) {
    return date.toDateString().slice(4) + ' ' + date.toTimeString().slice(0, 8);
  };

  exports.translateNow =
      function() { return exports.translateDate(new Date()); };

  exports.translateGoDate = function(s) {
    if (typeof s === 'undefined' || s.length === 0) {
      return exports.translateNow();
    }
    return exports.translateDate(new Date(s));
  };

  // Convert string format timespan to seconds.
  exports.timeSpanToSeconds = function(timeSpan) {
    var map = {
      's': 1,
      'm': 60,
      'h': 60 * 60,
      'd': 24 * 60 * 60,
    };
    var secs = 0, i, ch, measure, count;

    while (timeSpan.length > 0) {
      for (i = 0; i < timeSpan.length; i++) {
        ch = timeSpan[i];
        measure = map[ch];
        if (!isNaN(measure)) {
          count = +timeSpan.slice(0, i);
          secs += count * measure;
          timeSpan = timeSpan.slice(i + 1);
          break;
        }
      }
      if (i === timeSpan.length) return secs;
    }
    return secs;
  };

  // Covert seconds to string format.
  exports.secondsToTimespanString = function(seconds) {
    var numDays = parseInt(seconds / (24 * 60 * 60));
    seconds -= numDays * 24 * 60 * 60;
    var numHours = parseInt(seconds / (60 * 60));
    seconds -= numHours * 60 * 60;
    var numMinutes = parseInt(seconds / 60);
    seconds -= numMinutes * 60;
    var numSeconds = seconds;
    var s = '';
    if (numDays > 0) s += exports.format("%dd", numDays);
    if (numHours > 0) s += exports.format("%dh", numHours);
    if (numMinutes > 0) s += exports.format("%dm", numMinutes);
    if (numSeconds > 0) s += exports.format("%ds", numSeconds);
    return s
  };

  // Convert date to readable string.
  exports.dateToString = function(date) {
    if (!(date instanceof Date)) date = new Date(date);
    var year = date.getFullYear();
    var month = date.getMonth() + 1;
    var day = date.getDate();
    var hours = date.getHours();
    var minutes = date.getMinutes();
    var seconds = date.getSeconds();
    // normalize
    year = '' + year;
    month = ('00' + month).slice(-2);
    day = ('00' + day).slice(-2);
    hours = ('00' + hours).slice(-2);
    minutes = ('00' + minutes).slice(-2);
    seconds = ('00' + seconds).slice(-2);
    return [year, month, day].join('/') +
           ' ' + [hours, minutes, seconds].join(':');
  };

  // Format string.
  exports.format = function() {
    var args = arguments;
    var fmt = args[0];
    var i = 1;
    return fmt.replace(/%((%)|s|d|f)/g, function(m) {
      if (m[2]) {
        return m[2];
      }
      var val = args[i++];
      switch (m) {
        case '%d':
          return parseInt(val);
        case '%f':
          return parseFloat(val);
        case '%s':
          return val.toString();
      }
    });
  };

  return exports;
};
