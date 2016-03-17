/*@ngInject*/
module.exports = function () {
  var exports = {};
  exports.buildRepr = function (rule) {
    var parts = [];

    var trendUp = rule.trendUp || false;
    var trendDown = rule.trendDown || false;
    var thresholdMax = rule.thresholdMax || 0;
    var thresholdMin = rule.thresholdMin || 0;

    if (trendUp && thresholdMax === 0) {
      parts.push('trend ↑');
    }
    if (trendUp && thresholdMax !== 0) {
      parts.push('(trend ↑ && value >= ' + parseFloat(thresholdMax.toFixed(3)) + ')');
    }
    if (!trendUp && thresholdMax !== 0) {
      parts.push('value >= ' + parseFloat(thresholdMax.toFixed(3)));
    }
    if (trendDown && thresholdMin === 0) {
      parts.push('trend ↓');
    }
    if (trendDown && thresholdMin !== 0) {
      parts.push('(trend ↓ && value <= ' + parseFloat(thresholdMin.toFixed(3)) + ')');
    }
    if (!trendDown && thresholdMin !== 0) {
      parts.push('value <= ' + parseFloat(thresholdMin.toFixed(3)));
    }
    return parts.join(' || ');
  };

  exports.startsWith = function(s, p) {
    return s.indexOf(p) === 0;
  };

  exports.endsWith = function(s, p) {
    return s.indexOf(p, s.length - p.length) !== -1;
  };

  exports.isGraphiteName = function(name) {
    return exports.startsWith(name, 'stats.');
  };

  exports.translateGraphiteName = function(name) {
    var slug;
    if (exports.startsWith(name, 'stats.timers.')) {
      var arr = name.split('.');
      slug = arr.slice(2, arr.length-1).join('.');
      return 'timer.' + arr[arr.length-1] + '.' + slug;
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
      return false;
    }
    return true;
  };

  return exports;
};
